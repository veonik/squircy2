package main

import (
	"github.com/tyler-sommer/squircy2/squircy"
	"flag"
)

var daemonizeFlag = flag.Bool("daemonize", false, "Run as a daemon")

func main() {
	flag.Parse()

	mgr := squircy.NewManager()

	if (!*daemonizeFlag) {
		go mgr.Run()

		mgr.LoopCli()

	} else {
		mgr.Run()
	}
}
