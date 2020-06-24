package sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slack-backer-upper/slack"
	"strings"
)

// ArchiveStorage is a handle to the database plus prepared statements for
// adding to the archive
type ArchiveStorage struct {
	db             *sql.DB
	addMessage     *sql.Stmt
	addUser        *sql.Stmt
	updateChildren *sql.Stmt
}

// Close closes the storage hadle
func (s *ArchiveStorage) Close() error {
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

// NewArchiveStorage creates and returns a handle to the initialized database,
// creates the necessary tables, and prepares the necessary statements
func NewArchiveStorage() (ArchiveStorage, error) {
	db, err := initializeDb()
	if err != nil {
		return ArchiveStorage{}, err
	}
	addMessage, err := db.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		db.Close()
		return ArchiveStorage{}, err
	}
	addUser, err := db.Prepare("INSERT OR IGNORE INTO users VALUES (?, ?, ?)")
	if err != nil {
		addMessage.Close()
		db.Close()
		return ArchiveStorage{}, err
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
		return ArchiveStorage{}, err
	}
	return ArchiveStorage{
		db:             db,
		addMessage:     addMessage,
		addUser:        addUser,
		updateChildren: updateChildren,
	}, nil
}

// UpdateMessage updates a message in the DB with msg's new children
func (s *ArchiveStorage) UpdateMessage(channelName string, msg slack.StoredMessage) error {
	if _, err := s.updateChildren.Exec(
		channelName, msg.Timestamp, strings.Join(msg.Thread, ","), channelName, msg.Timestamp,
	); err != nil {
		return fmt.Errorf("Error updating message %#v: %v", msg, err)
	}
	return nil
}

// InsertMessage inserts msg into the DB associated with channelName
func (s *ArchiveStorage) InsertMessage(channelName string, msg slack.StoredMessage) error {
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
func (s *ArchiveStorage) InsertUsers(users map[string]slack.StoredUser) error {
	for id, user := range users {
		if _, err := s.addUser.Exec(id, user.RealName, user.DisplayName); err != nil {
			return fmt.Errorf("Error inserting user: %v", err)
		}
	}
	return nil
}
