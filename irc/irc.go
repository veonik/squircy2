package irc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/codegangsta/inject"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/config"
)

var halt = errors.New("Halt")

type ConnectionStatus int

const (
	Disconnected ConnectionStatus = iota
	Connecting
	Connected
)

type ConnectionManager struct {
	injector inject.Injector
	conn     *ircevent.Connection
	status   ConnectionStatus
	debug    bool
}

func NewIrcConnectionManager(injector inject.Injector) *ConnectionManager {
	return &ConnectionManager{injector, nil, Disconnected, false}
}

func (mgr *ConnectionManager) Connect() {
	mgr.injector.Invoke(connect)
}

func (mgr *ConnectionManager) Reconnect() {
	mgr.Quit()
	mgr.Connect()
}

func (mgr *ConnectionManager) Debug() bool {
	return mgr.debug
}

func (mgr *ConnectionManager) SetDebug(debug bool) {
	mgr.debug = debug
	if mgr.conn != nil {
		mgr.conn.Debug = mgr.debug
		mgr.conn.VerboseCallbackHandler = mgr.debug
	}
}

func (mgr *ConnectionManager) Quit() {
	defer func() {
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Timeout waiting for disconnect.")
			} else {
				fmt.Println("Unexpected panic: ", err)
			}
		}
		mgr.injector.Invoke(triggerDisconnected)
		mgr.status = Disconnected
		mgr.conn = nil
	}()
	if mgr.conn != nil && mgr.conn.Connected() {
		d := make(chan struct{})
		go func() {
			mgr.conn.Quit()
			mgr.conn.Disconnect()
			close(d)
		}()
		time.Sleep(1 * time.Second)
		select {
		case <-d:
			return

		default:
			panic(halt)
		}
	}
}

func (mgr *ConnectionManager) Status() ConnectionStatus {
	if mgr.conn == nil || !mgr.conn.Connected() {
		mgr.status = Disconnected
	}
	return mgr.status
}

func (mgr *ConnectionManager) Connection() *ircevent.Connection {
	return mgr.conn
}

func connect(mgr *ConnectionManager, conf *config.Configuration, l *log.Logger) {
	if mgr.conn == nil {
		mgr.conn = ircevent.IRC(conf.Nick, conf.Username)
		mgr.conn.Log = l
		mgr.injector.Map(mgr.conn)
		mgr.injector.Invoke(bindEvents)
	}

	mgr.conn.Debug = mgr.debug
	mgr.conn.VerboseCallbackHandler = mgr.debug

	if conf.TLS {
		mgr.conn.UseTLS = true
		// TODO: Don't skip cert verification
		mgr.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	mgr.status = Connecting
	mgr.injector.Invoke(triggerConnecting)
	mgr.conn.Connect(conf.Network)
}
