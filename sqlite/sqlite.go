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

// Sqlite is a handle to the SQLite database plus prepared statements
type Sqlite struct {
	db             *sql.DB
	addMessage     *sql.Stmt
	addUser        *sql.Stmt
	updateChildren *sql.Stmt
}

// Close closes the Sqlite hadle
func (s *Sqlite) Close() error {
	if err := s.addMessage.Close(); err != nil {
		return err
	}
	if err := s.addUser.Close(); err != nil {
		return err
	}
	if err := s.updateChildren.Close(); err != nil {
		return err
	}
	return s.db.Close()
}

// NewSqlite creates and returns a handle to the initialized SQLite database,
// creates the necessary tables, and prepares the necessary statements
func NewSqlite() (Sqlite, error) {
	db, err := sql.Open("sqlite3", "./slack.db?_journal=WAL")
	if err != nil {
		return Sqlite{}, err
	}
	if _, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages
			(channel TEXT NOT NULL, timestamp TEXT NOT NULL, txt TEXT, user TEXT, attachments TEXT, reacts TEXT, children TEXT);
		CREATE TABLE IF NOT EXISTS users
			(id TEXT, real_name TEXT, display_name TEXT);
	`); err != nil {
		db.Close()
		return Sqlite{}, err
	}
	addMessage, err := db.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		db.Close()
		return Sqlite{}, err
	}
	addUser, err := db.Prepare("INSERT OR IGNORE INTO users VALUES (?, ?, ?)")
	if err != nil {
		addMessage.Close()
		db.Close()
		return Sqlite{}, err
	}
	updateChildren, err := db.Prepare(`
		UPDATE messages SET children = (
			SELECT children FROM messages WHERE channel = ? AND timestamp = ?
		) || ',' || ? WHERE channel = ? AND timestamp = ?;
	`)
	if err != nil {
		addMessage.Close()
		addUser.Close()
		db.Close()
		return Sqlite{}, err
	}
	return Sqlite{
		db:             db,
		addMessage:     addMessage,
		addUser:        addUser,
		updateChildren: updateChildren,
	}, nil
}

// UpdateMessage updates a message in the DB with msg's new children
func (s *Sqlite) UpdateMessage(channelName string, msg slack.StoredMessage) error {
	if _, err := s.updateChildren.Exec(
		channelName, msg.Timestamp, strings.Join(msg.Thread, ","), channelName, msg.Timestamp,
	); err != nil {
		return fmt.Errorf("Error updating message %#v: %v", msg, err)
	}
	return nil
}

// InsertMessage inserts msg into the DB associated with channelName
func (s *Sqlite) InsertMessage(channelName string, msg slack.StoredMessage) error {
	attach, err := json.Marshal(msg.Attachments)
	if err != nil {
		return err
	}
	reacc, err := json.Marshal(msg.Reacts)
	if err != nil {
		return err
	}
	if _, err = s.addMessage.Exec(
		channelName, msg.Timestamp, msg.Text, msg.User, attach, reacc, strings.Join(msg.Thread, ","),
	); err != nil {
		return fmt.Errorf("Error inserting new message %#v: %v", msg, err)
	}
	return nil
}

// InsertUsers inserts users into the DB
func (s *Sqlite) InsertUsers(users map[string]slack.StoredUser) error {
	for id, user := range users {
		if _, err := s.addUser.Exec(id, user.RealName, user.DisplayName); err != nil {
			return fmt.Errorf("Error inserting user: %v", err)
		}
	}
	return nil
}
