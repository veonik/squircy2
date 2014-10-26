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
	config := NewDefaultConfiguration()
	db := newDatabaseConnection(config)
	loadConfig(db, config)

	manager = &Manager{martini.Classic()}
	manager.Map(db)
	manager.Map(config)
	manager.Map(log.New(os.Stdout, "[squircy] ", 0))
	manager.invokeAndMap(newIrcConnection)
	manager.Map(scriptRepository{db})
	h := manager.invokeAndMap(newHandlerCollection).(*HandlerCollection)
	nickservHandler := manager.invokeAndMap(newNickservHandler).(*NickservHandler)
	scriptHandler := manager.invokeAndMap(newScriptHandler).(*ScriptHandler)

	h.Add(nickservHandler)
	h.Add(scriptHandler)

	manager.configure(config)

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

func (manager *Manager) configure(config *Configuration) {
	manager.Use(martini.Static(config.RootPath + "/public"))
	manager.Use(render.Renderer(render.Options{
		Directory:  config.RootPath + "/views",
		Layout:     "layout",
		Extensions: []string{".tmpl", ".html"},
	}))
	manager.Get("/", indexAction)
	manager.Get("/status", statusAction)
	manager.Get("/manage", manageAction)
	manager.Post("/manage/update", manageUpdateAction)
	manager.Post("/connect", connectAction)
	manager.Post("/disconnect", disconnectAction)
	manager.Group("/script", func(r martini.Router) {
		r.Get("", scriptAction)
		r.Post("/reinit", scriptReinitAction)
		r.Get("/new", newScriptAction)
		r.Post("/create", createScriptAction)
		r.Get("/:id/edit", editScriptAction)
		r.Post("/:id/update", updateScriptAction)
		r.Post("/:id/remove", removeScriptAction)
		r.Get("/:id/execute", executeScriptAction)
	})
	manager.Group("/repl", func(r martini.Router) {
		r.Get("", replAction)
		r.Post("/execute", replExecuteAction)
	})
}
