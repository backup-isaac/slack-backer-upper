package files

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"sort"
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
	db, err := sqlite.InitDB()
	if db != nil {
		defer db.Close()
	}
	if err != nil {
		return fmt.Errorf("Error initializing database: %v", err)
	}
	entries, err := ioutil.ReadDir(name)
	if err != nil {
		return fmt.Errorf("Error listing folder contents: %v", err)
	}
	err = sqlite.PrepareQueries(db)
	defer sqlite.CloseQueries(db)
	if err != nil {
		return fmt.Errorf("Error preparing queries: %v", err)
	}
	results := make(chan error)
	goroutines := 0
	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			goroutines++
			go importChannel(name, entry.Name(), users, results)
		}
	}
	if serr := sqlite.InsertUsers(users); serr != nil {
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
	channelMessages := make(map[string]slack.StoredMessage)
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
			if msg.ParentTimestamp != "" && msg.ParentTimestamp != msg.Timestamp {
				parent, ok := channelMessages[msg.ParentTimestamp]
				if !ok {
					channelMessages[msg.ParentTimestamp] = slack.StoredMessage{
						Timestamp: msg.ParentTimestamp,
						Thread:    []string{msg.Timestamp},
					}
				} else {
					parent.Thread = append(parent.Thread, msg.Timestamp)
					sort.Strings(parent.Thread)
				}
			}
			channelMessages[msg.Timestamp] = msg
		}
	}
	for _, msg := range channelMessages {
		if err = sqlite.InsertMessage(strings.TrimSuffix(channelName, ".json"), msg); err != nil {
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
