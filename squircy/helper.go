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

type scriptHelper struct {
	handler *ScriptHandler
}

func (script *scriptHelper) AddHandler(typeName, fnName string) {
	switch {
	case typeName == "js":
		handler := newJavascriptScript(script.handler.conn, script.handler.jsVm, fnName)
		script.handler.handlers.Remove(handler)
		script.handler.handlers.Add(handler)

	case typeName == "lua":
		handler := newLuaScript(script.handler.conn, script.handler.luaVm, fnName)
		script.handler.handlers.Remove(handler)
		script.handler.handlers.Add(handler)

	case typeName == "lisp":
		handler := newLispScript(script.handler.conn, fnName)
		script.handler.handlers.Remove(handler)
		script.handler.handlers.Add(handler)
	}
}

func (script *scriptHelper) RemoveHandler(typeName, fnName string) {
	switch {
	case typeName == "js":
		script.handler.handlers.RemoveId("js-" + fnName)

	case typeName == "lua":
		script.handler.handlers.RemoveId("lua-" + fnName)

	case typeName == "lisp":
		script.handler.handlers.RemoveId("lisp-" + fnName)
	}
}
