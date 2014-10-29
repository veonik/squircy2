package squircy

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"io"
	"log"
	"os"
	"strings"
)

type Manager struct {
	*martini.ClassicMartini
}

type logHistory struct {
	limit int
	data  []string
}

func (hist *logHistory) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil

	if len(hist.data) >= hist.limit {
		hist.data = hist.data[1:]
	}

	hist.data = append(hist.data, string(p))

	return
}

func (hist *logHistory) ReadAll() string {
	return strings.Join(hist.data, "")
}

func NewManager() (manager *Manager) {
	manager = &Manager{martini.Classic()}
	manager.Map(manager)
	manager.Map(manager.Injector)

	conf := config.NewDefaultConfiguration()
	database := data.NewDatabaseConnection(conf.RootPath)
	config.LoadConfig(database, conf)

	hist := &logHistory{25, make([]string, 0)}
	out := io.MultiWriter(os.Stdout, hist)

	manager.Map(log.New(out, "[squircy] ", 0))
	manager.Map(hist)

	manager.Map(database)
	manager.Map(script.NewScriptRepository(database))
	manager.Map(conf)

	// Additional managers
	manager.invokeAndMap(event.NewEventManager)
	manager.invokeAndMap(irc.NewIrcConnectionManager)
	manager.invokeAndMap(script.NewScriptManager)

	manager.configure(conf)

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

func (manager *Manager) configure(conf *config.Configuration) {
	manager.Handlers(
		martini.Logger(),
		martini.Static(conf.RootPath+"/public", martini.StaticOptions{
			SkipLogging: false,
		}),
		render.Renderer(render.Options{
			Directory:  conf.RootPath + "/views",
			Layout:     "layout",
			Extensions: []string{".tmpl", ".html"},
		}))
	manager.Get("/", indexAction)
	manager.Get("/status", statusAction)
	manager.Group("/manage", func(r martini.Router) {
		r.Get("", manageAction)
		r.Post("/update", manageUpdateAction)
	})
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
	})
	manager.Group("/repl", func(r martini.Router) {
		r.Get("", replAction)
		r.Post("/execute", replExecuteAction)
	})
}
