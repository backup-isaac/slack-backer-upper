package server

import (
	"archive/zip"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"slack-backer-upper/slack"
	"time"

	"github.com/gorilla/mux"
)

const networkInterface = ":8080"

type serverStorage interface {
	GetChannels() ([]string, error)
	GetParentMessages(channelName string, from, to time.Time) ([]slack.StoredMessage, error)
	GetThreadReplies(channelName, timestamp string) ([]slack.ThreadMessage, error)
}

type serverArchiver interface {
	ImportZip(zip.Reader) error
}

// Server serves APIs from the archive
type Server struct {
	archiver serverArchiver
	storage  serverStorage
}

// New creates a new Server with the provided Archiver and storage
func New(a serverArchiver, s serverStorage) Server {
	return Server{
		archiver: a,
		storage:  s,
	}
}

// Start registers API routes then starts the HTTP server
func (s *Server) Start() error {
	router := mux.NewRouter()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("Could not find runtime caller")
	}
	router.PathPrefix("/static/").Handler(http.FileServer(http.Dir(path.Dir(filename))))
	router.HandleFunc("/", defaultPage)
	router.HandleFunc("/channels", s.listChannels).Methods("GET")
	router.HandleFunc("/messages", s.getMessages).Methods("GET")
	router.HandleFunc("/upload", s.uploadZip).Methods("POST")

	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, os.Interrupt)

	serveResult := make(chan error)

	log.Printf("Starting HTTP server on %s...", networkInterface)
	go func() {
		serveResult <- http.ListenAndServe(networkInterface, router)
	}()

	select {
	case err := <-serveResult:
		return fmt.Errorf("Error serving HTTP: %v", err)
	case s := <-sigChannel:
		return fmt.Errorf("Received signal %v", s)
	}
}
