package irc

import (
	"crypto/tls"
	"errors"
	"fmt"
	stdlog "log"
	"time"

	log "github.com/Sirupsen/logrus"
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
	conf     *config.Configuration
	conn     *ircevent.Connection
	status   ConnectionStatus
	debug    bool
	lastPong time.Time
}

func NewConnectionManager(injector inject.Injector, conf *config.Configuration) *ConnectionManager {
	return &ConnectionManager{injector, conf, nil, Disconnected, false, time.Now()}
}

func (mgr *ConnectionManager) AutoConnect() {
	if mgr.conf.AutoConnect {
		mgr.Connect()
	}
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

func connect(mgr *ConnectionManager, l *log.Logger) {
	if mgr.conn == nil {
		mgr.conn = ircevent.IRC(mgr.conf.Nick, mgr.conf.Username)
		if mgr.conf.SASL {
			mgr.conn.UseSASL = true
			mgr.conn.SASLLogin = mgr.conf.SASLUsername
			mgr.conn.SASLPassword = mgr.conf.SASLPassword
		}
		mgr.conn.Log = stdlog.New(l.Writer(), "", 0)
		mgr.injector.Map(mgr.conn)
		mgr.injector.Invoke(bindEvents)
	}

	mgr.conn.Debug = mgr.debug
	mgr.conn.VerboseCallbackHandler = mgr.debug

	if mgr.conf.TLS {
		mgr.conn.UseTLS = true
		// TODO: Don't skip cert verification
		mgr.conn.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}

	mgr.lastPong = time.Now()
	mgr.conn.PingFreq = 2 * time.Minute
	mgr.status = Connecting
	mgr.injector.Invoke(triggerConnecting)
	mgr.conn.Connect(mgr.conf.Network)

	go func() {
		wait := 1 * time.Minute
		t := time.NewTimer(wait)
		for {
			select {
			case <-t.C:
				if mgr.conn == nil || mgr.status == Disconnected {
					return
				} else if time.Now().Sub(mgr.lastPong) > 5*time.Minute {
					l.Debugln("Ping Timeout, disconnecting.")
					mgr.Quit()
					mgr.AutoConnect()
					return
				}
				t.Reset(wait)
			default:
				time.Sleep(wait)
			}
		}
	}()
}
