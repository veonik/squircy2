package main

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/veonik/squircy2"
	"gopkg.in/mattes/go-expand-tilde.v1"
)

var nonInteractiveFlag = flag.Bool("no-interactive", false, "Run without user interaction.")
var rootPathFlag = flag.String("root-path", "~/.squircy2", "Specify a custom root path.")
var versionFlag = flag.Bool("version", false, "Display the version and exit.")

var Version = "dev"
var GoVersion = runtime.Version()

func printVersion() {
	fmt.Printf("squIRcy2 version %s (%s)", Version, GoVersion)
}

func main() {
	flag.Usage = func() {
		printVersion()
		fmt.Println("\nUsage: squircy2 [-no-interactive] [-root-path <config root>] [-version]\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *versionFlag {
		printVersion()
		fmt.Println()
		return
	}

	root, err := tilde.Expand(*rootPathFlag)
	if err != nil {
		panic(err)
	}

	mgr := squircy2.NewManager(root)

	go mgr.Web().ListenAndServe()
	mgr.IRC().AutoConnect()

	quit := make(chan struct{})
	if !*nonInteractiveFlag {
		go func() {
			mgr.LoopCLI()
			close(quit)
		}()
	}

	select {
	case <-quit:
		// Exit
		// TODO: Non-interactive mode may end up disconnected and without a web server running.
	}
}
