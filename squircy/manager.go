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
	manager.invokeAndMap(newRedisClient)
	h := manager.invokeAndMap(newHandlerCollection).(*HandlerCollection)
	nickservHandler := manager.invokeAndMap(newNickservHandler).(*NickservHandler)
	scriptHandler := manager.invokeAndMap(newScriptHandler).(*ScriptHandler)

	h.Add(nickservHandler)
	h.Add(scriptHandler)

	manager.configure()

	return
}

func (manager *Manager) invokeAndMap(fn interface{}) interface{} {
	res, err := manager.Invoke(fn)
	if err != nil {
		panic(err)
	}

	val := res[0].Interface()
	manager.Map(val)

	return val
}

func (manager *Manager) configure() {
	manager.Use(render.Renderer(render.Options{
		Directory:  "views",
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	}))
	manager.Get("/", indexAction)
	manager.Get("/status", statusAction)
}
