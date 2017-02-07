Example Scripts
---------------

### Join channels on connect

```js
bind("irc.CONNECT", function(e) {
    Irc.Join('#squishyslab')
});
```

### Identify with Nickserv

> A more comprehensive example is defined in [examples/nickserv-auth.js](examples/nickserv-auth.js).

In the example below, a handler is bound to the `irc.NOTICE` event. When NickServ notices you,
requesting you identify, it will reply with your password.

```js
function handleNickserv(e) {
    if (e.Nick == "NickServ" && e.Message.indexOf("identify") >= 0) {
        Irc.Privmsg("NickServ", "IDENTIFY superlongandsecurepassword");
        console.log("Identified with Nickserv");
    }
}
bind("irc.NOTICE", handleNickserv);
```

When your handler function is invoked, an object (`e` in the example) is passed as 
the first parameter. This object has different properties depending on the event.

### Iterate over an event's properties

```js
bind("irc.WILDCARD", function(e) {
    for (var i in e) {
        console.log(i+": "+e[i]);
    }
});
```

### Keep track of channel participants

An example script that keeps track of who is participating in channels the bot is on is defined in [examples/channel-names.js](examples/channel-names.js).