squIRCy2
========

##### the scriptable IRC bot

squIRCy2 is an IRC bot written in Go and is scriptable using an embedded JavaScript runtime.

It features a robust CLI with REPL and auto-completion, a web management interface, an embedded document store, and event based in-program communication. 

Customize the bot's functionality with JavaScript scripts. Bind event handlers to handle events that occur in the application. Events come from IRC, the CLI, or even the web.


Installation
------------

Install squIRCy2 with `go get`.

```bash
go get -u github.com/tyler-sommer/squircy2/cmd/squircy2
```

After squIRCy2 is installed, you can run it immediately with `squircy2`. On first run, a default configuration will be initialized in `~/.squircy2`.

```bash
squircy2
```

> For information on modifying and customizing squIRcy2, see [Contributing](resources/customizing.md).


Usage
-----

squIRCy2 command-line usage.

```
squIRcy2 version dev (go1.8)
Usage: squircy2 [-no-interactive] [-root-path <config root>] [-version]

  -no-interactive
    	Run without user interaction.
  -root-path string
    	Specify a custom root path. (default "~/.squircy2")
  -version
    	Display the version and exit.
```

Customize where squIRCy2 stores data by specifying a custom `-root-path`. Specifying `-no-interactive` will disable the CLI.

### Command line

The squIRCy2 command line starts in "Command Mode" which offers a few basic commands, including connecting/disconnecting from IRC, enabling/disabling the web server, and starting a REPL.
 
Enter "REPL Mode" by entering the `repl` command. This REPL features history navigable with the arrow keys, history search with CTRL+R, and auto-completion with TAB.


Configuration
-------------

Once the bot is up and running, you can access the web management interface via `localhost:3000`. From this interface you can write scripts and use a web-based REPL, as well as configure squIRCy2 to your liking.

### Configuration overview

From the Settings page, you can configure:

* **IRC**
  * Configure the Network, Nickname and Username information for the bot. You can also enable TLS-- be sure to specify a TLS-capable port for the Network.
  * Enable SASL authentication and enter your NickServ username and password. **Note these are stored plaintext in a blob format.**
  * Owner Nick and Host can be used from within scripts to verify a user's ownership of the bot. See [the JavaScript reference](resources/js-api.md) for more information.
* **Script Management**
  * If you prefer to work with an external editor, you can have squIRCy2 load scripts on the filesystem. Specify a path, enable the option, and restart squIRCy2.
  * Import and Export scripts from the embedded data store.
* **Web Interface**
  * Disable the web interface completely by disabling both HTTP and HTTPS.
  * Configure HTTPS by specifying a certificate file and private key.
  * Configure HTTP(S) Basic Authentication with a Username and Password. **Note these are stored plaintext in a blob format.**


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
squIRCy2 embeds the [otto JavaScript VM](https://github.com/robertkrimen/otto) for scripting and it uses
[the stick templating engine](https://github.com/tyler-sommer/stick) for rendering HTML templates.
