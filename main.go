package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "cli" {
		if err := cli(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := server(); err != nil {
			log.Fatal(err)
		}
	}
}

func server() error {
	port := os.Getenv("PORT")
	if port == "" {
		return errors.New("$PORT must be set")
	}

	var httpServer http.Server

	http.HandleFunc("/feed", handler)
	log.Println("start http listening :" + port)
	httpServer.Addr = ":" + port
	log.Println(httpServer.ListenAndServe())

	return nil
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

func cli() error {
	c, err := NewCLI(os.Getenv("QIITA_TEAM_NAME"), os.Getenv("QIITA_ACCESS_TOKEN"))
	if err != nil {
		return err
	}

	feedItemNum, err := strconv.Atoi(os.Getenv("FEED_ITEM_NUM"))
	if err != nil || feedItemNum < 1 {
		return errors.New("$FEED_ITEM_NUM should be larger than zero")
	}

	return c.Run(uint(feedItemNum))
}
