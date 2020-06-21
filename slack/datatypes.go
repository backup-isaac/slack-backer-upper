package slack

import "regexp"

// Attachment is what we care about from attachments
// Also goes in the DB
type Attachment struct {
	URL      string `json:"from_url"`
	Fallback string `json:"fallback"`
	Title    string `json:"title"`
}

// File is what we care about from file uploads
type File struct {
	URL   string `json:"permalink"`
	Title string `json:"title"`
}

// React is what we care about from reaccs
type React struct {
	Name  string   `json:"name"`
	Users []string `json:"users"`
}

// StoredMessage goes in the db
type StoredMessage struct {
	Timestamp       string
	Text            string
	User            string
	ParentTimestamp string
	Subtype         string
	Attachments     []Attachment
	Reacts          map[string][]string
	Thread          []string
}

// RawMessage is what we care about from Slack
type RawMessage struct {
	Timestamp       string       `json:"ts"`
	Text            string       `json:"text"`
	User            string       `json:"user"`
	Username        string       `json:"username"`
	ParentTimestamp string       `json:"thread_ts"`
	Subtype         string       `json:"subtype"`
	Attachments     []Attachment `json:"attachments"`
	Files           []File       `json:"files"`
	Reacts          []React      `json:"reactions"`
}

// StoredUser goes in the db
type StoredUser struct {
	RealName    string `json:"real_name"`
	DisplayName string `json:"display_name"`
}

// RawUser is what we care about from Slack
type RawUser struct {
	Profile StoredUser `json:"profile"`
	ID      string     `json:"id"`
}

var (
	isComment      = regexp.MustCompile("<@U[A-Z0-9]{8}> commented on .+")
	atNotification = regexp.MustCompile("<@U[A-Z0-9]{8}>")
)

// FilterRawMessage transforms a RawMessage into a StoredMessage
// and replaces user IDs
func FilterRawMessage(message RawMessage, users map[string]StoredUser) StoredMessage {
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

	ret := StoredMessage{
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

	attachments := make([]Attachment, 0, 4)
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
				attachments = append(attachments, Attachment{
					URL:   file.URL,
					Title: file.Title,
				})
			} else {
				attachments = append(attachments, Attachment{
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
