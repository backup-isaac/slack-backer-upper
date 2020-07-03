package sqlite

import (
	"database/sql"
	"encoding/json"
	"slack-backer-upper/slack"
	"time"
)

// ViewerStorage is a handle to the database plus prepared statementst for
// viewing archives
type ViewerStorage struct {
	db          *sql.DB
	getMessages *sql.Stmt
}

// Close closes the storage hadle
func (s *ViewerStorage) Close() error {
	if err := s.getMessages.Close(); err != nil {
		return err
	}
	return s.db.Close()
}

// NewViewerStorage creates and returns a handle to the initialized database,
// creates the necessary tables, and prepares the necessary statements
func NewViewerStorage() (ViewerStorage, error) {
	db, err := initializeDb()
	if err != nil {
		return ViewerStorage{}, err
	}
	getMessages, err := db.Prepare(`
		SELECT timestamp, txt, user, attachments, reacts, parent, top_level
			FROM messages WHERE channel = ? AND timestamp >= ? AND timestamp < ?
			ORDER BY timestamp;
	`)
	if err != nil {
		return ViewerStorage{}, nil
	}
	return ViewerStorage{
		db:          db,
		getMessages: getMessages,
	}, err
}

// ListChannels enumerates the channels in the storage
func (s *ViewerStorage) ListChannels() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT channel FROM messages ORDER BY channel")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	channels := make([]string, 0, 64)
	for rows.Next() {
		var channel string
		if err := rows.Scan(&channel); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

// GetMessages gets messages in a channel during the specified time interval
func (s *ViewerStorage) GetMessages(channel string, from, to time.Time) ([]slack.StoredMessage, error) {
	fromSecs := float64(from.UnixNano()) / 1e9
	toSecs := float64(to.UnixNano()) / 1e9
	rows, err := s.getMessages.Query(channel, fromSecs, toSecs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	messages := make([]slack.StoredMessage, 0, 64)
	for rows.Next() {
		var msg slack.StoredMessage
		var attachJSON, reactsJSON []byte
		if err = rows.Scan(
			&msg.Timestamp, &msg.Text, &msg.User, &attachJSON, &reactsJSON, &msg.ParentTimestamp, &msg.DisplayTopLevel,
		); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attachJSON, &msg.Attachments); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(reactsJSON, &msg.Reacts); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}
