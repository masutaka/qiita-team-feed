package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/tools/blog/atom"
)

func main() {
	qiitaItems, err := getQiitaItems()
	if err != nil {
		log.Fatal(err)
	}

	atom, err := getAtom(qiitaItems)
	if err != nil {
		log.Fatal(err)
	}

	if err = save(atom); err != nil {
		log.Fatal(err)
	}
}

func getQiitaItems() ([]QiitaItem, error) {
	client := &http.Client{Timeout: time.Duration(5) * time.Second}

	req, err := http.NewRequest("GET", qiitaEndPoint(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", qiitaAuthorization())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	return extractQiitaItems(resp.Body)
}

func qiitaEndPoint() string {
	return "https://" + os.Getenv("QIITA_TEAM_NAME") + ".qiita.com/api/v2/items"
}

func qiitaAuthorization() string {
	return "Bearer " + os.Getenv("QIITA_ACCESS_TOKEN")
}

// QiitaItem is a Qiita Item
type QiitaItem struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	User      QiitaUser `json:"user"`
}

// QiitaUser is a Qiita User
type QiitaUser struct {
	ID              string `json:"id"`
	ProfileImageURL string `json:"profile_image_url"`
}

func extractQiitaItems(respBody io.Reader) ([]QiitaItem, error) {
	qiitaItems := make([]QiitaItem, 20)

	decoder := json.NewDecoder(respBody)
	err := decoder.Decode(&qiitaItems)

	return qiitaItems, err
}

func getAtom(qiitaItems []QiitaItem) ([]byte, error) {
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
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		return err
	}
	defer db.Close()

	name := os.Getenv("QIITA_TEAM_NAME")
	err = find(db, name)

	switch {
	case err == sql.ErrNoRows:
		if err := create(db, name, content); err != nil {
			return err
		}
	case err != nil:
		return err
	default:
		if err := update(db, name, content); err != nil {
			return err
		}
	}

	// Debug
	var dbContent []byte
	err = db.QueryRow(
		"SELECT content FROM feeds WHERE name = $1 LIMIT 1", name,
	).Scan(&dbContent)
	if err != nil {
		return err
	}
	os.Stdout.Write(dbContent)

	return nil
}

func find(db *sql.DB, name string) error {
	var dummy string

	return db.QueryRow(
		"SELECT name FROM feeds WHERE name = $1 LIMIT 1", name,
	).Scan(&dummy)
}

func create(db *sql.DB, name string, content []byte) error {
	_, err := db.Exec(
		`INSERT INTO feeds (name, content, created_at, updated_at)
         VALUES($1, $2, NOW(), NOW())`,
		name, content,
	)

	return err
}

func update(db *sql.DB, name string, content []byte) error {
	_, err := db.Exec(
		"UPDATE feeds SET content = $1, updated_at = NOW() WHERE name = $2",
		content, name,
	)

	return err
}
