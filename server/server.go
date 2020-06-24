package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"runtime"
	"slack-backer-upper/sqlite"

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
	}
	json.NewEncoder(res).Encode(channels)
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

	log.Fatalf("Error serving HTTP: %v\n", http.ListenAndServe(":8080", router))
}
