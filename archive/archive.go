package archive

import (
	"slack-backer-upper/slack"
)

type archiveStorage interface {
	AddMessage(channelName string, msg slack.StoredMessage) error
	AddUsers(users slack.Users) error
}

// Archiver adds messages to an archive
type Archiver struct {
	storage archiveStorage
}

// New creates an Archiver with the provided storage
func New(s archiveStorage) Archiver {
	return Archiver{
		storage: s,
	}
}
