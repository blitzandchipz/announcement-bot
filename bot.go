package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"stathat.com/c/jconfig"
	"strings"

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

func check(e error) {
	if e != nil {
		panic(e)
	}
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
	if strings.HasPrefix(m.Content, "!setgroup") {
		GroupURL = strings.TrimSpace(strings.TrimPrefix(m.Content, "!setgroup"))
		s.ChannelMessageSend(m.ChannelID, "Group url now set to: "+GroupURL)
	}

	// Gets a list of events for the currently set group
	if strings.HasPrefix(m.Content, "!getevents") {
		if GroupURL == "" {
			s.ChannelMessageSend(m.ChannelID, "Run !setgroup first")
			return
		}
		url := Hostname + GroupURL + "/events?key=" + APIKey + "?page=3"
		events, err := http.Get(url)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		contents, _ := ioutil.ReadAll(events.Body)
		fmt.Print(string(contents))
		s.ChannelMessageSend(m.ChannelID, string(contents))
	}
}
