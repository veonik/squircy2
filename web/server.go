package web

import (
	"crypto/tls"
	"errors"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/secure"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/tyler-sommer/squircy2/config"
	"github.com/tyler-sommer/squircy2/eventsource"
)

type Server struct {
	*martini.ClassicMartini

	conf *config.Configuration

	httpListener  io.Closer
	httpsListener io.Closer
}

func NewServer(injector inject.Injector, conf *config.Configuration, l *log.Logger) *Server {
	s := &Server{
		ClassicMartini: newCustomMartini(injector, l),
		conf:           conf,
	}
	configure(s)
	return s
}

func (s *Server) ListenAndServe() {
	s.Invoke(s.listenAndServe)
}

func (s *Server) StopListenAndServe() error {
	if s.httpListener == nil {
		return errors.New("web: not listening on http")
	}
	defer func() {
		s.httpListener = nil
	}()
	return s.httpListener.Close()
}

func (s *Server) StopListenAndServeTLS() error {
	if s.httpsListener == nil {
		return errors.New("web: not listening on https")
	}
	defer func() {
		s.httpsListener = nil
	}()
	return s.httpsListener.Close()
}

func (s *Server) listenAndServe(l log.FieldLogger) {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				l.Errorln("Could not start HTTP: ", err)
			}
			wg.Done()
		}()
		err := listenAndServe(s, l)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				l.Errorln("Could not start HTTPS: ", err)
			}
			wg.Done()
		}()
		err := listenAndServeTLS(s, l)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
}

func listenAndServe(s *Server, l log.FieldLogger) error {
	if !s.conf.WebInterface {
		return errors.New("web: web interface is disabled.")
	}
	l.Infoln("Starting HTTP, listening at", s.conf.HTTPHostPort)
	srv := &http.Server{Addr: s.conf.HTTPHostPort, Handler: s}
	listener, err := net.Listen("tcp", s.conf.HTTPHostPort)
	if err != nil {
		return err
	}
	go func() {
		err := srv.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)})
		if err != nil {
			l.Errorln(err.Error())
		}
	}()
	s.httpListener = listener.(io.Closer)
	return nil
}

func listenAndServeTLS(s *Server, l log.FieldLogger) error {
	if !s.conf.WebInterface {
		return errors.New("web: web interface is disabled.")
	}
	if !s.conf.HTTPS {
		return errors.New("HTTPS is disabled.")
	}
	l.Infoln("Starting HTTPS, listening at", s.conf.SSLHostPort)
	srv := &http.Server{Addr: s.conf.SSLHostPort, Handler: s}
	var err error
	srv.TLSConfig = &tls.Config{}
	srv.TLSConfig.NextProtos = []string{"http", "h2"}
	srv.TLSConfig.Certificates = make([]tls.Certificate, 1)
	srv.TLSConfig.Certificates[0], err = tls.LoadX509KeyPair(s.conf.SSLCertFile, s.conf.SSLCertKey)
	if err != nil {
		return err
	}
	listener, err := net.Listen("tcp", s.conf.SSLHostPort)
	if err != nil {
		return err
	}
	go func() {
		err := srv.Serve(tls.NewListener(tcpKeepAliveListener{listener.(*net.TCPListener)}, srv.TLSConfig))
		if err != nil {
			l.Errorln(err.Error())
		}
	}()
	s.httpsListener = listener.(io.Closer)
	return nil
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

func newCustomMartini(injector inject.Injector, l *log.Logger) *martini.ClassicMartini {
	r := martini.NewRouter()
	m := martini.New()
	m.Injector = injector
	m.Logger(stdlog.New(l.Writer(), "", 0))
	m.Use(martini.Logger())
	m.Use(martini.Recovery())
	m.Use(martini.Static("web/public"))
	m.MapTo(r, (*martini.Routes)(nil))
	m.Action(r.Handle)
	return &martini.ClassicMartini{m, r}
}

func configure(s *Server) {
	s.Handlers(
		counterHandler,
		newStaticHandler(),
		newStickHandler(),
		render.Renderer(),
		secure.Secure(secure.Options{
			BrowserXssFilter: true,
			FrameDeny:        true,
			SSLRedirect:      s.conf.RequireHTTPS,
			SSLHost:          s.conf.SSLHostPort,
			DisableProdCheck: true,
		}),
	)
	s.NotFound(func(req *http.Request, r render.Render, l log.FieldLogger) {
		r.Error(404)
	})

	s.Post("/webhooks/:webhook_id", webhookReceiveAction)

	// Admin web interface
	handlers := []martini.Handler{}
	if s.conf.HTTPAuth && len(s.conf.AuthUsername) > 0 && len(s.conf.AuthPassword) > 0 {
		handlers = append(handlers, auth.Basic(s.conf.AuthUsername, s.conf.AuthPassword))
	}
	s.Group("", func(rm martini.Router) {
		rm.Get("/metrics", promhttp.Handler())
		rm.Get("/event", func(es *eventsource.Broker, w http.ResponseWriter, r *http.Request) {
			es.ServeHTTP(w, r)
		})
		rm.Get("/", indexAction)
		rm.Get("/status", statusAction)
		rm.Group("/manage", func(r martini.Router) {
			r.Get("", manageAction)
			r.Post("/update", manageUpdateAction)
			r.Post("/export-scripts", manageExportScriptsAction)
			r.Post("/import-scripts", manageImportScriptsAction)
		})
		rm.Post("/connect", connectAction)
		rm.Post("/disconnect", disconnectAction)
		rm.Group("/script", func(r martini.Router) {
			r.Get("", scriptAction)
			r.Post("/reinit", scriptReinitAction)
			r.Get("/new", newScriptAction)
			r.Post("/create", createScriptAction)
			r.Get("/:id/edit", editScriptAction)
			r.Post("/:id/update", updateScriptAction)
			r.Post("/:id/remove", removeScriptAction)
			r.Post("/:id/toggle", toggleScriptAction)
		})
		rm.Group("/repl", func(r martini.Router) {
			r.Get("", replAction)
			r.Post("/execute", replExecuteAction)
		})
		rm.Group("/webhook", func(r martini.Router) {
			r.Get("", webhookAction)
			r.Get("/new", newWebhookAction)
			r.Post("/create", createWebhookAction)
			r.Get("/:id/edit", editWebhookAction)
			r.Post("/:id/update", updateWebhookAction)
			r.Post("/:id/remove", removeWebhookAction)
			r.Post("/:id/toggle", toggleWebhookAction)
		})
	}, handlers...)
}
