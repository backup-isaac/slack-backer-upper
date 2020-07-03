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
	getReplies  *sql.Stmt
}

// Close closes the storage hadle
func (s *ViewerStorage) Close() error {
	if err := s.getMessages.Close(); err != nil {
		return err
	}
	if err := s.getReplies.Close(); err != nil {
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
		SELECT timestamp, txt, user, attachments, reacts FROM messages
			WHERE channel = ? AND timestamp >= ? AND timestamp < ? AND top_level = true AND parent = ""
			ORDER BY timestamp;
	`)
	if err != nil {
		db.Close()
		return ViewerStorage{}, nil
	}
	getReplies, err := db.Prepare(`
		SELECT timestamp, txt, user, attachments, reacts, top_level FROM messages
			WHERE channel = ? AND parent = ? ORDER BY timestamp;
	`)
	if err != nil {
		getMessages.Close()
		db.Close()
		return ViewerStorage{}, nil
	}
	return ViewerStorage{
		db:          db,
		getMessages: getMessages,
		getReplies:  getReplies,
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

// GetParentMessages gets the parent messages in a channel
// (i.e. messages not replying in a thread)
// during the specified time interval
func (s *ViewerStorage) GetParentMessages(channel string, from, to time.Time) ([]slack.StoredMessage, error) {
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
			&msg.Timestamp, &msg.Text, &msg.User, &attachJSON, &reactsJSON,
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

// GetThreadReplies gets the replies to the specified message
func (s *ViewerStorage) GetThreadReplies(channel string, parentTimestamp string) ([]slack.ThreadMessage, error) {
	rows, err := s.getReplies.Query(channel, parentTimestamp)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	replies := make([]slack.ThreadMessage, 0, 4)
	for rows.Next() {
		var msg slack.ThreadMessage
		var attachJSON, reactsJSON []byte
		if err = rows.Scan(
			&msg.Timestamp, &msg.Text, &msg.User, &attachJSON, &reactsJSON, &msg.SentToChannel,
		); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(attachJSON, &msg.Attachments); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(reactsJSON, &msg.Reacts); err != nil {
			return nil, err
		}
		replies = append(replies, msg)
	}
	if len(replies) == 0 {
		return nil, nil
	}
	return replies, nil
}
