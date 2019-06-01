Event handlers
==============

Event handlers can be registered with `bind` and `unbind`. Bind takes two parameters: the name of the
event, and the name of the function to call when the given event is triggered. 

Event handlers receive an Event object with additional information. An example Javascript handler:

```js
function handler(e) {
    // e is an object with all the transmitted event details
}
```

## Binding a handler

An event handler can either be a named function, or more commonly, a function itself.

```js
bind("irc.PRIVMSG", function(e) {
    console.log("Received message from "+e.Nick);
});

// or
function privmsgHandler(e) {
    console.log("Received message from "+e.Nick);
}
bind("irc.PRIVMSG", privmsgHandler);
```

## Unbinding a handler

To unbind a handler, you must retain a reference to that function. Generally this means keeping
a reference to the original handler around.

```js
function privmsgHandler(e) {
    if (e.Nick == "Someone") {
        // Unbind after Someone sends a message
        unbind("irc.PRIVMSG", privmsgHandler);
    }
}
bind("irc.PRIVMSG", privmsgHandler);
```

Events
------

| Event Name | Description |
| ---------- | ----------- |
| `cli.INPUT` | Input received from terminal |
| `cli.OUTPUT` | Output sent to terminal |
| `irc.CONNECTING` | Fired when first connecting to the IRC server |
| `irc.CONNECT` | Successfully connected to the IRC server |
| `irc.DISCONNECT` | Disconnected from the IRC server |
| `irc.PRIVMSG` | A message received, in a channel or a private message |
| `irc.NOTICE` | A notice received |
| `irc.WILDCARD` | Any IRC event |
| `irc.[code]` | A specific IRC event, given its RFC Code. For example, 001 `irc.001` or NICK is `irc.NICK`. |
| `hook.WILDCARD` | Fired whenever a valid webhook is received. |
| `hook.[ID]` | Fired when the webhook with [ID] is received. |


