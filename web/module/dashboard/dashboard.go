package dashboard

import (
	"errors"
	"net/http"

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tyler-sommer/stick"
	"github.com/veonik/squircy2/event"
	"github.com/veonik/squircy2/eventsource"
	"github.com/veonik/squircy2/irc"
	"github.com/veonik/squircy2/web"
)

func init() {
	web.MustRegister(NewWithInjector)
}

type module struct {
	tracer  *event.Tracer
	broker  *eventsource.Broker
	manager *irc.ConnectionManager
}

func NewWithInjector(injector inject.Injector) (web.Module, error) {
	res, err := injector.Invoke(New)
	if err != nil {
		return nil, err
	}
	if m, ok := res[0].Interface().(web.Module); ok {
		return m, nil
	}
	return nil, errors.New("dashboard: unable to create web module")
}

func New(tracer *event.Tracer, broker *eventsource.Broker, manager *irc.ConnectionManager) *module {
	return &module{tracer, broker, manager}
}

func (m *module) Configure(s *web.Server) error {
	s.Group("", func(r martini.Router) {
		// TODO: Move to own module
		r.Get("/metrics", promhttp.Handler())
		r.Get("/event", func(w http.ResponseWriter, r *http.Request) {
			m.broker.ServeHTTP(w, r)
		})
		r.Get("/", m.indexAction)
		r.Get("/status", m.statusAction)
		r.Post("/connect", m.connectAction)
		r.Post("/disconnect", m.disconnectAction)
	})
	return nil
}

func (m *module) indexAction(s *web.StickHandler) {
	s.HTML(200, "index.html.twig", map[string]stick.Value{
		"log": m.tracer.History(event.EventType("log.OUTPUT")),
		"irc": m.tracer.History(irc.IrcEvent),
	})
}

func (m *module) statusAction(r render.Render) {
	r.JSON(200, struct{ Status irc.ConnectionStatus }{m.manager.Status()})
}

func (m *module) connectAction(r render.Render) {
	m.manager.Connect()

	r.JSON(200, nil)
}

func (m *module) disconnectAction(r render.Render) {
	m.manager.Quit()

	r.JSON(200, nil)
}
