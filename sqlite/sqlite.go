package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slack-backer-upper/slack"
	"strings"

	// go database drivers require _ import
	_ "github.com/mattn/go-sqlite3"
)

// InitDB creates and returns a handle to the initialized SQLite database
// and creates the necessary tables
func InitDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./slack.db?_journal=WAL")
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS messages (channel TEXT NOT NULL, timestamp TEXT NOT NULL, txt TEXT, user TEXT, attachments TEXT, reacts TEXT, children TEXT)",
	); err != nil {
		return db, err
	}
	if _, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS users (id TEXT, real_name TEXT, display_name TEXT)",
	); err != nil {
		return db, err
	}

	return db, nil
}

var (
	addMessage *sql.Stmt
	addUser    *sql.Stmt
)

// PrepareQueries prepares statements to use for repeated queries
func PrepareQueries(db *sql.DB) error {
	var err error
	addMessage, err = db.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	addUser, err = db.Prepare("INSERT OR IGNORE INTO users VALUES (?, ?, ?)")
	return err
}

// CloseQueries closes those prepared statements
func CloseQueries(db *sql.DB) error {
	var err error
	if addMessage != nil {
		err = addMessage.Close()
	}
	if addUser != nil {
		err = addUser.Close()
	}
	return err
}

// InsertMessage inserts msg into the DB associated with channelName
func InsertMessage(channelName string, msg slack.StoredMessage) error {
	attach, err := json.Marshal(msg.Attachments)
	if err != nil {
		return err
	}
	reacc, err := json.Marshal(msg.Reacts)
	if err != nil {
		return err
	}
	if _, err = addMessage.Exec(
		channelName, msg.Timestamp, msg.Text, msg.User, attach, reacc, strings.Join(msg.Thread, ","),
	); err != nil {
		return fmt.Errorf("Error inserting new message %#v: %v", msg, err)
	}
	return nil
}

// InsertUsers inserts users into the DB
func InsertUsers(users map[string]slack.StoredUser) error {
	for id, user := range users {
		if _, err := addUser.Exec(id, user.RealName, user.DisplayName); err != nil {
			return fmt.Errorf("Error inserting user: %v", err)
		}
	}
	return nil
}
