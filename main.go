package main

import (
	"flag"
	"github.com/tyler-sommer/squircy2/squircy"
)

var (
	config = flag.String("config", "", "configuration file")
)

func main() {
	flag.Parse()

	var c *squircy.Configuration
	if len(*config) == 0 {
		c = squircy.NewDefaultConfiguration()
	} else {
		c = squircy.NewConfiguration(*config)
	}

	mgr := squircy.NewManager(c)
	go mgr.Run()

	mgr.LoopCli()
}
