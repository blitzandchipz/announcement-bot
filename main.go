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
)

// hostname for meetup.com's api
const hostname = "https://api.meetup.com/"

var (
	// BotID of the user account
	BotID string
	// config for the bots settings
	config *Config
	// db for dynamic settings per guild
	db *bolt.DB
)

// Config stores the settings for the bot
type Config struct {
	APIKey   string `json:"apikey"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Token    string `json:"token"`
}

// Validate the config settings to ensure essential parameters are set
func (cfg Config) Validate() error {
	if config.APIKey == "" {
		return fmt.Errorf("Missing Meetup APIKey")
	}
	if config.Token == "" {
		if config.Email == "" || config.Password == "" {
			return fmt.Errorf("Missing Discord Token or Email and Password")
		}
	}
	return nil
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
	config = &Config{}

	path := "./config.json"
	if _, err := os.Stat(path); err == nil {
		configFile, err := ioutil.ReadFile(path)
		if err == nil {
			err = json.Unmarshal(configFile, &config)
		}
		if err != nil {
			log.Fatalf("Error opening config file: %s\n", err.Error())
		}
	}

	flag.StringVar(&config.APIKey, "a", config.APIKey, "Meetup API Key")
	flag.StringVar(&config.Email, "e", config.Email, "Account Email")
	flag.StringVar(&config.Password, "p", config.Password, "Account Password")
	flag.StringVar(&config.Token, "t", config.Token, "Account Token")
	flag.Parse()

	if APIKey := os.Getenv("APIKey"); APIKey != "" {
		config.APIKey = APIKey
	}

	if Email := os.Getenv("Email"); Email != "" {
		config.Email = Email
	}

	if Password := os.Getenv("Password"); Password != "" {
		config.Password = Password
	}

	if Token := os.Getenv("Token"); Token != "" {
		config.Token = Token
	}

	err := config.Validate()
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	// Open database
	var err error
	db, err = bolt.Open("settings.db", 0600, nil)
	if err != nil {
		log.Fatalf("Error opening bolt db: %s\n", err.Error())
	}
	defer db.Close()

	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(config.Email, config.Password, config.Token)
	if err != nil {
		log.Fatalf("Error creating Discord session: %s\n", err.Error())
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		log.Fatalf("Error obtaining account details: %s\n", err.Error())
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Get all the guilds the bot is in
	guilds, err := dg.UserGuilds()
	if err != nil {
		log.Fatalf("Error getting guilds: %s\n", err.Error())
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
	// Close the websocket
	dg.Close()
	fmt.Println("Got signal:", s)
	return
}
