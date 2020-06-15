package manual

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"regexp"
	"sort"
	"strings"

	// go database drivers require _ import
	_ "github.com/mattn/go-sqlite3"
)

type storedUser struct {
	RealName    string `json:"real_name"`
	DisplayName string `json:"display_name"`
}

type importedUser struct {
	Profile storedUser `json:"profile"`
	ID      string     `json:"id"`
}

type attachment struct {
	URL      string `json:"from_url"`
	Fallback string `json:"fallback"`
	Title    string `json:"title"`
}

type file struct {
	URL   string `json:"permalink"`
	Title string `json:"title"`
}

type react struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

type importedMessage struct {
	Timestamp       string       `json:"ts"`
	Text            string       `json:"text"`
	User            string       `json:"user"`
	Username        string       `json:"username"`
	ParentTimestamp string       `json:"thread_ts"`
	Subtype         string       `json:"subtype"`
	Attachments     []attachment `json:"attachments"`
	Files           []file       `json:"files"`
	Reacts          []react      `json:"reactions"`
}

type storedMessage struct {
	Timestamp       string
	Text            string
	User            string
	ParentTimestamp string
	Subtype         string
	Attachments     []attachment
	Reacts          map[string][]string
	Thread          []string
}

// ImportBackup imports a backup at backupDir
func ImportBackup(backupDir string) error {
	users, err := loadUsers(backupDir)
	if err != nil {
		return fmt.Errorf("Error loading users: %v", err)
	}
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("Error initializing database: %v", err)
	}

	entries, err := ioutil.ReadDir(backupDir)
	if err != nil {
		return err
	}

	insertMessage, err := db.Prepare("INSERT INTO messages VALUES (?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer insertMessage.Close()

	// var wg sync.WaitGroup

	results := make(chan error)
	goroutines := 0

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			// wg.Add(1)
			goroutines++
			go func(channelName string, resultsChannel chan error) {
				inPath := path.Join(backupDir, channelName)
				files, err := ioutil.ReadDir(inPath)
				if err != nil {
					resultsChannel <- err
					return
				}
				channelMessages := make(map[string]storedMessage)
				for _, file := range files {
					fileIn, err := ioutil.ReadFile(path.Join(inPath, file.Name()))
					if err != nil {
						resultsChannel <- err
						return
					}
					messagesIn := make([]importedMessage, 0, 16)
					err = json.Unmarshal(fileIn, &messagesIn)
					if err != nil {
						resultsChannel <- err
						return
					}
					for _, beegMsg := range messagesIn {
						msg := filterFields(beegMsg, users)
						if msg.ParentTimestamp != "" && msg.ParentTimestamp != msg.Timestamp {
							parent, ok := channelMessages[msg.ParentTimestamp]
							if !ok {
								channelMessages[msg.ParentTimestamp] = storedMessage{
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
					attach, err := json.Marshal(msg.Attachments)
					if err != nil {
						resultsChannel <- err
						return
					}
					reacc, err := json.Marshal(msg.Reacts)
					if err != nil {
						resultsChannel <- err
						return
					}
					var children sql.NullString
					if msg.ParentTimestamp == "" || msg.ParentTimestamp == msg.Timestamp || msg.Subtype == "thread_broadcast" {
						children = sql.NullString{
							String: "[]",
							Valid:  true,
						}
					}
					if _, err = insertMessage.Exec(
						strings.TrimSuffix(channelName, ".json"), msg.Timestamp, msg.Text, msg.User, attach, reacc, children,
					); err != nil {
						resultsChannel <- fmt.Errorf("Error inserting new message %#v: %v", msg, err)
						return
					}
				}
				resultsChannel <- nil
			}(entry.Name(), results)
		}
	}

	addUser, err := db.Prepare("INSERT OR IGNORE INTO users VALUES (?, ?, ?)")
	if err != nil {
		return err
	}
	defer addUser.Close()

	for id, user := range users {
		if _, err = addUser.Exec(id, user.RealName, user.DisplayName); err != nil {
			return err
		}
	}

	for i := 0; i < goroutines; i++ {
		completedError := <-results
		if completedError != nil {
			err = completedError
		}
	}

	return err
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./slack.db?_journal=WAL")
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS messages (channel TEXT NOT NULL, timestamp TEXT NOT NULL, txt TEXT, user TEXT, attachments TEXT, reacts TEXT, children TEXT)",
	); err != nil {
		return nil, err
	}
	if _, err = db.Exec(
		"CREATE TABLE IF NOT EXISTS users (id TEXT, real_name TEXT, display_name TEXT)",
	); err != nil {
		return nil, err
	}
	return db, nil
}

func loadUsers(backupDir string) (map[string]storedUser, error) {
	users := map[string]storedUser{
		"USLACKBOT": {"Slackbot", "Slackbot"},
	}
	rawJSON, err := ioutil.ReadFile(path.Join(backupDir, "users.json"))
	if err != nil {
		return nil, err
	}
	userList := make([]importedUser, 0, 128)
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

var (
	isComment      = regexp.MustCompile("<@U[A-Z0-9]{8}> commented on .+")
	atNotification = regexp.MustCompile("<@U[A-Z0-9]{8}>")
)

func filterFields(message importedMessage, users map[string]storedUser) storedMessage {
	var userid string
	if message.User != "" {
		userid = message.User
	} else if isComment.MatchString(message.Text) {
		userid = message.Text[2:11]
	} else {
		userid = message.Username
	}
	user, ok := users[userid]
	if ok {
		userid = user.RealName
	}

	ret := storedMessage{
		Timestamp: message.Timestamp,
		Text: atNotification.ReplaceAllStringFunc(message.Text, func(match string) string {
			user, ok := users[match[2:11]]
			if ok {
				return "@" + user.DisplayName
			}
			return "@<unknown>"
		}),
		User:            userid,
		ParentTimestamp: message.ParentTimestamp,
		Subtype:         message.Subtype,
	}

	attachments := make([]attachment, 0, 4)
	if message.Attachments != nil {
		for _, attach := range message.Attachments {
			if attach.URL != "" || attach.Fallback != "" {
				attachments = append(attachments, attach)
			}
		}
	}
	if message.Files != nil {
		for _, file := range message.Files {
			if file.URL != "" {
				attachments = append(attachments, attachment{
					URL:   file.URL,
					Title: file.Title,
				})
			} else {
				attachments = append(attachments, attachment{
					Title: "This file was deleted.",
				})
			}
		}
	}
	if len(attachments) > 0 {
		ret.Attachments = attachments
	}
	if message.Reacts != nil {
		ret.Reacts = make(map[string][]string)
		for _, reacc := range message.Reacts {
			names := make([]string, len(reacc.Users))
			for i, s := range reacc.Users {
				names[i] = users[s].RealName
			}
			ret.Reacts[reacc.Name] = names
		}
	}
	return ret
}
