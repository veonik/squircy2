package squircy

import (
	"encoding/json"
	"github.com/antage/eventsource"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Manager struct {
	*martini.ClassicMartini
}

func NewManager() (manager *Manager) {
	manager = &Manager{martini.Classic()}
	manager.Map(manager)
	manager.Map(manager.Injector)

	conf := config.NewDefaultConfiguration()
	database := data.NewDatabaseConnection(conf.RootPath)
	config.LoadConfig(database, conf)

	hist := &limitedLogger{25, make([]string, 0)}
	out := io.MultiWriter(os.Stdout, hist)
	logger := log.New(out, "", log.Ltime)
	manager.Map(logger)
	manager.Map(hist)

	manager.Map(database)
	manager.Map(script.NewScriptRepository(database))
	manager.Map(conf)

	// Additional managers
	manager.invokeAndMap(event.NewEventManager)
	manager.invokeAndMap(newEventSource)
	manager.invokeAndMap(irc.NewIrcConnectionManager)
	manager.invokeAndMap(script.NewScriptManager)

	manager.configure(conf)

	return
}

func newEventSource(evm event.EventManager) eventsource.EventSource {
	es := eventsource.New(nil, nil)

	var id int = -1
	evm.Bind(event.AllEvents, func(es eventsource.EventSource, ev event.Event) {
		id++
		data, _ := json.Marshal(ev.Data["Event"])
		go es.SendEventMessage(string(data), string(ev.Type), strconv.Itoa(id))
	})

	return es
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
		martini.Static(conf.RootPath+"/public", martini.StaticOptions{
			SkipLogging: true,
		}),
		render.Renderer(render.Options{
			Directory:  conf.RootPath + "/views",
			Layout:     "layout",
			Extensions: []string{".tmpl", ".html"},
		}))
	manager.Get("/event", func(es eventsource.EventSource, w http.ResponseWriter, r *http.Request) {
		es.ServeHTTP(w, r)
	})
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

type limitedLogger struct {
	limit int
	data  []string
}

func (l *limitedLogger) Write(p []byte) (n int, err error) {
	n = len(p)
	err = nil

	if len(l.data) >= l.limit {
		l.data = l.data[1:]
	}

	l.data = append(l.data, string(p))

	return
}

func (hist *limitedLogger) ReadAll() string {
	return strings.Join(hist.data, "")
}
