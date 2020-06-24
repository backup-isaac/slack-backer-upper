package sqlite

import "database/sql"

// ViewerStorage is a handle to the database plus prepared statementst for
// viewing archives
type ViewerStorage struct {
	db *sql.DB
}

// Close closes the storage hadle
func (s *ViewerStorage) Close() error {
	return s.db.Close()
}

// NewViewerStorage creates and returns a handle to the initialized database,
// creates the necessary tables, and prepares the necessary statements
func NewViewerStorage() (ViewerStorage, error) {
	db, err := initializeDb()
	return ViewerStorage{
		db: db,
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
