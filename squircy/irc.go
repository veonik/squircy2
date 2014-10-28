package squircy

import (
	"github.com/thoj/go-ircevent"
	"log"
	"reflect"
	"sync"
)

type IrcConnectionManager struct {
	manager    *Manager
	conn       *irc.Connection
	connected  bool
	connecting bool
}

func newIrcConnectionManager(manager *Manager) (mgr *IrcConnectionManager) {
	mgr = &IrcConnectionManager{manager, nil, false, false}

	res, _ := mgr.manager.Invoke(newIrcConnection)
	mgr.conn = res[0].Interface().(*irc.Connection)

	mgr.manager.Map(mgr.conn)

	return
}

func (mgr *IrcConnectionManager) Connect() {
	if mgr.conn == nil {
		mgr.conn = mgr.manager.invokeAndMap(newIrcConnection).(*irc.Connection)
	}

	config := mgr.manager.Injector.Get(reflect.TypeOf((*Configuration)(nil))).Interface().(*Configuration)
	l := mgr.manager.Injector.Get(reflect.TypeOf((*log.Logger)(nil))).Interface().(*log.Logger)

	h := mgr.manager.invokeAndMap(newHandlerCollection).(*HandlerCollection)
	scriptHandler := mgr.manager.invokeAndMap(newScriptHandler).(*ScriptHandler)

	h.Add(scriptHandler)

	mgr.conn.AddCallback("001", func(e *irc.Event) {
		l.Println("Connected")
		mgr.connecting = false
		mgr.connected = true
	})

	mgr.conn.AddCallback("ERROR", func(e *irc.Event) {
		l.Println("Disconnected")
		if mgr.Connected() {
			mgr.Quit()
		}
	})

	h.bind(mgr.conn)

	mgr.connecting = true
	mgr.conn.Connect(config.Network)
}

func (mgr *IrcConnectionManager) Quit() {
	mgr.connecting = false
	mgr.connected = false
	if mgr.conn != nil {
		mgr.conn.Quit()
	}

	mgr.conn = nil
}

func (mgr *IrcConnectionManager) Connected() bool {
	return mgr.connected
}

func (mgr *IrcConnectionManager) Connecting() bool {
	return mgr.connecting
}

type Handler interface {
	Id() string
	Matches(e *irc.Event) bool
	Handle(e *irc.Event)
}

type HandlerCollection struct {
	handlers map[string]Handler
	log      *log.Logger
}

func newHandlerCollection(config *Configuration, l *log.Logger) (c *HandlerCollection) {
	c = &HandlerCollection{make(map[string]Handler), l}

	return
}

func (c *HandlerCollection) bind(conn *irc.Connection) {
	mutex := &sync.Mutex{}
	matchAndHandle := func(e *irc.Event) {
		mutex.Lock()
		for _, h := range c.handlers {
			if h.Matches(e) {
				h.Handle(e)
			}
		}
		mutex.Unlock()
	}

	conn.AddCallback("*", matchAndHandle)
}

func newIrcConnection(config *Configuration, l *log.Logger) (conn *irc.Connection) {
	conn = irc.IRC(config.Nick, config.Username)
	conn.Log = l

	return
}

func (c *HandlerCollection) Remove(h Handler) {
	if _, ok := c.handlers[h.Id()]; ok {
		c.log.Println("Removing handler", h.Id())
		delete(c.handlers, h.Id())
	}
}

func (c *HandlerCollection) RemoveId(id string) {
	if handler, ok := c.handlers[id]; ok {
		c.Remove(handler)
	}
}

func (c *HandlerCollection) Add(h Handler) {
	c.log.Println("Adding handler", h.Id())
	c.handlers[h.Id()] = h
}
