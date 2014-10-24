squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and supports Javascript, Lua, and a small dialect of Lisp. 

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at runtime.


> This program exposes scripting languages to IRC and should not be considered safe.


Installation
------------

As a prerequisite, the Lua 5.1 package must be installed on your system.

[See here for more information](https://github.com/aarzilli/golua/blob/master/README.md) on getting Go Bindings for the Lua C API setup or just visit [the download page](http://www.lua.org/download.html).

Additionally, squIRCy2 requires access to a Redis server. [See here for more info](http://redis.io/)


Once Lua and Redis are ready to go, the easiest way to install squIRCy2 is by running:

```
go get github.com/tyler-sommer/squircy2
```

Once squIRCy2 is installed, you'll need to create a configuration file. You can copy defaults as such:

```
cd $GOPATH/src/github.com/tyler-sommer/squircy2
cp config.json.dist config.json
```

Then edit config.json to suit your needs. Then start the bot by running `squircy2`.


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`


Exposed API
-----------

squIRCy2 exposes a small API to each scripting language.

### Javascript API

```
Irc.Join(channel)
Irc.Part(channel)
Irc.Privmsg(target, message)
Data.Get(key)
Data.Set(key, val)
Script.AddHandler(type, fnName)
Script.RemoveHandler(type, fnName)
print(message)
```

### Lua API

```
joinchan(channel)
partchan(channel)
privmsg(target, channel)
getex(key)
setex(key, value)
addhandler(type, fnName)
removehandler(type, fnName)
print(message)
```

### Lisp API

```
(joinchan channel)
(partchan channel)
(privmsg target channel)
(getex key)
(setex key value)
(addhandler type fnName)
(removehandler type fnName)
(print message)
```

### Chat handlers

Chat handlers can be registered with `Script.AddHandler` in Javascript or `addhandler` in Lua and Lisp. This function takes a type, one of: js, lua, lisp. It also takes the name of a function that should have the following signature:

```js
function handler(target, nick, message) {
	// target received the message
	// nick sent the message
}
```


Additional Info
---------------

squIRCy2 leverages [go-irc-event ](https://github.com/thoj/go-ircevent) for IRC interaction. It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and dependency injection. Redis serves as a data storage service, with the [radix client](https://github.com/fzzy/radix) handling interaction.

Additionally, squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto), [Go language bindings for Lua](https://github.com/aarzilli/golua), and a [forked Lisp interpreter](https://github.com/veonik/go-lisp) based on [janne/go-lisp](https://github.com/janne/go-lisp).
