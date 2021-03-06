package squircy2 // import "github.com/veonik/squircy2"

//go:generate go-bindata -prefix "./web" -pkg generated -tags "!debug" -modtime 0 -o "./web/generated/bindata.go" ./web/public/...
//go:generate stickgen -path "./web/views" -out web/generated index.html.twig
//go:generate stickgen -path "./web/views" -out web/generated **/[a-z]*.twig

import (
	"encoding/json"
	"strconv"

	"github.com/veonik/squircy2/plugin"

	"github.com/codegangsta/inject"
	_ "github.com/jteeuwen/go-bindata"
	log "github.com/sirupsen/logrus"
	"github.com/veonik/squircy2/config"
	"github.com/veonik/squircy2/data"
	"github.com/veonik/squircy2/event"
	"github.com/veonik/squircy2/eventsource"
	"github.com/veonik/squircy2/irc"
	"github.com/veonik/squircy2/script"
	"github.com/veonik/squircy2/web"
	_ "github.com/veonik/squircy2/web/module"
	"github.com/veonik/squircy2/webhook"
)

type Manager struct {
	inject.Injector

	conf   *config.Configuration
	logger *log.Logger

	plugins *plugin.Manager
	scripts *script.ScriptManager
	web     *web.Server
	irc     *irc.ConnectionManager
	events  event.EventManager
}

func NewManager(rootPath string) *Manager {
	m := &Manager{Injector: inject.New()}

	m.conf = config.NewConfiguration(rootPath)

	events := event.NewEventManager(m.Injector)
	logger := newLogger(events)
	database := data.NewDatabaseConnection(m.conf.RootPath, logger)

	config.LoadConfig(database, logger, m.conf)
	scriptRepo := script.NewScriptRepository(database, m.conf, logger)

	m.web = web.NewServer(m.Injector, m.conf, logger)
	m.irc = irc.NewConnectionManager(logger, events, m.conf)
	m.scripts = script.NewScriptManager(scriptRepo, logger, events, m.irc, m.conf, database)
	m.logger = logger
	m.events = events
	m.plugins = plugin.NewManager(m.conf.PluginsPath, m.conf.PluginsEnabled...)

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

	if errs := m.plugins.Configure(m.Injector); len(errs) > 0 {
		for _, err := range errs {
			logger.Warnln(err)
		}
	}
	m.web.Configure()

	m.scripts.ReInit()

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
