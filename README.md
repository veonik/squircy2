squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and supports Javascript, Lua, and a small dialect of Lisp. 

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at runtime.


> Note: This program exposes scripting languages to IRC and should not be considered safe.


installation
------------

There is a to-do note about this somewhere...


additional info
---------------

squIRCy2 leverages [go-irc-event ](https://github.com/thoj/go-ircevent) for IRC interaction. It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and dependency injection. Redis serves as a data storage service, with the [radix client](https://github.com/fzzy/radix) handling interaction.

Additionally, squIRCy2 embeds the [otto Javascript VM](https://github.com/robertkrimen/otto), [Go language bindings for Lua](https://github.com/aarzilli/golua), and a [forked Lisp interpreter](https://github.com/veonik/go-lisp) based on [janne/go-lisp](https://github.com/janne/go-lisp).
