package main

import (
	"bytes"
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

	"github.com/garyburd/redigo/redis"
	"golang.org/x/tools/blog/atom"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "cli" {
		if err := cli(); err != nil {
			log.Fatal(err)
		}
	}
}

func cli() error {
	qiitaItems, err := getQiitaItems()
	if err != nil {
		return err
	}

	atom, err := getAtom(qiitaItems)
	if err != nil {
		return err
	}

	return save(atom)
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
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return err
	}
	defer c.Close()

	name := "feed:" + os.Getenv("QIITA_TEAM_NAME")

	if _, err := c.Do("SET", name, content); err != nil {
		return err
	}

	// s, err := redis.String(c.Do("GET", name))
	// if err != nil {
	// 	return err
	// }
	// os.Stdout.Write([]byte(s))

	return nil
}
