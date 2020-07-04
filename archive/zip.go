package archive

import (
	"archive/zip"
	"fmt"
	"os"
	"slack-backer-upper/slack"
	"slack-backer-upper/storage"
	"strings"
)

func loadZipFiles(
	db storage.ArchiveStorage,
	channelName string,
	files []*zip.File,
	users map[string]slack.StoredUser,
) error {
	channelMessages := make([]slack.StoredMessage, 0, 64)
	for _, f := range files {
		file, err := f.Open()
		if err != nil {
			return err
		}
		defer file.Close()
		messages, err := parseMessages(file, channelName, users)
		if err != nil {
			return nil
		}
		channelMessages = append(channelMessages, messages...)
	}
	for _, msg := range channelMessages {
		if err := db.InsertMessage(channelName, msg); err != nil {
			return err
		}
	}
	return nil
}

func importZip(reader zip.Reader) error {
	var users slack.Users
	files := make(map[string][]*zip.File)
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		if users == nil && f.Name == "users.json" {
			userFile, err := f.Open()
			if err != nil {
				return err
			}
			users, err = parseUsers(userFile)
			if err != nil {
				return err
			}
		} else {
			nameParts := strings.Split(f.Name, string(os.PathSeparator))
			if len(nameParts) != 2 {
				continue
			}
			files[nameParts[0]] = append(files[nameParts[0]], f)
		}
	}

	db, err := storage.NewArchiveStorage()
	if err != nil {
		return fmt.Errorf("Error initializing database: %v", err)
	}
	defer db.Close()

	results := make(chan error)
	for channelName, files := range files {
		go func(channelName string, files []*zip.File) {
			results <- loadZipFiles(db, channelName, files, users)
		}(channelName, files)
	}

	if serr := db.InsertUsers(users); serr != nil {
		err = serr
	}
	for i := 0; i < len(files); i++ {
		completedErr := <-results
		if completedErr != nil {
			err = completedErr
		}
	}
	return err
}

// ImportZipFile imports messages and users from the provided zip file
func ImportZipFile(src string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	return importZip(r.Reader)
}
