package script

import (
	"fmt"
	"github.com/tyler-sommer/squircy2/squircy/config"
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

type configHelper struct {
	conf *config.Configuration
}

func (h *configHelper) OwnerNick() string {
	return h.conf.OwnerNick
}

func (h *configHelper) OwnerHost() string {
	return h.conf.OwnerHost
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

func (h *ircHelper) CurrentNick() string {
	conn := h.manager.Connection()
	if conn == nil {
		return ""
	}
	return conn.GetNick()
}

func (h *ircHelper) Nick(newNick string) {
	conn := h.manager.Connection()
	if conn == nil {
		return
	}
	conn.Nick(newNick)
}

type scriptHelper struct {
	e          event.EventManager
	jsDriver   javascriptDriver
	handlers   map[string]event.EventHandler
}

func handlerId(scriptType ScriptType, eventType event.EventType, fnName string) string {
	return fmt.Sprintf("%v-%v-%v", scriptType, eventType, fnName)
}

// Bind adds a handler of the given script type for the given event type
func (s *scriptHelper) Bind(scriptType ScriptType, eventType event.EventType, fnName string) {
	id := handlerId(scriptType, eventType, fnName)
	var d scriptDriver
	switch {
	case scriptType == Javascript:
		d = s.jsDriver
	}

	handler := func(ev event.Event) {
		d.Handle(ev, fnName)
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

func (s *scriptHelper) Trigger(eventType event.EventType, data map[string]interface{}) {
	s.e.Trigger(eventType, data)
}
