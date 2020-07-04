package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"slack-backer-upper/slack"
)

// ArchiveDBHandle is a handle to the database plus resources
// needed to handle inserting information into it
type ArchiveDBHandle struct {
	db         *sql.DB
	addMessage *sql.Stmt
	addUser    *sql.Stmt
}

// Close closes resources specific to the ArchiveDBHandle
// but not the underlying DB itself
func (d *ArchiveDBHandle) Close() error {
	if err := d.addMessage.Close(); err != nil {
		return err
	}
	return d.addUser.Close()
}

// Archiver creates and returns a handle to the initialized database,
// creates the necessary tables, and prepares the necessary statements
func Archiver(db *sql.DB) (*ArchiveDBHandle, error) {
	addMessage, err := db.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}
	addUser, err := db.Prepare("INSERT OR IGNORE INTO users VALUES (?, ?, ?)")
	if err != nil {
		addMessage.Close()
		return nil, err
	}
	return &ArchiveDBHandle{
		db:         db,
		addMessage: addMessage,
		addUser:    addUser,
	}, nil
}

// AddMessage adds msg into the DB associated with channelName
func (d *ArchiveDBHandle) AddMessage(channelName string, msg slack.StoredMessage) error {
	attach, err := json.Marshal(msg.Attachments)
	if err != nil {
		return err
	}
	reacc, err := json.Marshal(msg.Reacts)
	if err != nil {
		return err
	}
	if _, err = d.addMessage.Exec(
		channelName, msg.Timestamp, msg.Text, msg.User, attach, reacc, msg.ParentTimestamp, msg.DisplayTopLevel,
	); err != nil {
		return err
	}
	return nil
}

// AddUsers inserts users into the DB
func (d *ArchiveDBHandle) AddUsers(users slack.Users) error {
	for id, user := range users {
		if _, err := d.addUser.Exec(id, user.RealName, user.DisplayName); err != nil {
			return fmt.Errorf("Error inserting user: %v", err)
		}
	}
	return nil
}
