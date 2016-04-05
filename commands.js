'use strict';

const Meetup = require("./meetup.js");

const Commands = {
    getEvents: (urlname) => {
      return Meetup.getGroupEvents(urlname)
      .then(function(events) {
          //send a message to the channel the ping message was sent in.
          return `Next up coming event: \`\`\`${events[0].name}\`\`\``;
        });
    }
};

module.exports = Commands;
