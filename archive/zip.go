package archive

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"slack-backer-upper/slack"
	"strings"
)

func (a *Archiver) loadZipFiles(
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
			return fmt.Errorf("Error parsing messages in %s: %v", file, err)
		}
		channelMessages = append(channelMessages, messages...)
	}
	for _, msg := range channelMessages {
		if err := a.storage.AddMessage(channelName, msg); err != nil {
			return fmt.Errorf("Error adding message: %v", err)
		}
	}
	return nil
}

// ImportZip imports messages and users from the provided zip.Reader
func (a *Archiver) ImportZip(reader zip.Reader) error {
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
				return fmt.Errorf("Error parsing users: %v", err)
			}
		} else {
			nameParts := strings.Split(f.Name, string(os.PathSeparator))
			if len(nameParts) != 2 {
				continue
			}
			files[nameParts[0]] = append(files[nameParts[0]], f)
		}
	}
	if users == nil {
		return fmt.Errorf("Users file missing")
	}

	results := make(chan error)
	for channelName, files := range files {
		go func(channelName string, files []*zip.File) {
			results <- a.loadZipFiles(channelName, files, users)
		}(channelName, files)
	}

	var err error
	if serr := a.storage.AddUsers(users); serr != nil {
		err = fmt.Errorf("Error adding users: %v", serr)
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
func (a *Archiver) ImportZipFile(src string) error {
	log.Printf("Importing from file %s", src)
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()
	return a.ImportZip(r.Reader)
}
