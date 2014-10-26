package squircy

import (
	"github.com/thoj/go-ircevent"
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
	conn *irc.Connection
}

func (irc *ircHelper) Privmsg(target, message string) {
	irc.conn.Privmsg(target, message)
}

func (irc *ircHelper) Join(target string) {
	irc.conn.Join(target)
}

func (irc *ircHelper) Part(target string) {
	irc.conn.Part(target)
}

const (
	handlerJs   string = "js"
	handlerLua         = "lua"
	handlerLisp        = "lisp"
)

type scriptHelper struct {
	handler *ScriptHandler
}

// On adds a handler of the given script type for the given event type
func (script *scriptHelper) On(sType string, eType string, fnName string) {
	var driver scriptDriver
	switch {
	case sType == handlerJs:
		driver = script.handler.jsDriver

	case sType == handlerLua:
		driver = script.handler.luaDriver

	case sType == handlerLisp:
		driver = script.handler.lispDriver

	default:
		return
	}

	handler := newEventListenerScript(driver, eType, fnName)
	script.handler.handlers.Remove(handler)
	script.handler.handlers.Add(handler)
}

// AddHandler adds a PRIVMSG handler of the gien script type
func (script *scriptHelper) AddHandler(sType, fnName string) {
	var driver scriptDriver
	switch {
	case sType == handlerJs:
		driver = script.handler.jsDriver

	case sType == handlerLua:
		driver = script.handler.luaDriver

	case sType == handlerLisp:
		driver = script.handler.lispDriver

	default:
		return
	}

	handler := newEventListenerScript(driver, "PRIVMSG", fnName)
	script.handler.handlers.Remove(handler)
	script.handler.handlers.Add(handler)
}

// RemoveHandler removes a PRIVMSG handler
func (script *scriptHelper) RemoveHandler(sType, fnName string) {
	script.handler.handlers.RemoveId("listener-PRIVMSG-" + sType + "-" + fnName)
}
