squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is written in Go and is scriptable using an embedded JavaScript runtime.

It sports a web management interface for writing scripts and bot management, as well as dynamic script reloading at runtime. 

Bind custom scripts to run when events occur in the application. Events can come from IRC, the CLI, or even the web.


Installation
------------

Installing squIRCy2 is as easy as running:

```bash
go get -u github.com/tyler-sommer/squircy2/...
```

With squIRCy2 is installed, a default configuration will be initialized in `~/.squircy2`. You can run it immediately with `squircy2`:

```bash
squircy2
```

> For information on modifying and customizing squIRcy2, see [CONTRIBUTING.md](CONTRIBUTING.md).


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`.

> squIRCy2 supports SSL and HTTP Basic authentication, too.

The Settings page allows you to modify squishy's nickname, username, and which server it connects to. Configure the Owner nickname and hostname to your information.

From the Scripts page, you can add and edit scripts.

> See the [JavaScript API reference](resources/js-api.md) for more details.

From the Dashboard page, you can see CLI and IRC history.

The Webhooks page allows you to create and configure squIRCy2 webhooks.

> See the [Webhooks section](resources/webhooks.md) for more details.

From the REPL, you can write, run, and see the result of code.


Scripting
---------

squIRCy2 embeds a JavaScript interpreter, allowing you to write scripts to implement various bot behaviors.

### JavaScript API

A full introduction to the squIRCy2 JavaScript API can be found in [the JavaScript API reference](resources/js-api.md).

#### Example scripts

Check the [Example Scripts section](resources/examples.md) for ideas for squIRCy2 scripts.

### Webhooks

See the [dedicated section on Webhooks](resources/webhooks.md).

Additional Info
---------------

squIRCy2 leverages [go-irc-event](https://github.com/thoj/go-ircevent) for IRC interaction. 
It makes use of [martini](https://github.com/go-martini/martini) for serving web requests and 
dependency injection. [Tiedot](https://github.com/HouzuoGuo/tiedot) is used as the storage engine. 
squIRCy2 embeds the [otto JavaScript VM](https://github.com/robertkrimen/otto). Finally, it uses
[the stick templating engine](https://github.com/tyler-sommer/stick) for rendering HTML templates.
