package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

// ClientParams is parameters for Client
type ClientParams struct {
	teamName    string // Qiita:Team name
	accessToken string // Qiita:Team access token
	baseURL     string
}

// A Client manages communication with the Qiita API.
type Client struct {
	client *http.Client // HTTP client used to communicate with the API.
	ClientParams
}

// NewClient returns a new Qiita:Team API client.
func NewClient(p ClientParams) *Client {
	c := &Client{}
	c.client = &http.Client{Timeout: time.Duration(5) * time.Second}
	c.teamName = p.teamName
	c.accessToken = p.accessToken

	c.baseURL = "https://" + c.teamName + ".qiita.com"
	if p.baseURL != "" {
		c.baseURL = p.baseURL
	}

	return c
}

// ListItems returns Qiita:Team items
// See https://qiita.com/api/v2/docs#get-apiv2items
func (c *Client) ListItems(number string) ([]QiitaItem, error) {
	endPoint := c.baseURL + "/api/v2/items"
	req, err := http.NewRequest("GET", endPoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+c.accessToken)

	if number != "" {
		q := req.URL.Query()
		q.Add("per_page", number)
		req.URL.RawQuery = q.Encode()
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	return extractQiitaItems(resp.Body)
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
