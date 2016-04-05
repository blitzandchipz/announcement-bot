'use strict';

const fetch = require("node-fetch");
const co = require('co');

const hostname = "https://api.meetup.com";
const api = require("./auth.json").api;

const Meetup = {
  getGroupEvents: function (urlname, scroll) {
    if (!urlname) {
      console.log('You must provide a urlname for getGroupEvents');
      return -1;
    }
    if (!scroll) scroll = 'next_upcoming';
    const url = `${hostname}/${urlname}/events?scroll=${scroll}&key=${api}&sign=true`;


    return co(function *() {
      const res = yield fetch(url);
      return res.json();
    })
  }
}

module.exports = Meetup;
