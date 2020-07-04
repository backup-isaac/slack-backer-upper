package server

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"slack-backer-upper/slack"
	"strconv"
	"time"
)

func defaultPage(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/static/index.html", http.StatusFound)
}

func (s *Server) listChannels(res http.ResponseWriter, req *http.Request) {
	channels, err := s.storage.GetChannels()
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

func (s *Server) queryMessages(channel string, from, to time.Time) ([]slack.ParentMessage, error) {
	parents, err := s.storage.GetParentMessages(channel, from, to)
	if err != nil {
		return nil, err
	}
	messages := make([]slack.ParentMessage, len(parents))
	for i, p := range parents {
		messages[i], err = slack.ParentMessageFromStored(p)
		if err != nil {
			return nil, err
		}
		replies, err := s.storage.GetThreadReplies(channel, p.Timestamp)
		if err != nil {
			return nil, err
		}
		messages[i].Thread = replies
	}
	return messages, nil
}

func (s *Server) getMessages(res http.ResponseWriter, req *http.Request) {
	channel, fromMillis, toMillis, err := parseGetMessageParams(req.URL.Query())
	if err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	from := time.Unix(0, fromMillis*1e6)
	to := time.Unix(0, toMillis*1e6)
	messages, err := s.queryMessages(channel, from, to)
	if err != nil {
		http.Error(res, fmt.Sprintf("Error getting messages: %v", err), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(res).Encode(messages)
}

func (s *Server) uploadZip(res http.ResponseWriter, req *http.Request) {
	if err := req.ParseMultipartForm(1048576); err != nil {
		http.Error(res, fmt.Sprintf("Error parsing multipart form: %v", err), http.StatusBadRequest)
		return
	}
	for _, fh := range req.MultipartForm.File {
		for _, f := range fh {
			file, err := f.Open()
			defer file.Close()
			if err != nil {
				http.Error(res, fmt.Sprintf("Error parsing multipart form: %v", err), http.StatusBadRequest)
				return
			}
			z, err := zip.NewReader(file, f.Size)
			if err != nil {
				http.Error(res, fmt.Sprintf("Error parsing multipart form: %v", err), http.StatusBadRequest)
				return
			}
			if err = s.archiver.ImportZip(*z); err != nil {
				http.Error(res, fmt.Sprintf("Error importing zip: %v", err), http.StatusBadRequest)
				return
			}
		}
	}
	res.Write([]byte{})
}
