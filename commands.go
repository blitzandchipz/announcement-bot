package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/bwmarrin/discordgo"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

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
		urlName := strings.TrimSpace(strings.TrimPrefix(m.Content, "!setgroup"))
		url := hostname + "/" + urlName + "?key=" + APIKey
		resp, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting group: %s\n", err.Error())
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		defer resp.Body.Close()

		// meetup 404s on nonexistent group names
		if resp.StatusCode == 404 {
			// TODO pull error message from meetup's json response
			s.ChannelMessageSend(m.ChannelID, "Invalid group urlname")
			return
		}

		channel, err := getChannel(s, m.ChannelID)
		if err != nil {
			log.Printf("Error getting channel: %s\n", err.Error())
			return
		}

		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(channel.GuildID))
			err := b.Put([]byte("urlname"), []byte(urlName))
			return err
		})
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Group url now set to: %v\n", urlName))
		getNext(channel)
		printGuild(channel.GuildID)
	}

	// Gets a list of events for the currently set group
	// TODO: finish outputt message to server
	if strings.HasPrefix(m.Content, "!getevents") {
		channel, err := getChannel(s, m.ChannelID)
		if err != nil {
			log.Printf("Error getting channel: %s\n", err.Error())
			return
		}

		urlName, err := getURLName(channel.GuildID)
		if err != nil {
			log.Printf("Error getting urlname from GuildID: %s\n", err.Error())
			return
		}

		if urlName == "" {
			s.ChannelMessageSend(m.ChannelID, "Run !setgroup first")
			return
		}
		url := hostname + urlName + "/events?key=" + APIKey + "&page=25"
		r, err := http.Get(url)
		if err != nil {
			log.Printf("Error getting events: %s\n", err.Error())
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}
		defer r.Body.Close()
		contents, _ := ioutil.ReadAll(r.Body)
		fmt.Print(string(contents))
		s.ChannelMessageSend(m.ChannelID, string(contents))
	}

	// Returns the next upcoming, public event
	if strings.HasPrefix(m.Content, "!nextevent") {
		channel, err := getChannel(s, m.ChannelID)
		if err != nil {
			log.Printf("Error getting channel: %s\n", err.Error())
			return
		}

		urlName, err := getURLName(channel.GuildID)
		if err != nil {
			log.Printf("Error getting urlname: %s\n", err.Error())
			return
		}

		if urlName == "" {
			s.ChannelMessageSend(m.ChannelID, "Run !setgroup first")
			return
		}
		events := getNext(channel)
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

func getNext(channel *discordgo.Channel) []Event {
	var events []Event

	urlName, err := getURLName(channel.GuildID)
	if err != nil {
		log.Printf("Error getting urlName: %s\n", err.Error())
	}

	url := hostname + urlName + "/events?key=" + APIKey + "&page=1"
	err = getJSON(url, &events)
	if err != nil {
		log.Printf("Error getJSON: %s\n", err.Error())
	}

	if len(events) > 0 {
		event, _ := json.Marshal(events[0])
		db.Update(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(channel.GuildID))
			err := b.Put([]byte("nextevent"), event)
			return err
		})
	}

	return events
}
