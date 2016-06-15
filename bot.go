package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"stathat.com/c/jconfig"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == BotID {
		return
	}

	// Sets the meetup group needed for future commands
	// TODO Add permissions: only admins should be able to set the group for
	// the server
	if strings.HasPrefix(m.Content, "!setgroup") {
		tempURL := strings.TrimSpace(strings.TrimPrefix(m.Content, "!setgroup"))
		url := Hostname + "/" + tempURL + "?key=" + APIKey
		resp, err := http.Get(url)
		if err != nil {
			errMsg := errMsg(err)
			s.ChannelMessageSend(m.ChannelID, errMsg)
			return
		}
		defer resp.Body.Close()

		// meetup 404s on nonexistent group names
		if resp.StatusCode == 404 {
			// TODO pull error message from meetup's json response
			s.ChannelMessageSend(m.ChannelID, "Invalid group urlname")
			return
		}
		// Store the group in global
		// TODO rework this so it works per discord guild
		GroupURL = tempURL
		s.ChannelMessageSend(m.ChannelID, "Group url now set to: "+GroupURL)
	}

	// Gets a list of events for the currently set group
	// TODO: finish outputt message to server
	if strings.HasPrefix(m.Content, "!getevents") {
		if GroupURL == "" {
			s.ChannelMessageSend(m.ChannelID, "Run !setgroup first")
			return
		}
		url := Hostname + GroupURL + "/events?key=" + APIKey + "&page=25"
		r, err := http.Get(url)
		if err != nil {
			errMsg := errMsg(err)
			s.ChannelMessageSend(m.ChannelID, errMsg)
			return
		}
		defer r.Body.Close()
		contents, _ := ioutil.ReadAll(r.Body)
		fmt.Print(string(contents))
		s.ChannelMessageSend(m.ChannelID, string(contents))
	}

	// Returns the next upcoming, public event
	if strings.HasPrefix(m.Content, "!nextevent") {
		if GroupURL == "" {
			s.ChannelMessageSend(m.ChannelID, "Run !setgroup first")
			return
		}
		url := Hostname + GroupURL + "/events?key=" + APIKey + "&page=1"
		var events []Event
		err := getJSON(url, &events)
		if err != nil {
			errMsg := errMsg(err)
			s.ChannelMessageSend(m.ChannelID, errMsg)
			return
		}
		msg := "No future, public events found"

		// Check if theres any events
		if len(events) > 0 {
			event := events[0]
			// Only consider public and upcoming events
			if (event.Visibility == "public") && (event.Status == "upcoming") {
				venueStr := ""
				// Check if a venue exists
				if event.Venue.Name != "" {
					venue := event.Venue
					// Just print the name if there's no address
					// TODO Print address even if there's no name
					if venue.Address1 == "" {
						venueStr = fmt.Sprintf("\nAt: `%v`", venue.Name)
					} else {
						// Print full location details
						// TODO test for missing location information
						venueStr = fmt.Sprintf("\nAt: `%v` - %v %v, %v %v",
							venue.Name, venue.Address1, venue.City, venue.State, venue.Zip)
					}
				}
				time, _ := msToTime(event.Time)
				// description := truncate(event.Description)
				msg = fmt.Sprintf("Next event: `%v` - %v%v\n%v",
					event.Name, time, venueStr, event.Link)
			}
		}

		s.ChannelMessageSend(m.ChannelID, msg)
	}
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
