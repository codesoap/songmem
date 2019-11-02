package songs

import (
	"database/sql"
	"errors"
	sqlite3 "github.com/mattn/go-sqlite3"
	"time"
)

type SongDB struct {
	*sql.DB
}

func InitDB(filepath string) (SongDB, error) {
	db, err := sql.Open("sqlite3", filepath)
	if err == nil && db == nil {
		err = errors.New("db is nil")
	}
	return SongDB{db}, err
}

func (db SongDB) CreateSchemaIfNotExists() (err error) {
	commands := [...]string{
		`CREATE TABLE IF NOT EXISTS song(
		     id      INTEGER PRIMARY KEY AUTOINCREMENT,
		     name    TEXT NOT NULL,
		     addedAt TEXT NOT NULL,
		     CONSTRAINT name_unique UNIQUE(name COLLATE NOCASE)
		 )`,
		`CREATE INDEX IF NOT EXISTS song_name ON song(name)`,
		`CREATE INDEX IF NOT EXISTS song_addedAt ON song(addedAt)`,
		`CREATE TABLE IF NOT EXISTS hearing(
		     songId  INTEGER NOT NULL,
		     heardAt TEXT NOT NULL,
		     FOREIGN KEY(songId) REFERENCES song(id)
		 )`,
		`CREATE INDEX IF NOT EXISTS hearing_songId ON hearing(songId)`,
		`CREATE INDEX IF NOT EXISTS hearing_heardAt ON hearing(heardAt)`}

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
	t := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO hearing(songId, heardAt)
	                  VALUES (
	                      (SELECT id FROM song WHERE name = ? COLLATE NOCASE), ?
	                  )`, song, t)
	return
}

// AddHearingAndSongIfNeeded registers that the song was listened to
// and, if necessary, adds the song to the database before that.
func (db SongDB) AddHearingAndSongIfNeeded(song string) error {
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
	rows, err := db.Query(`SELECT DISTINCT name FROM hearing
	                       INNER JOIN song ON song.id = hearing.songId
	                       ORDER BY heardAt DESC`)
	if err != nil {
		return
	}
	return extractSongs(rows)
}

// ListFavouriteSongs lists all songs, listing those first, that you
// heard most often.
func (db SongDB) ListFavouriteSongs() (songs []string, err error) {
	rows, err := db.Query(`SELECT name FROM hearing
	                       INNER JOIN song ON song.id = hearing.songId
	                       GROUP BY hearing.songId
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

// ListFrecentSongs list songs you lately heard a lot, most frecent first.
func (db SongDB) ListFrecentSongs() (songs []string, err error) {
	// FIXME: If performance becomes an issue: limit results to last year,
	//        or so.
	rows, err := db.Query(`SELECT name, heardAt FROM hearing
	                       INNER JOIN song ON song.id = hearing.songId`)
	if err != nil {
		return
	}

	var fis []frecencyInput
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
		fis = append(fis, frecencyInput{name, date})
	}
	return frecencyInputsToSongs(fis), nil
}
