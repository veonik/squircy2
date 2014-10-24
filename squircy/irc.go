package squircy

import (
	"github.com/thoj/go-ircevent"
	"log"
	"sync"
)

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
