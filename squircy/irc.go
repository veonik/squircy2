package squircy

import (
	"github.com/thoj/go-ircevent"
	"fmt"
	"sync"
	handler "./irc"
)

type Handler interface {
	Id() string
	Matches(e *irc.Event) bool
	Handle(e *irc.Event)
}

type HandlerCollection struct {
	handlers map[string]Handler
}

func newHandlerCollection(conn *irc.Connection, config *Configuration) (c *HandlerCollection) {
	c = &HandlerCollection{make(map[string]Handler)}
	
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

	conn.AddCallback("001", func(e *irc.Event) { conn.Join(config.Channel) })

	conn.AddCallback("PRIVMSG", matchAndHandle)
	conn.AddCallback("NOTICE", matchAndHandle)
	
	c.Add(handler.NewNickservHandler(conn, config.Password))
		
	return
}

func newIrcConnection(config *Configuration) (conn *irc.Connection) {
	conn = irc.IRC(config.Nick, config.Username)

	err := conn.Connect(config.Network)
	if err != nil {
		panic(err)
	}
	
	return
}

func (c *HandlerCollection) Remove(h Handler) {
	if _, ok := c.handlers[h.Id()]; ok {
		fmt.Println("Removing handler ", h.Id())
		delete(c.handlers, h.Id())
	}
}

func (c *HandlerCollection) RemoveId(id string) {
	if handler, ok := c.handlers[id]; ok {
		c.Remove(handler)
	}
}

func (c *HandlerCollection) Add(h Handler) {
	fmt.Println("Adding handler ", h.Id())
	c.handlers[h.Id()] = h
}