package main

import (
	"github.com/tyler-sommer/squircy2/squircy"
	"flag"
)

var daemonize bool
func init() {
	flag.BoolVar(&daemonize, "daemonize", true, "Run as a daemon")
}

func main() {
	mgr := squircy.NewManager()

	if (!daemonize) {
		go mgr.Run()

		mgr.LoopCli()

	} else {
		mgr.Run()
	}
}
