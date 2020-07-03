package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"slack-backer-upper/sqlite"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// archiveViewer encapsulates the archive viewing API
type archiveViewer struct {
	storage sqlite.ViewerStorage
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

func (a *archiveViewer) getMessages(res http.ResponseWriter, req *http.Request) {
	channel := req.URL.Query().Get("channel")
	if channel == "" {
		http.Error(res, fmt.Sprintf("Missing channel"), http.StatusBadRequest)
		return
	}
	fromStr := req.URL.Query().Get("from")
	if fromStr == "" {
		http.Error(res, fmt.Sprintf("Missing from"), http.StatusBadRequest)
		return
	}
	fromMillis, err := strconv.ParseInt(fromStr, 10, 64)
	if err != nil {
		http.Error(res, fmt.Sprintf("Invalid from"), http.StatusBadRequest)
		return
	}
	from := time.Unix(0, fromMillis*1e6)
	toStr := req.URL.Query().Get("to")
	if toStr == "" {
		http.Error(res, fmt.Sprintf("Missing to"), http.StatusBadRequest)
		return
	}
	toMillis, err := strconv.ParseInt(toStr, 10, 64)
	if err != nil {
		http.Error(res, fmt.Sprintf("Invalid to"), http.StatusBadRequest)
		return
	}
	to := time.Unix(0, toMillis*1e6)
	messages, err := a.storage.GetMessages(channel, from, to)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error getting messages: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(messages)
}

// Start registers API routes then starts the HTTP server
func Start() {
	router := mux.NewRouter()
	log.Println("Starting HTTP server on :8080...")

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("Could not find runtime caller")
	}
	router.PathPrefix("/static/").Handler(http.FileServer(http.Dir(path.Dir(filename))))
	router.HandleFunc("/", defaultPage)

	storage, err := sqlite.NewViewerStorage()
	if err != nil {
		log.Fatalf("Error initializing server: %v\n", err)
	}
	defer storage.Close()
	a := archiveViewer{
		storage: storage,
	}
	router.HandleFunc("/channels", a.listChannels).Methods("GET")
	router.HandleFunc("/messages", a.getMessages).Methods("GET")

	log.Fatalf("Error serving HTTP: %v\n", http.ListenAndServe(":8080", router))
}
