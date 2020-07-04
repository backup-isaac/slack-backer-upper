package archive

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"

	"slack-backer-upper/slack"
)

// ImportFolder imports messages and users from the named folder
func (a *Archiver) ImportFolder(name string) error {
	log.Printf("Importing from folder %s...", name)
	userFile, err := os.Open(path.Join(name, "users.json"))
	if err != nil {
		return err
	}
	defer userFile.Close()
	users, err := parseUsers(userFile)
	if err != nil {
		return fmt.Errorf("Error parsing users: %v", err)
	}
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return err
	}
	results := make(chan error)
	goroutines := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			goroutines++
			go func(fileName string) {
				results <- a.loadFolder(name, fileName, users)
			}(entry.Name())
		}
	}
	if serr := a.storage.AddUsers(users); serr != nil {
		err = fmt.Errorf("Error adding users: %v", serr)
	}
	for i := 0; i < goroutines; i++ {
		completedErr := <-results
		if completedErr != nil {
			err = completedErr
		}
	}
	return err
}

func (a *Archiver) loadFolder(
	dirname string,
	channelName string,
	users map[string]slack.StoredUser,
) error {
	inPath := path.Join(dirname, channelName)
	files, err := ioutil.ReadDir(inPath)
	if err != nil {
		return err
	}
	channelMessages := make([]slack.StoredMessage, 0, 64)
	for _, f := range files {
		file, err := os.Open(path.Join(inPath, f.Name()))
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
		if err = a.storage.AddMessage(channelName, msg); err != nil {
			return fmt.Errorf("Error adding message: %v", err)
		}
	}
	return nil
}
