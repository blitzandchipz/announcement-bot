'use strict';

const Discord = require("discord.js");

// Get the email and password
const AuthDetails = require("./auth.json");
const Commands = require("./commands.js");

const bot = new Discord.Client();

//when the bot is ready
bot.on("ready", function () {
	console.log("Meetup bot ready! Serving in " + bot.channels.length + " channels");
	console.log("\nAvaiable commands:");
	Object.keys(Commands).map((command, i) => {
		console.log(`!${command}`);
	})
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
	// If message mathes a command from the command file
	Object.keys(Commands).map((command) => {
		if (msg.content.indexOf(`!${command}`) === 0) {
			// Run matched command function and then print out the result
			Commands[command](msg.content).then((result) => {
					bot.sendMessage(msg.channel, result);
				});
		}
	});
});

bot.login(AuthDetails.email, AuthDetails.password);
