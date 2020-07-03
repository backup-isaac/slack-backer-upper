package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"slack-backer-upper/slack"
	"slack-backer-upper/sqlite"
)

// ImportFolder imports messages and users from the named folder
func ImportFolder(name string) error {
	users, err := loadUsers(name)
	if err != nil {
		return fmt.Errorf("Error loading users: %v", err)
	}
	db, err := sqlite.NewArchiveStorage()
	if err != nil {
		return fmt.Errorf("Error initializing database: %v", err)
	}
	defer db.Close()
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return fmt.Errorf("Error listing folder contents: %v", err)
	}
	if err != nil {
		return fmt.Errorf("Error preparing queries: %v", err)
	}
	results := make(chan error)
	goroutines := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			goroutines++
			go importChannel(db, name, entry.Name(), users, results)
		}
	}
	if serr := db.InsertUsers(users); serr != nil {
		err = serr
	}
	for i := 0; i < goroutines; i++ {
		completedErr := <-results
		if completedErr != nil {
			err = completedErr
		}
	}
	return err
}

func importChannel(
	db sqlite.ArchiveStorage,
	dirname string,
	channelName string,
	users map[string]slack.StoredUser,
	resultsChannel chan error,
) {
	inPath := path.Join(dirname, channelName)
	files, err := ioutil.ReadDir(inPath)
	if err != nil {
		resultsChannel <- err
		return
	}
	channelMessages := make([]slack.StoredMessage, 0, 64)
	for _, file := range files {
		fileIn, err := ioutil.ReadFile(path.Join(inPath, file.Name()))
		if err != nil {
			resultsChannel <- err
			return
		}
		messagesIn := make([]slack.RawMessage, 0, 16)
		err = json.Unmarshal(fileIn, &messagesIn)
		if err != nil {
			resultsChannel <- err
			return
		}
		for _, beegMsg := range messagesIn {
			msg := slack.FilterRawMessage(beegMsg, users)
			channelMessages = append(channelMessages, msg)
		}
	}
	for _, msg := range channelMessages {
		if err = db.InsertMessage(channelName, msg); err != nil {
			resultsChannel <- err
			return
		}
	}
	resultsChannel <- nil
}

func loadUsers(backupDir string) (map[string]slack.StoredUser, error) {
	users := map[string]slack.StoredUser{
		"USLACKBOT": {
			RealName:    "Slackbot",
			DisplayName: "Slackbot",
		},
	}
	rawJSON, err := ioutil.ReadFile(path.Join(backupDir, "users.json"))
	if err != nil {
		return nil, err
	}
	userList := make([]slack.RawUser, 0, 128)
	if err = json.Unmarshal(rawJSON, &userList); err != nil {
		return nil, err
	}
	for _, user := range userList {
		if user.Profile.DisplayName == "" {
			user.Profile.DisplayName = user.Profile.RealName
		}
		users[user.ID] = user.Profile
	}
	return users, nil
}
