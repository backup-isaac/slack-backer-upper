package storage

import (
	"database/sql"

	// go database drivers require _ import
	_ "github.com/mattn/go-sqlite3"
)

// Storage is an interface to persistent message storage
type Storage interface {
	Close() error
}

// New creates a new Storage backed by SQLite
func New() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./slack.db?_journal=WAL")
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			channel TEXT NOT NULL, timestamp TEXT NOT NULL, txt TEXT, user TEXT,
			attachments TEXT, reacts TEXT, parent TEXT, top_level BOOLEAN,
			UNIQUE(channel, timestamp)
		);
		CREATE TABLE IF NOT EXISTS users (
			id TEXT UNIQUE, real_name TEXT, display_name TEXT
		);
	`); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
