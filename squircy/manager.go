package squircy

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"log"
	"os"
)

type Manager struct {
	*martini.ClassicMartini
}

func NewManager() (manager *Manager) {
	manager = &Manager{martini.Classic()}
	manager.Map(NewConfiguration("config.json"))
	manager.Map(log.New(os.Stdout, "[squircy] ", 0))
	manager.invokeAndMap(newIrcConnection)
	manager.invokeAndMap(newHandlerCollection)
	manager.invokeAndMap(newRedisClient)

	manager.configure()

	return
}

func (manager *Manager) invokeAndMap(fn interface{}) {
	res, err := manager.Invoke(fn)
	if err != nil {
		panic(err)
	}
	manager.Map(res[0].Interface())
}

func (manager *Manager) configure() {
	manager.Use(render.Renderer(render.Options{
		Directory:  "views",
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	}))
	manager.Get("/", indexAction)
}
