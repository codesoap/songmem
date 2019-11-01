package songs

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

func InitDB(filepath string) (db *sql.DB, err error) {
	if db, err = sql.Open("sqlite3", filepath); err == nil && db == nil {
		err = errors.New("db is nil")
	}
	return
}

func CreateSchemaIfNotExists(db *sql.DB) (err error) {
	commands := [...]string{
		`CREATE TABLE IF NOT EXISTS song(
		     id      INTEGER PRIMARY KEY AUTOINCREMENT,
		     name    TEXT NOT NULL,
		     addedAt TEXT NOT NULL,
		     CONSTRAINT name_unique UNIQUE(name)
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
func AddSong(db *sql.DB, song string) (err error) {
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
func AddHearing(db *sql.DB, song string) (err error) {
	t := time.Now().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO hearing(songId, heardAt)
	                  VALUES ((SELECT id FROM song WHERE name = ?), ?)`, song, t)
	return
}
