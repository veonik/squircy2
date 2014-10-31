squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and supports Javascript, Lua, a dialect of Lisp, and the Anko scripting language. 

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at 
runtime.


Installation
------------

As a prerequisite, the Lua 5.1 package must be installed on your system.

[See here for more information](https://github.com/aarzilli/golua/blob/master/README.md) on getting Go Bindings for the 
Lua C API setup or just visit [the Lua download page](http://www.lua.org/download.html).


**Installing Lua 5.1 on Ubuntu or Debian**

Lua 5.1 is available on either platform by default. Install them with aptitude by running:

```
sudo apt-get install lua5.1 lua5.1-dev
```

**Installing Lua 5.1 on CentOS**

Some versions of CentOS comes with Lua. Ensure it is installed by running:

```
sudo yum install lua lua-devel
```

**Installing Lua 5.1 on Mac OSX**

[Download the prebuilt binary](http://luabinaries.sourceforge.net/download.html) and put it in `/usr/local/lib`


### Installing squIRCy2

Once Lua is ready to go, installing squIRCy2 is as easy as running:

```
go get -u github.com/tyler-sommer/squircy2
```

> On some platforms, like RHEL, you may need to specify `-tags llua` on the command line to properly locate the library.

With squIRCy2 is installed, you can run it immediately with `squircy2` and a default configuration will be initialized.


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`.

The Settings page allows you to modify squishy's nickname, username, and which server he connects to. Configure the 
Owner nickname and hostname to your information. These values are available from within each scripting language at 
runtime.

> The Owner Nickname and Hostname settings limit in-IRC REPL to that user. Ensure these are configured properly.

From the Scripts page, you can add and edit scripts.

From the Dashboard page, you can re-initialize scripts and connect or disconnect from IRC.


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
| Data.Get(key) | Gets a value with the given from the cross-vm storage |
| Data.Set(key, val) | Sets a value with the given key in the cross-vm storage |
| Http.Get(url) | Fetch the given url using a GET HTTP request |
| bind(eventName, fnName) | Add a handler of the given event type and function name |
| unbind(eventName, fnName) | Removes a handler of the given type and function name |

### Lua API

[Go-lua](https://github.com/aarzilli/golua) binds many built in libraries

| Method | Description |
| ------ | ----------- |
| joinchan(channel) | Joins the given channel |
| partchan(channel) | Parts the given channel |
| privmsg(target, message) | Messages target with message. Target can be a user or a channel |
| getex(key) | Gets a value with the given from the cross-vm storage |
| setex(key, val) | Sets a value with the given key in the cross-vm storage |
| httpget(url) | Fetch the given url using a GET HTTP request |
| bind(eventName, fnName) | Add a handler of the given event type and function name |
| unbind(eventName, fnName) | Removes a handler of the given type and function name |

### Lisp API

[Glisp](https://github.com/zhemao/glisp) also has many built-in functions.

| Method | Description |
| ------ | ----------- |
| (joinchan channel) | Joins the given channel |
| (partchan channel) | Parts the given channel |
| (privmsg target message) | Messages target with message. Target can be a user or a channel |
| (getex key) | Gets a value with the given from the cross-vm storage |
| (setex key val) | Sets a value with the given key in the cross-vm storage |
| (httpget url) | Fetch the given url using a GET HTTP request |
| (bind eventName fnName) | Add a handler of the given event type and function name |
| (unbind eventName fnName) | Removes a handler of the given type and function name |

### Anko API

[Anko](https://github.com/mattn/anko) offers a myriad of built-in functions and has a go-like syntax

| Method | Description |
| ------ | ----------- |
| irc.Join(channel) | Joins the given channel |
| irc.Part(channel) | Parts the given channel |
| irc.Privmsg(target, message) | Messages target with message. Target can be a user or a channel |
| data.Get(key) | Gets a value with the given from the cross-vm storage |
| data.Set(key, val) | Sets a value with the given key in the cross-vm storage |
| bind(eventName, fnName) | Add a handler of the given event type and function name |
| unbind(eventName, fnName) | Removes a handler of the given type and function name |


### Event handlers

Event handlers can be registered with `bind` and `unbind` in all languages. Bind takes two parameters: the name of the
event, and the name of the function to call when the given event is triggered. 

> Bind will only bind functions to the language it is called from; Lua scripts can only `bind` or `unbind` Lua scripts. 

Event handlers should take four parameters: code, target, nick, message. An example javascript handler:

```js
function handler(code, target, nick, message) {
	// code is the IRC event code, like PRIVMSG, NOTICE, or 001
	// target received the message
	// nick sent the message
}
```

#### Events

| Event Name | Description |
| ---------- | ----------- |
| irc.CONNECTING | Fired when first connecting to the IRC server |
| irc.CONNECT | Successfully connected to the IRC server |
| irc.DISCONNECT | Disconnected from the IRC server |
| irc.PRIVMSG | A message received, in a channel or a private message |
| irc.NOTICE | A notice received |
| irc.WILDCARD | Any IRC event |

> The IRC module also fires any IRC code as `irc.<code>`, for example 001 is `irc.001`, or `irc.NICK`.


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

squIRCy2 leverages [go-irc-event](https://github.com/thoj/go-ircevent) for IRC interaction. It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and dependency injection. [Tiedot](https://github.com/HouzuoGuo/tiedot) is used as the storage engine.

squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto), [Go language bindings for Lua](https://github.com/aarzilli/golua), [Glisp lisp interpreter](https://github.com/zhemao/glisp), and the [Anko scripting language](https://github.com/mattn/anko).
