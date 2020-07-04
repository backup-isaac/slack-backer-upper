package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path"
	"runtime"
	"slack-backer-upper/slack"
	"slack-backer-upper/storage"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// archiveViewer encapsulates the archive viewing API
type archiveViewer struct {
	storage storage.ViewerStorage
}

func defaultPage(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/static/index.html", http.StatusFound)
}

func (a *archiveViewer) listChannels(res http.ResponseWriter, req *http.Request) {
	channels, err := a.storage.ListChannels()
	if err != nil {
		http.Error(res, fmt.Sprintf("Error listing channels: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(channels)
}

func parseGetMessageParams(query url.Values) (string, int64, int64, error) {
	channel := query.Get("channel")
	if channel == "" {
		return "", 0, 0, fmt.Errorf("Missing channel")
	}
	fromStr := query.Get("from")
	if fromStr == "" {
		return "", 0, 0, fmt.Errorf("Missing from")
	}
	fromMillis, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("Invalid from: %v", err)
	}
	toStr := query.Get("to")
	if toStr == "" {
		return "", 0, 0, fmt.Errorf("Missing to")
	}
	toMillis, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		return "", 0, 0, fmt.Errorf("Invalid to: %v", err)
	}
	return channel, fromMillis, toMillis, nil
}

func (a *archiveViewer) queryMessages(channel string, from, to time.Time) ([]slack.ParentMessage, error) {
	parents, err := a.storage.GetParentMessages(channel, from, to)
	if err != nil {
		return nil, err
	}
	messages := make([]slack.ParentMessage, len(parents))
	for i, p := range parents {
		messages[i], err = slack.ParentMessageFromStored(p)
		if err != nil {
			return nil, err
		}
		replies, err := a.storage.GetThreadReplies(channel, p.Timestamp)
		if err != nil {
			return nil, err
		}
		messages[i].Thread = replies
	}
	return messages, nil
}

func (a *archiveViewer) getMessages(res http.ResponseWriter, req *http.Request) {
	channel, fromMillis, toMillis, err := parseGetMessageParams(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	from := time.Unix(0, fromMillis*1e6)
	to := time.Unix(0, toMillis*1e6)
	messages, err := a.queryMessages(channel, from, to)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error getting messages: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(messages)
}

// Start registers API routes then starts the HTTP server
func Start() error {
	router := mux.NewRouter()
	log.Println("Starting HTTP server on :8080...")

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("Could not find runtime caller")
	}
	router.PathPrefix("/static/").Handler(http.FileServer(http.Dir(path.Dir(filename))))
	router.HandleFunc("/", defaultPage)

	storage, err := storage.NewViewerStorage()
	if err != nil {
		return fmt.Errorf("Error initializing server: %v", err)
	}
	defer storage.Close()

	a := archiveViewer{
		storage: storage,
	}
	router.HandleFunc("/channels", a.listChannels).Methods("GET")
	router.HandleFunc("/messages", a.getMessages).Methods("GET")

	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt)

	serveResult := make(chan error)

	go func() {
		serveResult <- http.ListenAndServe(":8080", router)
	}()

	select {
	case err = <-serveResult:
		return fmt.Errorf("Error serving HTTP: %v", err)
	case s := <-sigChannel:
		return fmt.Errorf("Received signal %v", s)
	}
}
