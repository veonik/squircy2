package main

import (
	"flag"
	"github.com/tyler-sommer/squircy2/squircy"
)

var daemonizeFlag = flag.Bool("daemonize", false, "Run as a daemon")

func main() {
	flag.Parse()

	mgr := squircy.NewManager()

	if !*daemonizeFlag {
		go mgr.Run()

		mgr.LoopCli()

	} else {
		mgr.Run()
	}
}
