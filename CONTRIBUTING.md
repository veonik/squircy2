Contributing Guidelines
=======================

This document describes how to modify and customize squIRCy2, and optionally contribute
your changes back to the original squIRCy2 codebase.


Overview
--------

The master branch contains active development. It should be considered unstable. Versions are tagged in the repository.

squIRCy2 is made up of a few different parts:

* The **CLI interface** lets a user interact with squIRCy from the command line.
  > See `squircy/cli.go` for the CLI related code.
  
* The **Web interface** allows a user to work with squIRCy from the web.
  > See `squircy/controller.go` for code related to all controller actions.
  
* **Static assets** (like CSS) and **views** are embedded in the squIRCy2 binary using the 
  [go-bindata utility](https://github.com/jteeuwen/go-bindata).
  > Running `go generate ./...` prior to building squIRCy2 will regenerate the binary forms of 
    the assets into `squircy/bindata.go`. See ["Building the project." for more details](#3-building-the-project).

* A **persistence layer** is available using [tiedot](https://github.com/HouzuoGuo/tiedot). A basic repository
  is implemented in `squircy/data` with more specific implementations in `squircy/config/model.go` and 
  `squircy/script/model.go`.

* The **Javascript VM** is a lightly wrapped [otto runtime](https://github.com/robertkrimen/otto). Code related to
  customizing the VM and executing predefined scripts is in `squircy/script`.

* Other parts include an event dispatcher in `squircy/event`, a wrapper for 
  [go-ircevent](https://github.com/thoj/go-ircevent) in `squircy/irc`, and an overall "manager" that embeds an
  [inject.Injector](https://github.com/codegangsta/inject) in `squircy/manager.go`.


Modifying squIRCy2
------------------

### 1. Fork this repository.

Fork the squIRCy2 repository on GitHub.

### 2. Clone the repository and checkout the master branch.

```bash
git clone git@github.com/you/squircy2.git squircy2
cd squircy2
```

### 3. Building the project.

There are two options for building squIRCy2: default (release) and debug. Debug mode is recommended for development.


#### 3a. Build the project in debug mode.

Building squIRCy2 in debug mode makes it so that changes to views and static assets do not require a rebuilding 
of the squIRCy2 code.

```bash
go build -tags debug ./cmd/squircy2/...
./squircy2
```

Alternatively, you can run squIRCy2 with `go run`:
 
```bash
go run -tags debug ./cmd/squircy2/main.go
```

> **Note:** When running in debug mode, squIRCy2 will look in the current working directory for views and assets.
  Specifically, views must be in `$PWD/views` and static assets in `$PWD/public`.


#### 3b. Build the project in default mode (aka release mode).

Building squIRCy2 for release takes two steps. First, we have to convert our assets and views into binary form. 
This relies on the [go-bindata utility](https://github.com/jteeuwen/go-bindata) which is installed along-side squIRCy2 
and must be in your PATH.

```bash
go generate ./...
go build ./cmd/squircy2/...
./squircy2
```

### 4. Make your changes.

With squIRCy forked and set up for development, you're ready to get started. Depending on how you're
building squIRCy, you may need to do different things to see your changes take effect.

Modifying Go source code will always require that you stop the currently running squIRCy2 process, then 
running `go build && ./squircy2` or `go run main.go` to rebuild the binary and run it.
 
When running in debug mode, modifying assets requires a refresh in your browser. However, if you're running
in default (release) mode, then you will have to run `go generate ./...` before rebuilding and running the 
squIRCy2 binary.


Contributing back to squIRCy2
-----------------------------

If you'd like to create a new feature or fix a bug and contribute your modification to the original repository, 
follow these steps.

### 1. Create a new branch.

```bash
cd /path/to/your/forks/code
git checkout -b new-and-awesome master
```

This will checkout a new branch named `new-and-awesome` based on the current master branch.


### 2. Implement and publish your change.

```bash
git add .
git commit -m "Made some changes"
git push origin new-and-awesome
```


### 3. Create a pull request.

From your repository on the GitHub interface, click the pull request button. Select your feature branch and ensure
the master branch of squIRCy2 is selected.
