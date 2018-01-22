package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
)

// Server is a qiita-team-feed Server
type Server struct {
	port string
}

// NewServer returns a new Server
func NewServer(port string) (*Server, error) {
	if port == "" {
		return nil, errors.New("$PORT must be set")
	}

	return &Server{port: port}, nil
}

// Run saves Qiita:Team feed to Redis
func (s *Server) Run() {
	var httpServer http.Server

	http.HandleFunc("/feed", handler)
	log.Println("start http listening :" + s.port)
	httpServer.Addr = ":" + s.port
	log.Println(httpServer.ListenAndServe())
}

func handler(w http.ResponseWriter, r *http.Request) {
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(dump))

	// Todo: http.StatusForbidden も返すようにする
	feed, err := getFeed(r.URL.Query().Get("user"), r.URL.Query().Get("token"))
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, feed)
}

func getFeed(user, token string) (string, error) {
	if user == "" {
		return "", errors.New("user is required")
	}

	if token == "" {
		return "", errors.New("token is required")
	}

	db, err := NewDB()
	if err != nil {
		return "", err
	}
	defer db.Close()

	if t, err := db.GetToken(user); err != nil || t != token {
		return "", errors.New("Invalid user token")
	}

	s, err := db.GetFeed()
	if err != nil || s == "" {
		return "", errors.New("Failure to get feed")
	}

	return s, nil
}
