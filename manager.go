package squircy2

//go:generate go-bindata -prefix "./" -pkg squircy2 -tags "!debug" ./public/...
//go:generate stickgen -path "./views" -out generated index.html.twig
//go:generate stickgen -path "./views" -out generated **/[a-z]*.twig

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-martini/martini"
	_ "github.com/jteeuwen/go-bindata"
	"github.com/tyler-sommer/squircy2/config"
	"github.com/tyler-sommer/squircy2/data"
	"github.com/tyler-sommer/squircy2/event"
	"github.com/tyler-sommer/squircy2/eventsource"
	"github.com/tyler-sommer/squircy2/irc"
	"github.com/tyler-sommer/squircy2/script"
	"github.com/tyler-sommer/squircy2/webhook"
)

type Manager struct {
	*martini.ClassicMartini

	httpListener  io.Closer
	httpsListener io.Closer
}

func NewManager(rootPath string) (manager *Manager) {
	manager = &Manager{ClassicMartini: martini.Classic()}
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

func (manager *Manager) StopListenAndServe() error {
	if manager.httpListener == nil {
		return nil
	}
	defer func() {
		manager.httpListener = nil
	}()
	return manager.httpListener.Close()
}

func (manager *Manager) StopListenAndServeTLS() error {
	if manager.httpsListener == nil {
		return nil
	}
	defer func() {
		manager.httpsListener = nil
	}()
	return manager.httpsListener.Close()
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

func (manager *Manager) autoConnect(conf *config.Configuration, ircmgr *irc.ConnectionManager) {
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

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
//
// Taken from stdlib net/http.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func listenAndServe(manager *Manager, conf *config.Configuration, l *log.Logger) error {
	if !conf.WebInterface {
		return errors.New("Web Interface is disabled.")
	}
	l.Println("Starting HTTP, listening at", conf.HTTPHostPort)
	s := &http.Server{Addr: conf.HTTPHostPort, Handler: manager}
	listener, err := net.Listen("tcp", conf.HTTPHostPort)
	if err != nil {
		return err
	}
	go func() {
		err := s.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)})
		if err != nil {
			l.Println(err.Error())
		}
	}()
	manager.httpListener = listener.(io.Closer)
	return nil
}

func listenAndServeTLS(manager *Manager, conf *config.Configuration, l *log.Logger) error {
	if !conf.WebInterface {
		return errors.New("Web Interface is disabled.")
	}
	if !conf.HTTPS {
		return errors.New("HTTPS is disabled.")
	}
	l.Println("Starting HTTPS, listening at", conf.SSLHostPort)
	s := &http.Server{Addr: conf.SSLHostPort, Handler: manager}
	var err error
	s.TLSConfig = &tls.Config{}
	s.TLSConfig.NextProtos = []string{"http", "h2"}
	s.TLSConfig.Certificates = make([]tls.Certificate, 1)
	s.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(conf.SSLCertFile, conf.SSLCertKey)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", conf.SSLHostPort)
	if err != nil {
		return err
	}
	go func() {
		err := s.Serve(tls.NewListener(tcpKeepAliveListener{listener.(*net.TCPListener)}, s.TLSConfig))
		if err != nil {
			l.Println(err.Error())
		}
	}()
	manager.httpsListener = listener.(io.Closer)
	return nil
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
