package archive

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"slack-backer-upper/slack"
)

func parseMessages(source io.Reader, channel string, users slack.Users) ([]slack.StoredMessage, error) {
	contents, err := ioutil.ReadAll(source)
	if err != nil {
		return nil, err
	}
	messagesIn := make([]slack.RawMessage, 0, 8)
	err = json.Unmarshal(contents, &messagesIn)
	if err != nil {
		return nil, err
	}
	messagesOut := make([]slack.StoredMessage, len(messagesIn))
	for i, msg := range messagesIn {
		messagesOut[i] = slack.FilterRawMessage(msg, users)
	}
	return messagesOut, err
}

func parseUsers(source io.Reader) (slack.Users, error) {
	users := slack.Users{
		"USLACKBOT": {
			RealName:    "Slackbot",
			DisplayName: "Slackbot",
		},
	}
	rawJSON, err := ioutil.ReadAll(source)
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
