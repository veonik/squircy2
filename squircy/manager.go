package squircy

import (
	"github.com/go-martini/martini"
)

type Manager struct {
	*martini.ClassicMartini
}

func NewManager() (manager *Manager) {
	manager = &Manager{martini.Classic()}
	manager.Map(NewConfiguration("config.json"))
	res, err := manager.Invoke(newIrcConnection)
	if err != nil {
		panic(err)
	}
	conn := res[0].Interface()
	manager.Map(conn)
	res, err = manager.Invoke(newHandlerCollection)
	if err != nil {
		panic(err)
	}
	handlers := res[0].Interface()
	manager.Map(handlers)
	
	return
}