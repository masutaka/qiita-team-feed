package main

import (
	"os"
	"testing"
)

func TestNewDB(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestDBClose(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestDBGetToken(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestDBGetFeed(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestDBSetFeed(t *testing.T) {
	t.Skip("This test is pending.")
}

func TestDBGetFeedKey(t *testing.T) {
	os.Setenv("QIITA_TEAM_NAME", "example")
	defer os.Unsetenv("QIITA_TEAM_NAME")

	actual := getFeedKey()

	const expected = "feed:example"
	if actual != expected {
		t.Errorf("expected %s but got %s", expected, actual)
	}
}
