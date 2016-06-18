package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"net/http"
	"time"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func errMsg(err error) string {
	msg := err.Error()
	return msg
}

func printGuild(guildID string) {
	DB.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(guildID))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
}

// Helper function to unmarshle json data into structs
func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

// Helper function to convert ms since epoch to ANSIC time format
func msToTime(ms int64) (string, error) {
	return time.Unix(0, ms*int64(time.Millisecond)).Format(time.ANSIC), nil
}

// Helper function to truncates a string and adds ellipsis
func truncate(str string) string {
	if len(str) > 150 {
		return string(str[:47] + "...")
	}
	return str
}
