package storage

import (
	"database/sql"

	// go database drivers require _ import
	_ "github.com/mattn/go-sqlite3"
)

func initializeDb() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./slack.db?_journal=WAL")
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages
			(channel TEXT NOT NULL, timestamp TEXT NOT NULL, txt TEXT, user TEXT,
			attachments TEXT, reacts TEXT, parent TEXT, top_level BOOLEAN);
		CREATE TABLE IF NOT EXISTS users
			(id TEXT, real_name TEXT, display_name TEXT);
	`); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
