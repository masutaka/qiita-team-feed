package main

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"html/template"
	"strings"
	"time"

	"github.com/ktsujichan/qiita-sdk-go/qiita"
	"golang.org/x/tools/blog/atom"
)

// CLI is a qiita-team-feed CLI
type CLI struct {
	client   *qiita.Client
	teamName string // Qiita:Team name
}

// NewCLI returns a new CLI
func NewCLI(teamName, accessToken string) (*CLI, error) {
	config := qiita.NewConfig()
	config.WithEndpoint(fmt.Sprintf("https://%s.qiita.com", teamName))

	c, err := qiita.NewClient(accessToken, *config)
	if err != nil {
		return nil, err
	}

	return &CLI{client: c, teamName: teamName}, nil
}

// Run saves Qiita:Team feed to Redis
func (c *CLI) Run(feedItemNum uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	qiitaItems, err := c.client.ListItems(ctx, 1, feedItemNum, "*")
	if err != nil {
		return err
	}

	atom, err := generateAtom(c.teamName, *qiitaItems)
	if err != nil {
		return err
	}

	return save(atom)
}

func generateAtom(teamName string, qiitaItems qiita.Items) ([]byte, error) {
	links := []atom.Link{
		atom.Link{Href: "https://" + teamName + ".qiita.com"},
	}

	author := &atom.Person{
		Name: teamName,
	}

	entries := []*atom.Entry{}

	for _, item := range qiitaItems {
		createdAt, err := stringToTime(item.CreatedAt)
		if err != nil {
			return nil, err
		}

		updatedAt, err := stringToTime(item.UpdatedAt)
		if err != nil {
			return nil, err
		}

		entries = append(entries, &atom.Entry{
			Title:     item.Title,
			ID:        item.Id,
			Link:      []atom.Link{atom.Link{Href: item.Url}},
			Published: createdAt,
			Updated:   updatedAt,
			Author: &atom.Person{
				Name: item.User.Id,
				URI:  "https://" + teamName + ".qiita.com/" + item.User.Id + "/items",
			},
			Content: &atom.Text{
				Type: "html",
				Body: generateContent(*item.User),
			},
		})
	}

	feed := atom.Feed{
		Title:   teamName + " Qiita:Team",
		ID:      "https://" + teamName + ".qiita.com",
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

func stringToTime(str string) (atom.TimeStr, error) {
	t, err := time.Parse(time.RFC3339, str)
	return atom.Time(t), err
}

func generateContent(user qiita.User) string {
	m := map[string]interface{}{
		"userID":      user.Id,
		"userIconURL": user.ProfileImageUrl,
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
