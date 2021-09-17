package songmem

import (
	"database/sql"
	"errors"
	sqlite3 "github.com/mattn/go-sqlite3"
	"time"
)

type SongDB struct {
	*sql.DB
}

type songHearing struct {
	Name string
	Date time.Time
}

func InitDB(filepath string) (SongDB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err == nil {
		if db == nil {
			err = errors.New("db is nil")
		} else {
			_, err = db.Exec(`PRAGMA foreign_keys = ON`)
		}
	}
	return SongDB{db}, err
}

func (db SongDB) CreateSchemaIfNotExists() (err error) {
	// id is explicitly used instead of rowid, so that AUTOINCREMENT can be set.
	// This ensures, that one can, for example, delete the last added hearing.
	commands := [...]string{
		`CREATE TABLE IF NOT EXISTS song(
		     id      INTEGER PRIMARY KEY AUTOINCREMENT,
		     name    TEXT NOT NULL,
		     addedAt TEXT NOT NULL,
		     CONSTRAINT name_unique UNIQUE(name COLLATE NOCASE)
		 )`,
		`CREATE INDEX IF NOT EXISTS song_name ON song(name)`,
		`CREATE TABLE IF NOT EXISTS hearing(
		     id      INTEGER PRIMARY KEY AUTOINCREMENT,
		     songID  INTEGER NOT NULL,
		     heardAt TEXT NOT NULL,
		     FOREIGN KEY(songID) REFERENCES song(id)
		 )`}

	tx, err := db.Begin()
	defer tx.Rollback()
	if err != nil {
		return
	}
	for _, c := range commands {
		if _, err = tx.Exec(c); err != nil {
			return
		}
	}
	return tx.Commit()
}

// AddSong adds the song with the given name and the current timestamp
// to the database.
//
// Feel free to include the artist's name in the song.
//
// The current timestamp will be stored with the local timezone, to
// enable sorting songs by time of day, even when traveling around
// timezones.
func (db SongDB) AddSong(song string) (err error) {
	if len(song) == 0 {
		return errors.New("the given song is empty")
	}
	t := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO song(name, addedAt)
	                  VALUES (?, ?)`, song, t)
	return
}

// AddHearing registers that the given song was listened to at the
// current timestamp.
//
// song must exactly match an already existing song from the database.
//
// The current timestamp will be stored with the local timezone, to
// enable sorting hearings by time of day, even when traveling around
// timezones.
func (db SongDB) AddHearing(song string) (err error) {
	if len(song) == 0 {
		return errors.New("the given song is empty")
	}
	t := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO hearing(songID, heardAt)
	                  VALUES (
	                      (SELECT id FROM song WHERE name = ? COLLATE NOCASE), ?
	                  )`, song, t)
	return
}

// AddHearingAndSongIfNeeded registers that the song was listened to
// and, if necessary, adds the song to the database before that.
func (db SongDB) AddHearingAndSongIfNeeded(song string) error {
	if len(song) == 0 {
		return errors.New("the given song is empty")
	}
	err := db.AddSong(song)
	if err != nil {
		// The sqlite3.ErrConstraintUnique just indicates, that the song
		// is already in the database.
		sqliteErr, ok := err.(sqlite3.Error)
		if !ok || sqliteErr.ExtendedCode != sqlite3.ErrConstraintUnique {
			return err
		}
	}
	return db.AddHearing(song)
}

// ListSongsInOrderOfAddition lists all songs in the order they were
// added. Newest additions will be listed first.
func (db SongDB) ListSongsInOrderOfAddition() (songs []string, err error) {
	rows, err := db.Query(`SELECT name FROM song ORDER BY id DESC`)
	if err != nil {
		return
	}
	return extractSongs(rows)
}

// ListSongsInOrderOfLastHearing lists all songs in the order they were
// last heard. The songs that were heard last will be listed first.
func (db SongDB) ListSongsInOrderOfLastHearing() (songs []string, err error) {
	rows, err := db.Query(`SELECT name
	                       FROM (
	                           SELECT songID, MAX(heardAt) heardAt
	                           FROM hearing
	                           GROUP BY (songID)
	                       ) sub
	                       INNER JOIN song ON song.id = sub.songID
	                       ORDER BY sub.heardAt DESC`)
	if err != nil {
		return
	}
	return extractSongs(rows)
}

// ListFavouriteSongs lists all songs, listing those first, that you
// heard most often.
func (db SongDB) ListFavouriteSongs() (songs []string, err error) {
	rows, err := db.Query(`SELECT name FROM hearing
	                       INNER JOIN song ON song.id = hearing.songID
	                       GROUP BY hearing.songID
	                       ORDER BY COUNT(*) DESC`)
	if err != nil {
		return
	}
	return extractSongs(rows)
}

