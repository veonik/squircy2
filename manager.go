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
	_ "github.com/tyler-sommer/squircy2/web/module"
	"github.com/tyler-sommer/squircy2/webhook"
)

type Manager struct {
	inject.Injector

	conf   *config.Configuration
	logger *log.Logger

	scripts *script.ScriptManager
	web     *web.Server
	irc     *irc.ConnectionManager
	events  event.EventManager
}

func NewManager(rootPath string) *Manager {
	m := &Manager{Injector: inject.New()}

	m.conf = config.NewConfiguration(rootPath)
	database := data.NewDatabaseConnection(m.conf.RootPath)
	config.LoadConfig(database, m.conf)
	events := event.NewEventManager(m.Injector)
	logger := newLogger(events)
	scriptRepo := script.NewScriptRepository(database, m.conf, logger)

	m.web = web.NewServer(m.Injector, m.conf, logger)
	m.irc = irc.NewConnectionManager(logger, events, m.conf)
	m.scripts = script.NewScriptManager(scriptRepo, logger, events, m.irc, m.conf, database)
	m.logger = logger
	m.events = events

	m.Map(m)
	m.Map(m.conf)
	m.Map(m.irc)
	m.Map(m.web)
	m.Map(m.scripts)
	m.Map(m.logger)
	m.Map(m.events)

	m.Map(database)
	m.Map(event.NewTracer(events))
	m.Map(newEventSource(events, logger))
	m.Map(scriptRepo)
	m.Map(webhook.NewWebhookRepository(database))

	m.web.Configure()

	return m
}

func (m *Manager) Conf() *config.Configuration {
	return m.conf
}

func (m *Manager) Web() *web.Server {
	return m.web
}

func (m *Manager) IRC() *irc.ConnectionManager {
	return m.irc
}

func (m *Manager) Script() *script.ScriptManager {
	return m.scripts
}

func (m *Manager) Logger() *log.Logger {
	return m.logger
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
