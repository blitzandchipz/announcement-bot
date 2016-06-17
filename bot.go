package main

import (
	"encoding/json"
	"flag"
	"fmt"
	// "github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"net/http"
	"stathat.com/c/jconfig"
	"time"
)

// Hostname for meetup.com's api
const Hostname = "https://api.meetup.com/"

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
	// GroupURL for a meetup group
	GroupURL string
	// Config for the bots settings
	Config *jconfig.Config
)

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

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func errMsg(err error) string {
	msg := err.Error()
	return msg
}

func init() {
	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.StringVar(&APIKey, "a", "", "Meetup API Key")
	flag.Parse()

	Config = jconfig.LoadConfig("config.json")

	if Config != nil {
		if Email == "" && Config.GetString("Email") != "" {
			Email = Config.GetString("Email")
		}

		if Password == "" && Config.GetString("Password") != "" {
			Password = Config.GetString("Password")
		}

		if Token == "" && Config.GetString("Token") != "" {
			Token = Config.GetString("Token")
		}

		if APIKey == "" && Config.GetString("APIKey") != "" {
			APIKey = Config.GetString("APIKey")
		}
	}
}

func main() {
	// Open database
	// db, err := bolt.Open("settings.db", 0600, nil)
	// Create a new Discord session using the provided login information.
	dg, err := discordgo.New(Email, Password, Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Get the account information.
	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	// Store the account ID for later use.
	BotID = u.ID

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(messageCreate)

	// Open the websocket and begin listening.
	dg.Open()

	fmt.Println("Meetup Bot is now running.  Press CTRL-C to exit.")
	// Simple way to keep program running until CTRL-C is pressed.
	<-make(chan struct{})
	return
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

// Helper function to truncates a string and adds ellipsis
func truncate(str string) string {
	if len(str) > 150 {
		return string(str[:47] + "...")
	}
	return str
}

// Helper function to convert ms since epoch to ANSIC time format
func msToTime(ms int64) (string, error) {
	return time.Unix(0, ms*int64(time.Millisecond)).Format(time.ANSIC), nil
}
