package main

import (
	"./squircy"
)

func main() {
	mgr := squircy.NewManager()
	
	mgr.Run()
}