package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const (
	Hostname = "https://api.meetup.com/"
)

var (
	Email    string
	Password string
	Token    string
	BotID    string
	APIKey   string
	GroupURL string
)

func init() {

	flag.StringVar(&Email, "e", "", "Account Email")
	flag.StringVar(&Password, "p", "", "Account Password")
	flag.StringVar(&Token, "t", "", "Account Token")
	flag.StringVar(&APIKey, "a", "", "Meetup API Key")
	flag.Parse()
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

	// If the message is "ping" reply with "Pong!"
	if strings.HasPrefix(m.Content, "!setgroup") {
		GroupURL = strings.TrimSpace(strings.TrimPrefix(m.Content, "!setgroup"))
		s.ChannelMessageSend(m.ChannelID, "Group url now set to: "+GroupURL)
	} else if strings.HasPrefix(m.Content, "!getevents") {
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
