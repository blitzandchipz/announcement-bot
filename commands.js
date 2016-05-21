'use strict';

const Meetup = require("./meetup.js");
const fs = require('fs');

try {
  fs.accessSync('./config.json');
} catch (e) {
  fs.writeFileSync('config.json', '{}');
}

const config = require("./config.json");

const Commands = {
  setgroup: (content) => {
    config.urlname = content;
    const newConfig = JSON.stringify(config, null, 2);
    fs.writeFileSync('config.json', newConfig);
    return require('co')(function * () {
      return 'a';
    })
  },
  nextevent: (content) => {
    // returns a promise
    return Meetup.getGroupEvents(config.urlname).then(function (events) {
      // get first returned event
      let event = events[0];
      // TODO: check if this is corrent for empty events
      if (!event)
        return 'No upcoming events. Check back later.';

      // Finds if there is a number specified
      let number = content.match(/\d+/);
      // If theres no number assume 1
      number === null
        ? number = 1
        : number = number[0];
      // Create intro string
      let str = `**Next up coming event${number > 1
        ? 's**:\n'
        : '**: '}`
      let eventStr = '';
      for (let i = 0; i < number; i++) {
        event = events[i];
        // Event string template
        eventStr += `\`${event.name}\`\nWith ${event.yes_rsvp_count} ${event.group.who}\nLink: <${event.link}>\n\n`
      }
      return `${str}${eventStr}`;
    });
  }
};

module.exports = Commands;
