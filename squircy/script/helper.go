package script

import (
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"io/ioutil"
	"net/http"
)

type httpHelper struct{}

func (client *httpHelper) Get(url string) string {
	resp, err := http.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

type dataHelper struct {
	d map[string]interface{}
}

func (db *dataHelper) Get(key string) interface{} {
	if val, ok := db.d[key]; ok {
		return val
	}

	return nil
}

func (db *dataHelper) Set(key string, val interface{}) {
	db.d[key] = val
}

type ircHelper struct {
	conn *ircevent.Connection
}

func (irc *ircHelper) Privmsg(target, message string) {
	if irc.conn == nil {
		return
	}
	irc.conn.Privmsg(target, message)
}

func (irc *ircHelper) Join(target string) {
	if irc.conn == nil {
		return
	}
	irc.conn.Join(target)
}

func (irc *ircHelper) Part(target string) {
	if irc.conn == nil {
		return
	}
	irc.conn.Part(target)
}

type scriptHelper struct {
	e          event.EventManager
	jsDriver   javascriptDriver
	luaDriver  luaDriver
	lispDriver lispDriver
	handlers   map[string]event.EventHandler
}

func handlerId(scriptType ScriptType, eventType event.EventType, fnName string) string {
	return fmt.Sprintf("%v-%v-%v", scriptType, eventType, fnName)
}

// Bind adds a handler of the given script type for the given event type
func (s *scriptHelper) Bind(scriptType ScriptType, eventType event.EventType, fnName string) {
	id := handlerId(scriptType, eventType, fnName)
	handler := func(ev event.Event) {
		switch {
		case scriptType == Javascript:
			s.jsDriver.Handle(ev, fnName)

		case scriptType == Lua:
			s.luaDriver.Handle(ev, fnName)

		case scriptType == Lisp:
			s.lispDriver.Handle(ev, fnName)
		}
	}
	s.handlers[id] = handler
	s.e.Bind(eventType, handler)
}

// Unbind removes a handler of the given script type for the given event type
func (s *scriptHelper) Unbind(scriptType ScriptType, eventType event.EventType, fnName string) {
	id := handlerId(scriptType, eventType, fnName)
	handler, ok := s.handlers[id]
	if !ok {
		return
	}
	s.e.Unbind(eventType, handler)
	delete(s.handlers, id)
}
