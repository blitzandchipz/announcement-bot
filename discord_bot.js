/*
	this bot is a ping pong bot, and every time a message
	beginning with "ping" is sent, it will reply with
	"pong!".
*/

var Discord = require("discord.js");

// Get the email and password
var AuthDetails = require("./auth.json");
var meetup = require("./meetup.js");

var bot = new Discord.Client();

//when the bot is ready
bot.on("ready", function () {
	console.log("Ready to begin! Serving in " + bot.channels.length + " channels");
});

//when the bot disconnects
bot.on("disconnected", function () {
	//alert the console
	console.log("Disconnected!");

	//exit node.js with an error
	process.exit(1);
});

//when the bot receives a message
bot.on("message", function (msg) {
	//if message begins with "ping"
	if (msg.content.indexOf("!getevents") === 0) {

		meetup.getGroupEvents('23bshop')
		.then(function(events) {
			console.log("EVENTS: ", events);
			//send a message to the channel the ping message was sent in.
			bot.sendMessage(msg.channel, `Next up coming event: \`\`\`${events[1].name}\`\`\``);
		});
	}

});

bot.login(AuthDetails.email, AuthDetails.password);
