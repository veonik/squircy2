squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and is scriptable using Javascript.

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at 
runtime. Bind custom scripts to run when events occur in the application. Events can come from IRC, the CLI, or
even the web.


Installation
------------

Installing squIRCy2 is as easy as running:

```
go get -u github.com/tyler-sommer/squircy2
```

With squIRCy2 is installed, you can run it immediately with `squircy2` and a default configuration will be initialized
in `~/.squircy2`.

> For information on modifying and customizing squircy2 itself, see [CONTRIBUTING.md](CONTRIBUTING.md).


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`.

The Settings page allows you to modify squishy's nickname, username, and which server he connects to. Configure the 
Owner nickname and hostname to your information. These values are available from within each scripting language at 
runtime.

> The Owner Nickname and Hostname settings are available within your scripts.

From the Scripts page, you can add and edit scripts.

From the Dashboard page, you can see CLI and IRC history.


Scripting
---------

squIRCy2 embeds a Javascript interpreter, allowing you to write scripts to implement various bot behaviors.

### Javascript API

[otto](https://github.com/robertkrimen/otto) supports ECMAScript 5, less a regular expression incompatibility.
Additionally, the following functions are available to interact with the various squIRCy2 modules:

| Method | Description |
| ------ | ----------- |
| Irc.Join(channel) | Joins the given channel |
| Irc.Part(channel) | Parts the given channel |
| Irc.Privmsg(target, message) | Messages target with message. Target can be a user or a channel |
| Irc.CurrentNick() | Get the bot's current nickname |
| Irc.Nick(newNick) | Change the bot's nickname |
| Data.Get(key) | Gets a value with the given from the cross-vm storage |
| Data.Set(key, val) | Sets a value with the given key in the cross-vm storage |
| Http.Get(url) | Fetch the given url using a GET HTTP request |
| Config.OwnerNick() | Get the configured Owner Nickname |
| Config.OwnerHost() | Get the configured Owner Host |
| bind(eventName, fnName) | Add a handler of the given event type and function name |
| unbind(eventName, fnName) | Removes a handler of the given type and function name |
| setTimeout(fnName, delay) | Executes fnName after delay milliseconds |
| setInterval(fnName, delay) | Executes fnName every delay milliseconds |
| use(coll) | Opens and returns a repository for the given collection |

#### Repository methods

These are methods available on a repository returned by `use`.

| Method | Description |
| ------ | ----------- |
| repo.Fetch(id) | Attempts to load and return an entity with the given id |
| repo.FetchAll() | Returns a collection of all the entities in the repository |
| repo.Save(entity) | Saves the given entity |

### Event handlers

Event handlers can be registered with `bind` and `unbind`. Bind takes two parameters: the name of the
event, and the name of the function to call when the given event is triggered. 

Event handlers receive an Event object with additional information. An example Javascript handler:

```js
function handler(e) {
    // e is an object with all the transmitted event details
}
```

#### Binding a handler

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

#### Unbinding a handler

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

#### Events

| Event Name | Description |
| ---------- | ----------- |
| cli.INPUT | Input received from terminal |
| cli.OUTPUT | Output sent to terminal |
| irc.CONNECTING | Fired when first connecting to the IRC server |
| irc.CONNECT | Successfully connected to the IRC server |
| irc.DISCONNECT | Disconnected from the IRC server |
| irc.PRIVMSG | A message received, in a channel or a private message |
| irc.NOTICE | A notice received |
| irc.WILDCARD | Any IRC event |

> The IRC module also fires any IRC code as `irc.<code>`, for example 001 is `irc.001`, or NICK is `irc.NICK`.


Example Scripts
---------------

### Join channels on connect

```js
bind("irc.CONNECT", function(e) {
    Irc.Join('#squishyslab')
});
```

### Identify with Nickserv

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

Additional Info
---------------

squIRCy2 leverages [go-irc-event](https://github.com/thoj/go-ircevent) for IRC interaction. 
It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and 
dependency injection. [Tiedot](https://github.com/HouzuoGuo/tiedot) is used as the storage engine. 
squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto). Finally, it uses
[the stick templating engine](https://github.com/tyler-sommer/stick) for rendering HTML templates.
