package main

import (
	"errors"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "cli" {
		c, err := NewCLI(
			os.Getenv("QIITA_TEAM_NAME"),
			os.Getenv("QIITA_ACCESS_TOKEN"),
		)
		if err != nil {
			log.Fatal(err)
		}

		// The default of per_page is 20.
		// See https://qiita.com/api/v2/docs
		num := 20
		if numStr := os.Getenv("FEED_ITEM_NUM"); numStr != "" {
			num, err = strconv.Atoi(numStr)
			if err != nil {
				log.Fatal(err)
			}
		}

		if num < 1 {
			log.Fatal(errors.New("$FEED_ITEM_NUM should be larger than zero"))
		}

		if err := c.Run(uint(num)); err != nil {
			log.Fatal(err)
		}
	} else {
		s, err := NewServer(os.Getenv("PORT"))
		if err != nil {
			log.Fatal(err)
		}

		s.Run()
	}
}
