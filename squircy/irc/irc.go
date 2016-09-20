package irc

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/codegangsta/inject"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/config"
)

var halt = errors.New("Halt")

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

func (mgr *IrcConnectionManager) Reconnect() {
	mgr.Quit()
	mgr.Connect()
}

func (mgr *IrcConnectionManager) Quit() {
	defer func() {
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Timeout waiting for disconnect.")
			} else {
				fmt.Println("Unexpected panic: ", err)
			}
		}
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

func (mgr *IrcConnectionManager) Status() ConnectionStatus {
	if mgr.conn == nil || !mgr.conn.Connected() {
		mgr.status = Disconnected
	}
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
