package main

import (
	"os"

	"github.com/garyburd/redigo/redis"
)

// DB behaves a specific database server.
type DB interface {
	// Close closes the connection.
	Close() error

	// GetToken gets a token by user
	GetToken(user string) (string, error)

	// GetFeed gets a feed of this Qiita:Item
	GetFeed() (string, error)

	// SetFeed sets a feed of this Qiita:Item
	SetFeed(content []byte) error
}

type db struct {
	conn redis.Conn
}

// NewDB returns a new DB connection
func NewDB() (DB, error) {
	c, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		return nil, err
	}

	return &db{conn: c}, nil
}

func (d db) Close() error {
	return d.conn.Close()
}

func (d db) GetToken(user string) (string, error) {
	return redis.String(d.conn.Do("GET", "user:"+user))
}

func (d db) GetFeed() (string, error) {
	return redis.String(d.conn.Do("GET", getFeedKey()))
}

func (d db) SetFeed(content []byte) error {
	_, err := d.conn.Do("SET", getFeedKey(), content)
	return err
}

func getFeedKey() string {
	return "feed:" + os.Getenv("QIITA_TEAM_NAME")
}
