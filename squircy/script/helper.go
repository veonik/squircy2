package script

import (
	"fmt"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
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
	manager *irc.IrcConnectionManager
}

func (h *ircHelper) Privmsg(target, message string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Privmsg(target, message)
}

func (h *ircHelper) Join(target string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Join(target)
}

func (h *ircHelper) Part(target string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Part(target)
}

type scriptHelper struct {
	e          event.EventManager
	jsDriver   javascriptDriver
	luaDriver  luaDriver
	lispDriver lispDriver
	ankoDriver ankoDriver
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

		case scriptType == Anko:
			s.ankoDriver.Handle(ev, fnName)
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
