package squircy

//go:generate go-bindata -prefix "../" -pkg squircy -tags "!debug" ../views/... ../public/...

import (
	"encoding/json"
	"strconv"

	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/go-martini/martini"
	_ "github.com/jteeuwen/go-bindata"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/eventsource"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"github.com/tyler-sommer/squircy2/squircy/webhook"
)

type Manager struct {
	*martini.ClassicMartini
}

func NewManager(rootPath string) (manager *Manager) {
	manager = &Manager{martini.Classic()}
	manager.Map(manager)
	manager.Map(manager.Injector)

	conf := config.NewConfiguration(rootPath)
	database := data.NewDatabaseConnection(conf.RootPath)
	config.LoadConfig(database, conf)
	manager.Map(conf)

	manager.invokeAndMap(event.NewEventManager)
	manager.invokeAndMap(newEventTracer)
	manager.Invoke(configureLog)

	manager.Map(database)
	manager.invokeAndMap(script.NewScriptRepository)
	manager.Map(webhook.NewWebhookRepository(database))

	manager.Invoke(configureWeb)
	manager.invokeAndMap(newEventSource)
	manager.invokeAndMap(irc.NewIrcConnectionManager)
	manager.invokeAndMap(script.NewScriptManager)

	return
}

func (manager *Manager) ListenAndServe() {
	manager.Invoke(manager.listenAndServe)
}

func (manager *Manager) listenAndServe(conf *config.Configuration, l *log.Logger) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				l.Println("Could not start HTTP: ", err)
			}
			wg.Done()
		}()
		err := listenAndServe(manager, conf, l)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				l.Println("Could not start HTTPS: ", err)
			}
			wg.Done()
		}()
		err := listenAndServeTLS(manager, conf, l)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}

func (manager *Manager) AutoConnect() {
	manager.Invoke(manager.autoConnect)
}

func (manager *Manager) autoConnect(conf *config.Configuration, ircmgr *irc.IrcConnectionManager) {
	if conf.AutoConnect {
		ircmgr.Connect()
	}
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

func listenAndServe(manager *Manager, conf *config.Configuration, l *log.Logger) error {
	if !conf.WebInterface {
		return errors.New("Web Interface is disabled.")
	}
	l.Println("Starting HTTP, listening at", conf.HTTPHostPort)
	return http.ListenAndServe(conf.HTTPHostPort, manager)
}

func listenAndServeTLS(manager *Manager, conf *config.Configuration, l *log.Logger) error {
	if !conf.WebInterface {
		return errors.New("Web Interface is disabled.")
	}
	if !conf.HTTPS {
		return errors.New("HTTPS is disabled.")
	}
	l.Println("Starting HTTPS, listening at", conf.SSLHostPort)
	return http.ListenAndServeTLS(conf.SSLHostPort, conf.SSLCertFile, conf.SSLCertKey, manager)
}

func newEventSource(evm event.EventManager) *eventsource.Broker {
	es := eventsource.New()

	var id int = -1
	evm.Bind(event.AllEvents, func(es *eventsource.Broker, ev event.Event) {
		id++
		d, _ := json.Marshal(ev.Data)
		es.Notify(&eventsource.Message{strconv.Itoa(id), string(ev.Type), string(d)})
	})

	return es
}

type eventTracer struct {
	limit int
	data  map[event.EventType][]map[string]interface{}
}

func newEventTracer(evm event.EventManager) *eventTracer {
	t := &eventTracer{25, make(map[event.EventType][]map[string]interface{}, 0)}
	evm.Bind(event.AllEvents, func(ev event.Event) {
		history, ok := t.data[ev.Type]
		if !ok {
			history = make([]map[string]interface{}, 0)
		}

		if len(history) >= t.limit {
			history = history[1:]
		}

		history = append(history, ev.Data)

		t.data[ev.Type] = history
	})

	return t
}

func (t *eventTracer) History(evt event.EventType) []map[string]interface{} {
	if history, ok := t.data[evt]; ok {
		return history
	}

	return nil
}
