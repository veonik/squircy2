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

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	log "github.com/sirupsen/logrus"
	"github.com/tyler-sommer/squircy2/config"
)

type Server struct {
	*martini.ClassicMartini

	l    *log.Logger
	conf *config.Configuration

	httpListener  io.Closer
	httpsListener io.Closer
}

func NewServer(injector inject.Injector, conf *config.Configuration, l *log.Logger) *Server {
	s := &Server{
		ClassicMartini: newCustomMartini(injector, l),
		l:              l,
		conf:           conf,
	}
	return s
}

func (s *Server) ListenAndServe() {
	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				s.l.Errorln("Could not start HTTP: ", err)
			}
			wg.Done()
		}()
		err := listenAndServe(s)
		if err != nil {
			panic(err)
		}
	}()
	go func() {
		defer func() {
			if err, ok := recover().(error); ok {
				s.l.Errorln("Could not start HTTPS: ", err)
			}
			wg.Done()
		}()
		err := listenAndServeTLS(s)
		if err != nil {
			panic(err)
		}
	}()
	wg.Wait()
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

func listenAndServe(s *Server) error {
	if !s.conf.WebInterface {
		return errors.New("web: web interface is disabled.")
	}
	s.l.Infoln("Starting HTTP, listening at", s.conf.HTTPHostPort)
	srv := &http.Server{Addr: s.conf.HTTPHostPort, Handler: s}
	listener, err := net.Listen("tcp", s.conf.HTTPHostPort)
	if err != nil {
		return err
	}
	go func() {
		err := srv.Serve(tcpKeepAliveListener{listener.(*net.TCPListener)})
		if err != nil {
			s.l.Errorln(err.Error())
		}
	}()
	s.httpListener = listener.(io.Closer)
	return nil
}

func listenAndServeTLS(s *Server) error {
	if !s.conf.WebInterface {
		return errors.New("web: web interface is disabled.")
	}
	if !s.conf.HTTPS {
		return errors.New("HTTPS is disabled.")
	}
	s.l.Infoln("Starting HTTPS, listening at", s.conf.SSLHostPort)
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
			s.l.Errorln(err.Error())
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
