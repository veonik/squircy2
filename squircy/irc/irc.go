package irc

import (
	"github.com/codegangsta/inject"
	"github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"log"
	"reflect"
)

type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connecting
	Connected
)

type IrcConnectionManager struct {
	injector inject.Injector
	conn     *irc.Connection
	status   ConnectionStatus
}

func NewIrcConnectionManager(injector inject.Injector) (mgr *IrcConnectionManager) {
	mgr = &IrcConnectionManager{injector, nil, Disconnected}

	return
}

func (mgr *IrcConnectionManager) newConnection() {
	res, _ := mgr.injector.Invoke(newIrcConnection)
	mgr.conn = res[0].Interface().(*irc.Connection)
	mgr.injector.Map(mgr.conn)
	mgr.injector.Invoke(bindEvents)
}

func (mgr *IrcConnectionManager) Connect() {
	if mgr.conn == nil {
		mgr.newConnection()
	}

	config := mgr.injector.Get(reflect.TypeOf((*config.Configuration)(nil))).Interface().(*config.Configuration)

	mgr.status = Connecting
	mgr.injector.Invoke(triggerConnecting)
	mgr.conn.Connect(config.Network)
}

func (mgr *IrcConnectionManager) Quit() {
	mgr.status = Disconnected
	if mgr.conn != nil {
		mgr.conn.Quit()
	}

	mgr.conn = nil
}

func (mgr *IrcConnectionManager) Status() ConnectionStatus {
	return mgr.status
}

func (mgr *IrcConnectionManager) Connection() *irc.Connection {
	return mgr.conn
}

func newIrcConnection(conf *config.Configuration, l *log.Logger) (conn *irc.Connection) {
	conn = irc.IRC(conf.Nick, conf.Username)
	conn.Log = l

	return
}
