package main

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient(ClientParams{
		teamName:    "example",
		accessToken: "testToken",
	})

	if got, want := c.baseURL, "https://example.qiita.com"; got != want {
		t.Errorf("NewClient baseURL is %v, want %v", got, want)
	}
}

func TestClientListItems(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestExtractQiitaItems(t *testing.T) {
	t.Skip("This test is pending.")
}
