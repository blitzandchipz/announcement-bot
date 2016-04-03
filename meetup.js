var fetch = require("node-fetch");

var hostname = "https://api.meetup.com";
var api = require("./auth.json").api;

var meetup = new Object();
meetup.getGroupEvents = function (urlname, scroll) {
  if (!urlname) {
    console.log('You must provide a urlname for getGroupEvents');
    return -1;
  }
  if (!scroll) scroll = 'next_upcoming';
  var url = `${hostname}/${urlname}/events?scroll=${scroll}&key=${api}&sign=true`;

  return fetch(url)
  .then(function(res){
    return res.json();
  })
  .then(function(json){
    return json;
  });
}

module.exports = meetup;
