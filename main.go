package main

import (
	"github.com/tyler-sommer/squircy2"
)

func main() {
	mgr := squircy.NewManager()
	go mgr.Run()

	mgr.LoopCli()
}
