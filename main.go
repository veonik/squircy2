package main

import (
	"flag"
	"github.com/tyler-sommer/squircy2/squircy"
)

var daemonizeFlag = flag.Bool("daemonize", false, "Run as a daemon.")
var rootPathFlag = flag.String("root-path", "", "Specify a custom root path.")

func main() {
	flag.Parse()

	mgr := squircy.NewManager(*rootPathFlag)

	if !*daemonizeFlag {
		go mgr.Run()

		mgr.LoopCli()

	} else {
		mgr.Run()
	}
}
