squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and supports Javascript, Lua, and a small dialect of Lisp. 

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at runtime.


> This program exposes scripting languages to IRC. Take care to set the Owner nickname and hostname settings properly.


Installation
------------

As a prerequisite, the Lua 5.1 package must be installed on your system.

[See here for more information](https://github.com/aarzilli/golua/blob/master/README.md) on getting Go Bindings for the Lua C API setup or just visit [the Lua download page](http://www.lua.org/download.html).


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

The Settings page allows you to modify squishy's nickname, username, and which server he connects to. Configure the Owner nickname and hostname to your information.

> The Owner Nickname and Hostname settings limit in-IRC REPL to that user. Ensure these are configured properly.

From the Scripts page, you can add and edit scripts.

From the Dashboard page, you can re-initialize scripts and connect or disconnect from IRC.


Exposed API
-----------

squIRCy2 exposes a small API to each scripting language.

### Javascript API

| Method | Description |
| ------ | ----------- |
| `Irc.Join(channel)` | Joins the given channel |
| `Irc.Part(channel)` | Parts the given channel |
| `Irc.Privmsg(target, message)` | Messages target with message. Target can be a user or a channel |
| `Data.Get(key)` | Gets a value with the given from the cross-vm storage |
| `Data.Set(key, val)` | Sets a value with the given key in the cross-vm storage |
| `Script.AddHandler(type, fnName)` | Add a PRIVMSG handler of the given type ("js", "lua", or "lisp") and function name |
| `Script.RemoveHandler(type, fnName)` | Removes a PRIVMSG handler of the given type and function name |
| `Script.On(type, event, fnName)` | Add a handler of the given type to a specific event like "NOTICE" or "001" |
| `replyTarget()` | The current reply target. If the current message was received in a channel, this will be the channel name. Otherwise it will be a nickname |
| `print(message)`  | Replies to the current reply target with the given message |

### Lua API

```
joinchan(channel)
partchan(channel)
privmsg(target, channel)
getex(key)
setex(key, value)
addhandler(type, fnName)
removehandler(type, fnName)
on(type, event, fnName)
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
(on type event fnName)
(print message)
```

### IRC event handlers

Chat (PRIVMSG) handlers can be registered with `Script.AddHandler` in Javascript or `addhandler` in Lua and Lisp. This function takes a type, one of: js, lua, lisp. It also takes the name of a function that should have the following signature:

```js
function handler(code, target, nick, message) {
	// code is the irc event code, like PRIVMSG, NOTICE, or 001
	// target received the message
	// nick sent the message
}
```

Additionally, exposed to the Javascript VM is `Script.On` which takes a script type, an event type, and a function name. The signature of the function is the same as a chat handler.


In-IRC REPL
-----------

squIRCy2 comes with an in-IRC REPL, though the print functionality needs to be explicitly called. A REPL session can be started by messaging squIRCy2, either in channel or private message: `!repl js`. Replace "js" with "lua" or "lisp" to start a session of that type. Message `!repl end` to end the REPL session.

```
<veonik> !repl js
<squishyj> Javascript REPL session started.
<veonik> function test(x, y) { print(x * y) }
<veonik> test(10, 5)
<squishyj> 50
<veonik> !repl end
<squishyj> Javascript REPL session ended.
```


Example Scripts
---------------

### Join channels on connect (Javascript example)

```js
function handleWelcome(code, target, nick, message) {
    Irc.Join('#squishyslab')
}
Script.On("js", "001", "handleWelcome");
```

### Identify with Nickserv (Javascript example)

```js
function handleNickserv(code, target, nick, message) {
    if (nick == "NickServ" && message.indexOf("identify") >= 0) {
        Irc.Privmsg("NickServ", "IDENTIFY superlongandsecurepassword");
        console.log("Identified with Nickserv");
    }
}
Script.On("js", "NOTICE", "handleNickserv");
```


Additional Info
---------------

squIRCy2 leverages [go-irc-event](https://github.com/thoj/go-ircevent) for IRC interaction. It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and dependency injection. [Tiedot](https://github.com/HouzuoGuo/tiedot) is used as the storage engine.

Additionally, squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto), [Go language bindings for Lua](https://github.com/aarzilli/golua), and a [forked Lisp interpreter](https://github.com/veonik/go-lisp) based on [janne/go-lisp](https://github.com/janne/go-lisp).
