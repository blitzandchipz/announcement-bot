'use strict';

const Discord = require("discord.js");

// Get the email and password
const AuthDetails = require("./auth.json");
const Commands = require("./commands.js");

const bot = new Discord.Client();

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
	// if message matches the command
	// TODO: replace this with single dynamic match of commands object
	if (msg.content.indexOf("!getnext") === 0) {
		//
		Commands.getEvents('23bshop')
			.then((result) => {
				bot.sendMessage(msg.channel, result);
			});
	}
});

bot.login(AuthDetails.email, AuthDetails.password);
