package main

import (
	"flag"

	"gopkg.in/mattes/go-expand-tilde.v1"
	"github.com/tyler-sommer/squircy2"
)

var nonInteractiveFlag = flag.Bool("no-interactive", false, "Run without user interaction.")
var rootPathFlag = flag.String("root-path", "~/.squircy2", "Specify a custom root path.")

func main() {
	flag.Parse()

	root, err := tilde.Expand(*rootPathFlag)
	if err != nil {
		panic(err)
	}

	mgr := squircy2.NewManager(root)

	go mgr.ListenAndServe()
	mgr.AutoConnect()

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
