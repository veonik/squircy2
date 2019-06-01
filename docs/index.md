Scriptable IRC bot
==================

squIRCy2 is an IRC bot written in Go and is scriptable using an embedded JavaScript runtime.

It features a robust CLI with REPL and auto-completion, a web management interface, an embedded document store, and event based in-program communication. 

Customize the bot's functionality with JavaScript. Bind event handlers to handle events that occur in the application. Events come from IRC, the CLI, or even the web.


Installation
------------

Install squIRCy2 with `go get`.

```bash
go get -u "github.com/veonik/squircy2/..."
```

After squIRCy2 is installed, you can run it immediately with `squircy2`. On first run, a default configuration will be initialized in `~/.squircy2`.

```
squircy2
```

> For information on modifying and customizing squIRcy2, see [Contributing](customizing.md).


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
 
Enter the interactive Read-Evaluate-Print Loop (or "REPL") with the `repl` command. This REPL features history navigable with the arrow keys, history search with CTRL+R, and auto-completion with TAB.


### Scripting

squIRCy2 embeds a JavaScript interpreter, allowing you to write scripts to implement various bot behaviors.

A full introduction to the squIRCy2 JavaScript API can be found in [the JavaScript API reference](js-api.md). 

Check the [Example Scripts section](examples.md) for ideas for squIRCy2 scripts.

### Webhooks

See the [dedicated section on Webhooks](webhooks.md).