func extractSongs(nameRows *sql.Rows) (songs []string, err error) {
	for nameRows.Next() {
		var song string
		if err = nameRows.Scan(&song); err != nil {
			return
		}
		songs = append(songs, song)
	}
	return
}

// ListFrecentSongs lists songs you lately heard a lot, most frecent first.
func (db SongDB) ListFrecentSongs() (songs []string, err error) {
	// FIXME: If performance becomes an issue: limit results to last year,
	//        or so.
	rows, err := db.Query(`SELECT name, heardAt FROM hearing
	                       INNER JOIN song ON song.id = hearing.songID`)
	if err != nil {
		return
	}
	shs, err := rowsToSongHearings(rows)
	if err != nil {
		return
	}
	return songHearingsToFrecentSongs(shs), nil
}

// ListSuggestions lists songs that you aften hear before or after
// hearing the given song. Best suggestions first.
func (db SongDB) ListSuggestions(song string) (songs []string, err error) {
	rows, err := db.Query(`SELECT name, heardAt FROM hearing
	                       INNER JOIN song ON song.id = hearing.songID`)
	if err != nil {
		return
	}
	shs, err := rowsToSongHearings(rows)
	if err != nil {
		return
	}
	return songHearingsToSuggestions(shs, song)
}

func rowsToSongHearings(rows *sql.Rows) (shs []songHearing, err error) {
	for rows.Next() {
		var name string
		var dateStr string
		var date time.Time
		if err = rows.Scan(&name, &dateStr); err != nil {
			return
		}
		if date, err = time.Parse(time.RFC3339, dateStr); err != nil {
			return
		}
		shs = append(shs, songHearing{name, date})
	}
	return
}

// RemoveLastHearing removes the latest hearing. Fails if there is no
// hearing in the database.
func (db SongDB) RemoveLastHearing() (song string, err error) {
	rows, err := db.Query(`SELECT hearing.id, name FROM hearing
	                       INNER JOIN song ON hearing.songID = song.id
	                       ORDER BY hearing.id DESC
	                       LIMIT 1`)
	if err != nil {
		return
	}
	if rows.Next() == false {
		return "", errors.New("no hearing found")
	}
	var id int64
	if err = rows.Scan(&id, &song); err != nil {
		return
	}
	if err = rows.Close(); err != nil {
		return
	}

	r, err := db.Exec(`DELETE FROM hearing WHERE id = ?`, id)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("hearing got lost in transit")
	}
	return
}

// RemoveLastHearingOf removes the latest hearing of the given song.
// Fails if the song was never heard or the song does not exist.
func (db SongDB) RemoveLastHearingOf(song string) (err error) {
	r, err := db.Exec(`DELETE FROM hearing WHERE id = (
	                       SELECT hearing.id from hearing
	                       INNER JOIN song ON hearing.songID = song.id
	                       WHERE name = ?
	                       ORDER BY hearing.id DESC
	                       LIMIT 1
	                   )`, song)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("no hearing for the song was found")
	}
	return
}

// RemoveSong removes the song with the given name from the database.
// Fails if there is still an entry in the hearing table, that
// references the song.
func (db SongDB) RemoveSong(song string) (err error) {
	r, err := db.Exec(`DELETE FROM song WHERE name = ?`, song)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("song not found")
	}
	return
}

// RemoveLastAddedSong removes the last added song from the database.
// Fails if there is still an entry in the hearing table, that
// references the song.
//
// Returns the removed song's name.
func (db SongDB) RemoveLastAddedSong() (song string, err error) {
	rows, err := db.Query(`SELECT id, name FROM song ORDER BY id DESC LIMIT 1`)
	if err != nil {
		return
	}
	if rows.Next() == false {
		return "", errors.New("no song found")
	}
	var id int64
	if err = rows.Scan(&id, &song); err != nil {
		return
	}
	if err = rows.Close(); err != nil {
		return
	}

	r, err := db.Exec(`DELETE FROM song WHERE id = ?`, id)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("song got lost in transit")
	}
	return
}

// RenameSong renames the given song to newName.
func (db SongDB) RenameSong(song, newName string) (err error) {
	if len(newName) == 0 {
		return errors.New("the new name is empty")
	}
	r, err := db.Exec(`UPDATE song SET name = ? WHERE name = ?;`, newName, song)
	if err != nil {
		return
	}
	n, err := r.RowsAffected()
	if err != nil {
		return
	}
	if n != 1 {
		err = errors.New("song not found")
	}
	return
}
