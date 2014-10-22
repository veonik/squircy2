package main

import (
	"squircy2/squircy"
)

func main() {
	mgr := squircy.NewManager()
	go mgr.Run()

	mgr.LoopCli()
}
