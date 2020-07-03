package slack

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
	DisplayTopLevel bool
	Attachments     []Attachment
	Reacts          map[string][]string
}

// ThreadMessage is returned from the API / to the front end
type ThreadMessage struct {
	Timestamp     uint64              `json:"timestamp"`
	Text          string              `json:"text"`
	User          string              `json:"user"`
	Attachments   []Attachment        `json:"attachments"`
	Reacts        map[string][]string `json:"reacts"`
	SentToChannel bool                `json:"sent"`
}

// ParentMessage is returned from the API / to the front end
type ParentMessage struct {
	Timestamp   uint64              `json:"timestamp"`
	Text        string              `json:"text"`
	User        string              `json:"user"`
	Attachments []Attachment        `json:"attachments"`
	Reacts      map[string][]string `json:"reacts"`
	Thread      []ThreadMessage     `json:"thread"`
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
