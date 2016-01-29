package irc

import (
	"crypto/tls"
	"github.com/codegangsta/inject"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"log"
)

type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connecting
	Connected
)

type IrcConnectionManager struct {
	injector inject.Injector
	conn     *ircevent.Connection
	status   ConnectionStatus
}

func NewIrcConnectionManager(injector inject.Injector) (mgr *IrcConnectionManager) {
	mgr = &IrcConnectionManager{injector, nil, Disconnected}

	return
}

func (mgr *IrcConnectionManager) Connect() {
	mgr.injector.Invoke(connect)
}

func (mgr *IrcConnectionManager) Quit() {
	mgr.status = Disconnected
	if mgr.conn != nil && mgr.conn.Connected() {
		mgr.conn.Quit()
	}

	mgr.conn = nil
}

func (mgr *IrcConnectionManager) Status() ConnectionStatus {
	return mgr.status
}

func (mgr *IrcConnectionManager) Connection() *ircevent.Connection {
	return mgr.conn
}

func connect(mgr *IrcConnectionManager, conf *config.Configuration, l *log.Logger) {
	if mgr.conn == nil {
		mgr.conn = ircevent.IRC(conf.Nick, conf.Username)
		mgr.conn.Log = l
		mgr.injector.Map(mgr.conn)
		mgr.injector.Invoke(bindEvents)
	}

	if conf.TLS {
		mgr.conn.UseTLS = true
		mgr.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	mgr.status = Connecting
	mgr.injector.Invoke(triggerConnecting)
	mgr.conn.Connect(conf.Network)
}
