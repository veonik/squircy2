package main

import (
	"./squircy"
)

func main() {
	mgr := squircy.NewManager()
	go mgr.Run()

	mgr.LoopCli()
}
