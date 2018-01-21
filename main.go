package main

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"golang.org/x/tools/blog/atom"
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
	c := NewClient(ClientParams{
		teamName:    os.Getenv("QIITA_TEAM_NAME"),
		accessToken: os.Getenv("QIITA_ACCESS_TOKEN"),
	})

	qiitaItems, err := c.ListItems(os.Getenv("FEED_ITEM_NUM"))
	if err != nil {
		return err
	}

	atom, err := generateAtom(qiitaItems)
	if err != nil {
		return err
	}

	return save(atom)
}

func generateAtom(qiitaItems []QiitaItem) ([]byte, error) {
	team := os.Getenv("QIITA_TEAM_NAME")

	links := []atom.Link{
		atom.Link{Href: "https://" + team + ".qiita.com"},
	}

	author := &atom.Person{
		Name: team,
	}

	entries := []*atom.Entry{}

	for _, item := range qiitaItems {
		entries = append(entries, &atom.Entry{
			Title:     item.Title,
			ID:        item.ID,
			Link:      []atom.Link{atom.Link{Href: item.URL}},
			Published: atom.Time(item.CreatedAt),
			Updated:   atom.Time(item.UpdatedAt),
			Author: &atom.Person{
				Name: item.User.ID,
				URI:  "https://" + team + ".qiita.com/" + item.User.ID + "/items",
			},
			Content: &atom.Text{
				Type: "html",
				Body: generateContent(item.User),
			},
		})
	}

	feed := atom.Feed{
		Title:   team + " Qiita:Team",
		ID:      "https://" + team + ".qiita.com",
		Link:    links,
		Author:  author,
		Updated: atom.Time(time.Now()),
		Entry:   entries,
	}

	xmlBody, err := xml.Marshal(feed)
	if err != nil {
		return nil, err
	}

	return append([]byte(strings.TrimSpace(xml.Header)), xmlBody...), nil
}

func generateContent(user QiitaUser) string {
	m := map[string]interface{}{
		"userID":      user.ID,
		"userIconURL": user.ProfileImageURL,
	}
	t := template.Must(template.New("").Parse(
		`<a href="/{{.userID}}/items" rel="noreferrer">
<img alt="@{{.userID}}" width="32" height="32" src="{{.userIconURL}}">
</a>
`))

	var rendered bytes.Buffer
	t.Execute(&rendered, m)

	return rendered.String()
}

func save(content []byte) error {
	db, err := NewDB()
	if err != nil {
		return err
	}
	defer db.Close()

	return db.SetFeed(content)
}
