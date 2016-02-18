package main

import (
	"flag"

	"github.com/tyler-sommer/squircy2/squircy"
	"gopkg.in/mattes/go-expand-tilde.v1"
)

var daemonizeFlag = flag.Bool("daemonize", false, "Run as a daemon.")
var rootPathFlag = flag.String("root-path", "~/.squircy2", "Specify a custom root path.")

func main() {
	flag.Parse()

	root, err := tilde.Expand(*rootPathFlag)
	if err != nil {
		panic(err)
	}

	mgr := squircy.NewManager(root)

	if !*daemonizeFlag {
		go mgr.Run()

		mgr.LoopCli()

	} else {
		mgr.Run()
	}
}
