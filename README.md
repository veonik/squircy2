squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and is scriptable using Javascript.

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at 
runtime. Bind custom scripts to run when events occur in the application. Events can come from IRC, the CLI, or
even the web.


Installation
------------

installing squIRCy2 is as easy as running:

```
go get -u github.com/tyler-sommer/squircy2
```

With squIRCy2 is installed, you can run it immediately with `squircy2` and a default configuration will be initialized.


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`.

The Settings page allows you to modify squishy's nickname, username, and which server he connects to. Configure the 
Owner nickname and hostname to your information. These values are available from within each scripting language at 
runtime.

> The Owner Nickname and Hostname settings are available within your scripts.

From the Scripts page, you can add and edit scripts.

From the Dashboard page, you can see CLI and IRC history.


Exposed API
-----------

squIRCy2 exposes a small API to each scripting language.

### Javascript API

[otto](https://github.com/robertkrimen/otto) supports ECMAScript 5, less a regular expression incompatibility.

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

### Event handlers

Event handlers can be registered with `bind` and `unbind`. Bind takes two parameters: the name of the
event, and the name of the function to call when the given event is triggered. 

Event handlers receive an Event object with additional information. An example Javascript handler:

```js
function handler(e) {
    // e is an object with all the transmitted event details
}
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

### Join channels on connect (Javascript example)

```js
function handleWelcome(code, target, nick, message) {
    Irc.Join('#squishyslab')
}
bind("irc.CONNECT", "handleWelcome");
```

### Identify with Nickserv (Javascript example)

```js
function handleNickserv(code, target, nick, message) {
    if (nick == "NickServ" && message.indexOf("identify") >= 0) {
        Irc.Privmsg("NickServ", "IDENTIFY superlongandsecurepassword");
        console.log("Identified with Nickserv");
    }
}
bind("irc.NOTICE", "handleNickserv");
```


Additional Info
---------------

squIRCy2 leverages [go-irc-event](https://github.com/thoj/go-ircevent) for IRC interaction. It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and dependency injection. [Tiedot](https://github.com/HouzuoGuo/tiedot) is used as the storage engine. squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto).
