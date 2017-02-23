package squircy2

//go:generate go-bindata -prefix "./web" -pkg web -tags "!debug" -o "./web/bindata.go" ./web/public/...
//go:generate stickgen -path "./web/views" -out web/generated index.html.twig
//go:generate stickgen -path "./web/views" -out web/generated **/[a-z]*.twig

import (
	"encoding/json"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/inject"
	_ "github.com/jteeuwen/go-bindata"
	"github.com/tyler-sommer/squircy2/config"
	"github.com/tyler-sommer/squircy2/data"
	"github.com/tyler-sommer/squircy2/event"
	"github.com/tyler-sommer/squircy2/eventsource"
	"github.com/tyler-sommer/squircy2/irc"
	"github.com/tyler-sommer/squircy2/script"
	"github.com/tyler-sommer/squircy2/web"
	"github.com/tyler-sommer/squircy2/webhook"
)

type Manager struct {
	inject.Injector
	s   *web.Server
	irc *irc.ConnectionManager

	conf *config.Configuration
}

func NewManager(rootPath string) *Manager {
	m := &Manager{Injector: inject.New()}
	m.Map(m)
	m.Map(m.Injector)

	conf := config.NewConfiguration(rootPath)
	database := data.NewDatabaseConnection(conf.RootPath)
	config.LoadConfig(database, conf)
	m.Map(conf)
	m.conf = conf

	evm := event.NewEventManager(m.Injector)
	m.Map(evm)
	m.invokeAndMap(event.NewTracer)
	l := newLogger(evm)
	m.Map(l)

	m.Map(database)
	m.invokeAndMap(script.NewScriptRepository)
	m.Map(webhook.NewWebhookRepository(database))

	m.s = web.NewServer(m.Injector, conf, l)
	m.invokeAndMap(newEventSource)
	m.irc = irc.NewConnectionManager(m.Injector, conf)
	m.Map(m.irc)
	m.invokeAndMap(script.NewScriptManager)

	return m
}

func (m *Manager) Conf() *config.Configuration {
	return m.conf
}

func (m *Manager) Web() *web.Server {
	return m.s
}

func (m *Manager) IRC() *irc.ConnectionManager {
	return m.irc
}

func (m *Manager) invokeAndMap(fn interface{}) interface{} {
	res, err := m.Invoke(fn)
	if err != nil {
		panic(err)
	}

	val := res[0].Interface()
	m.Map(val)

	return val
}

func newEventSource(evm event.EventManager, l log.FieldLogger) *eventsource.Broker {
	es := eventsource.New()
	es.FieldLogger = l

	var id int = -1
	evm.Bind(event.AllEvents, func(es *eventsource.Broker, ev event.Event) {
		id++
		d, _ := json.Marshal(ev.Data)
		es.Notify(&eventsource.Message{strconv.Itoa(id), string(ev.Type), string(d)})
	})

	return es
}
