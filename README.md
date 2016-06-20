# Commands
 * `!setgroup` : This command sets the meetup group for the server. **Must be ran before most commands**
 * `!nextevent` : Gets the next upcoming event for the set group and prints it in chat  

# Instructions
## Run your own bot
1. Clone this repository  
2. Three methods of config below:
  1. Create `config.json` file from the example
    1. Enter your meetup.com API key for `APIKey` value
    2. Enter your bot's Discord token OR email and password for `Token` or `Email` and `Password`
  2. Environment variables
     1. Create environment variables for the same values as above
  3. Command line arguments (these take precident over files and env vars)
3. Run `go install`  
4. Run `meetup-bot`  
5. [Add your bot to your server](https://discordapp.com/developers/docs/topics/oauth2#adding-bots-to-guilds)

## [Add live bot](https://discordapp.com/oauth2/authorize?client_id=184056719863709706&scope=bot&permissions=0)
