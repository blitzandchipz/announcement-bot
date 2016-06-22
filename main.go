package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

// hostname for meetup.com's api
const hostname = "https://api.meetup.com/"

var (
	// Email for discord user account
	Email string
	// Password for discord user account
	Password string
	// Token for discord bot account
	Token string
	// BotID of the user account
	BotID string
	// APIKey for meetup.com
	APIKey string
	// config for the bots settings
	config *Config
	// db for dynamic settings per guild
	db *bolt.DB
)

// Config stores the settings for the bot
type Config struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
	APIKey   string `json:"apikey"`
}

// Event is a single event from meetup.com
type Event struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	Time          int64  `json:"time"`
	Updated       int64  `json:"updated"`
	UTCOffset     int64  `json:"utc_offset"`
	WaitlistCount int    `json:"waitlist_count"`
	YesRSVPCount  int    `json:"yes_rsvp_count"`
	Venue         Venue  `json:"venue"`
	Link          string `json:"link"`
	Description   string `json:"description"`
	Visibility    string `json:"visibility"`
}

// Venue is a location where events happen
type Venue struct {
	ID                   int     `json:"id"`
	Name                 string  `json:"name"`
	Lat                  float64 `json:"lat"`
	Lon                  float64 `json:"lon"`
	Repinned             bool    `json:"repinned"`
	Address1             string  `json:"address_1"`
	Address2             string  `json:"address_2"`
	City                 string  `json:"city"`
	Country              string  `json:"country"`
	LocalizedCountryName string  `json:"localized_country_name"`
	Zip                  string  `json:"zip"`
	State                string  `json:"state"`
}

func init() {
	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.StringVar(&APIKey, "a", "", "Meetup API Key")
	flag.Parse()

	path, err := filepath.Abs("config.json")
	if err == nil {
		configFile := []byte("")
		configFile, err = ioutil.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(configFile, config)
		}
	}
	if err != nil {
		log.Printf("Error opening config file: %s\n", err.Error())
	}

	if config != nil {
		if Email == "" && config.Email != "" {
			Email = config.Email
		}

		if Password == "" && config.Password != "" {
			Password = config.Password
		}

		if Token == "" && config.Token != "" {
			Token = config.Token
		}

		if APIKey == "" && config.APIKey != "" {
			APIKey = config.APIKey
		}
	} else {
		if Email == "" && os.Getenv("Email") != "" {
			Email = os.Getenv("Email")
		}

		if Password == "" && os.Getenv("Password") != "" {
			Password = os.Getenv("Password")
		}

		if Token == "" && os.Getenv("Token") != "" {
			Token = os.Getenv("Token")
		}

		if APIKey == "" && os.Getenv("APIKey") != "" {
			APIKey = os.Getenv("APIKey")
		}
	}
}

func main() {
	// Open database
	db, err := bolt.Open("settings.db", 0600, nil)
	if err != nil {
		log.Printf("Error opening bolt db: %s\n", err.Error())
	}
	defer db.Close()

	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		log.Printf("Error creating Discord session: %s\n", err.Error())
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		log.Printf("Error obtaining account details: %s\n", err.Error())
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Get all the guilds the bot is in
	guilds, err := dg.UserGuilds()
	if err != nil {
		log.Printf("Error getting guilds: %s\n", err.Error())
	}

	// Make sure a bucket exists for each guild
	db.Update(func(tx *bolt.Tx) error {
		for _, guild := range guilds {
			_, err := tx.CreateBucket([]byte(guild.ID))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}
		}
		return nil
	})

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	dg.Open()

	fmt.Println("Meetup Bot is now running.  Press CTRL-C to exit.")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	s := <-c
	fmt.Println("Got signal:", s)
	return
}
