package slack

import "regexp"

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
		DisplayTopLevel: message.ParentTimestamp == "" || message.ParentTimestamp == message.Timestamp || message.Subtype == "thread_broadcast",
	}

	if message.ParentTimestamp != "" && message.ParentTimestamp != message.Timestamp {
		ret.ParentTimestamp = message.ParentTimestamp
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

// ParentMessageFromStored creates a ParentMessage from a StoredMessage
func ParentMessageFromStored(message StoredMessage) ParentMessage {
	return ParentMessage{
		Timestamp: message.Timestamp,
		Text: message.Text,
		User: message.User,
		Attachments: message.Attachments,
		Reacts: message.Reacts,
	}
}
