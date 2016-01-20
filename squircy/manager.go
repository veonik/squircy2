package squircy

import (
	"encoding/json"
	"github.com/antage/eventsource"
	"github.com/go-martini/martini"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"strconv"
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

	manager.Map(database)
	manager.Map(script.NewScriptRepository(database))
	manager.Map(conf)

	manager.invokeAndMap(event.NewEventManager)
	manager.invokeAndMap(newEventTracer)
	manager.Invoke(configureLog)
	manager.Invoke(configureWeb)

	manager.invokeAndMap(newEventSource)
	manager.invokeAndMap(irc.NewIrcConnectionManager)
	manager.invokeAndMap(script.NewScriptManager)

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

func newEventSource(evm event.EventManager) eventsource.EventSource {
	es := eventsource.New(nil, nil)

	var id int = -1
	evm.Bind(event.AllEvents, func(es eventsource.EventSource, ev event.Event) {
		id++
		data, _ := json.Marshal(ev.Data)
		go es.SendEventMessage(string(data), string(ev.Type), strconv.Itoa(id))
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
