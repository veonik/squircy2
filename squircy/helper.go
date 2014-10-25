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

type handlerType string

const (
	handlerJs   handlerType = "js"
	handlerLua              = "lua"
	handlerLisp             = "lisp"
)

type eventType string

const (
	eventJoin    eventType = "join"
	eventPart              = "part"
	eventMessage           = "privmsg"
	eventNotice            = "notice"
	eventConnect           = "connect"
)

var eventMap = map[eventType]string{
	eventJoin:    "JOIN",
	eventPart:    "PART",
	eventMessage: "PRIVMSG",
	eventNotice:  "NOTICE",
	eventConnect: "001",
}

type scriptHelper struct {
	handler *ScriptHandler
}

// On is exposed by javascript and allows strings instead of typed strings
func (script *scriptHelper) On(sType string, eType string, fnName string) {
	script.on(handlerType(sType), eventType(eType), fnName)
}

func (script *scriptHelper) on(sType handlerType, eType eventType, fnName string) {
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

	handler := newEventListenerScript(driver, eventMap[eType], fnName)
	script.handler.handlers.Remove(handler)
	script.handler.handlers.Add(handler)
}

func (script *scriptHelper) AddHandler(typeName, fnName string) {
	var driver scriptDriver
	switch sType := handlerType(typeName); {
	case sType == handlerJs:
		driver = script.handler.jsDriver

	case sType == handlerLua:
		driver = script.handler.luaDriver

	case sType == handlerLisp:
		driver = script.handler.lispDriver

	default:
		return
	}

	handler := newEventListenerScript(driver, eventMap[eventMessage], fnName)
	script.handler.handlers.Remove(handler)
	script.handler.handlers.Add(handler)
}

func (script *scriptHelper) RemoveHandler(typeName, fnName string) {
	script.handler.handlers.RemoveId("listener-" + eventMap[eventMessage] + "-" + typeName + "-" + fnName)
}
