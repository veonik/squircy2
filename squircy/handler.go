package squircy

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/robertkrimen/otto"
	"github.com/thoj/go-ircevent"
	"github.com/veonik/go-lisp/lisp"
	"log"
	"strconv"
	"strings"
)

var replNames = map[string]string{
	"lua":  "Lua",
	"js":   "Javascript",
	"lisp": "Lisp",
}

func replyTarget(e *irc.Event) string {
	if strings.HasPrefix(e.Arguments[0], "#") {
		return e.Arguments[0]
	} else {
		return e.Nick
	}
}

func parseCommand(msg string) (string, []string) {
	fields := strings.Fields(msg)
	if len(fields) < 1 {
		return "", nil
	}

	command := fields[0][1:]
	args := fields[1:]

	return command, args
}

type NickservHandler struct {
	conn     *irc.Connection
	log      *log.Logger
	handlers *HandlerCollection
	config   *Configuration
}

func newNickservHandler(conn *irc.Connection, log *log.Logger, handlers *HandlerCollection, config *Configuration) (h *NickservHandler) {
	h = &NickservHandler{conn, log, handlers, config}

	return
}

func (h *NickservHandler) Id() string {
	return "nickserv"
}

func (h *NickservHandler) Matches(e *irc.Event) bool {
	return strings.Contains(strings.ToLower(e.Message()), "identify") && e.User == "NickServ"
}

func (h *NickservHandler) Handle(e *irc.Event) {
	h.conn.Privmsgf("NickServ", "IDENTIFY %s", h.config.Password)
	h.log.Println("Identified with Nickserv")
	h.handlers.Remove(h)
}

func scriptRecoveryHandler(conn *irc.Connection, e *irc.Event) {
	if err := recover(); err != nil {
		fmt.Println("An error occurred", err)
		if err == halt {
			conn.Privmsgf(replyTarget(e), "Script halted")
		}
	}
}

type scriptDriver interface {
	Handle(e *irc.Event, fnName string)
	String() string
}

type javascriptDriver struct {
	vm *otto.Otto
}

func (d javascriptDriver) Handle(e *irc.Event, fnName string) {
	d.vm.Set("replyTarget", func(call otto.FunctionCall) otto.Value {
		val, _ := otto.ToValue(replyTarget(e))
		return val
	})

	d.vm.Interrupt = make(chan func(), 1)
	d.vm.Call(fnName, otto.NullValue(), e.Code, e.Arguments[0], e.Nick, e.Message())
}

func (d javascriptDriver) String() string {
	return "js"
}

type luaDriver struct {
	vm *lua.State
}

func (d luaDriver) Handle(e *irc.Event, fnName string) {
	d.vm.Register("replytarget", func(vm *lua.State) int {
		vm.PushString(replyTarget(e))
		return 1
	})

	d.vm.GetGlobal(fnName)
	d.vm.PushString(e.Code)
	d.vm.PushString(e.Arguments[0])
	d.vm.PushString(e.Nick)
	d.vm.PushString(e.Message())
	d.vm.Call(4, 0)
}

func (d luaDriver) String() string {
	return "lua"
}

type lispDriver struct{}

func (d lispDriver) Handle(e *irc.Event, fnName string) {
	lisp.SetHandler("replytarget", func(vars ...lisp.Value) (lisp.Value, error) {
		return lisp.StringValue(replyTarget(e)), nil
	})
	_, err := runUnsafeLisp(fmt.Sprintf("(%s \"%s\" \"%s\" \"%s\" %s)", fnName, e.Code, e.Arguments[0], e.Nick, strconv.Quote(e.Message())))

	if err == halt {
		panic(err)
	}
}

func (d lispDriver) String() string {
	return "lisp"
}

type ScriptHandler struct {
	conn       *irc.Connection
	handlers   *HandlerCollection
	config     *Configuration
	luaVm      *lua.State
	jsVm       *otto.Otto
	helper     *scriptHelper
	repo       scriptRepository
	repl       bool
	replType   string
	jsDriver   javascriptDriver
	luaDriver  luaDriver
	lispDriver lispDriver
}

func newScriptHandler(conn *irc.Connection, handlers *HandlerCollection, config *Configuration, repo scriptRepository) *ScriptHandler {
	h := &ScriptHandler{conn, handlers, config, nil, nil, nil, repo, false, "", javascriptDriver{}, luaDriver{}, lispDriver{}}

	h.init()

	return h
}

func (h *ScriptHandler) Id() string {
	return "scripting"
}

func (h *ScriptHandler) Matches(e *irc.Event) bool {
	return e.Nick == h.config.OwnerNick && e.Host == h.config.OwnerHost
}

func (h *ScriptHandler) Handle(e *irc.Event) {
	defer scriptRecoveryHandler(h.conn, e)

	if h.repl == true {
		msg := e.Message()
		if strings.HasPrefix(msg, "!repl end") {
			h.conn.Privmsgf(replyTarget(e), "%s REPL session ended.", replNames[h.replType])
			h.repl = false
			h.replType = ""
			return
		}

		switch {
		case h.replType == "lua":
			h.luaVm.Register("print", func(vm *lua.State) int {
				o := vm.ToString(1)
				h.conn.Privmsgf(replyTarget(e), o)
				return 0
			})
			h.luaVm.Register("replytarget", func(vm *lua.State) int {
				vm.PushString(replyTarget(e))
				return 1
			})
			err := runUnsafeLua(h.luaVm, msg)
			if err != nil {
				h.conn.Privmsgf(replyTarget(e), err.Error())
			}

		case h.replType == "js":
			h.jsVm.Set("print", func(call otto.FunctionCall) otto.Value {
				message, _ := call.Argument(0).ToString()
				h.conn.Privmsgf(replyTarget(e), message)
				return otto.Value{}
			})
			h.jsVm.Set("replyTarget", func(call otto.FunctionCall) otto.Value {
				val, _ := otto.ToValue(replyTarget(e))
				return val
			})
			_, err := runUnsafeJavascript(h.jsVm, msg)
			if err != nil {
				h.conn.Privmsgf(replyTarget(e), err.Error())

				return
			}

		case h.replType == "lisp":
			lisp.SetHandler("print", func(vars ...lisp.Value) (lisp.Value, error) {
				if len(vars) == 1 {
					h.conn.Privmsgf(replyTarget(e), vars[0].String())
				}
				return lisp.Nil, nil
			})
			lisp.SetHandler("replytarget", func(vars ...lisp.Value) (lisp.Value, error) {
				return lisp.StringValue(replyTarget(e)), nil
			})
			_, err := runUnsafeLisp(msg)
			if err != nil {
				h.conn.Privmsgf(replyTarget(e), err.Error())

				return
			}
		}

		return
	}

	switch command, args := parseCommand(e.Message()); {
	case command == "":
		break

	case command == "register":
		if len(args) != 2 && args[0] != "js" && args[0] != "lua" && args[0] != "lisp" {
			h.conn.Privmsgf(replyTarget(e), "Invalid syntax. Usage: !register <js|lua|lisp> <fn name>")

			return
		}

		h.conn.Privmsgf(replyTarget(e), "Registered", replNames[args[0]], "handler", args[1])
		h.helper.AddHandler(args[0], args[1])

	case command == "unregister":
		if len(args) != 2 && args[0] != "js" && args[0] != "lua" && args[0] != "lisp" {
			h.conn.Privmsgf(replyTarget(e), "Invalid syntax. Usage: !unregister <js|lua|lisp> <fn name>")

			return
		}

		h.conn.Privmsgf(replyTarget(e), "Unregistered", replNames[args[0]], "handler", args[1])
		h.helper.RemoveHandler(args[0], args[1])

	case command == "repl":
		if len(args) != 1 && args[0] != "js" && args[0] != "lua" && args[0] != "lisp" {
			h.conn.Privmsgf(replyTarget(e), "Invalid syntax. Usage: !repl <js|lua|lisp>")
			return
		}

		h.repl = true
		h.replType = args[0]
		h.conn.Privmsgf(replyTarget(e), "%s REPL session started.", replNames[h.replType])
	}
}

func (h *ScriptHandler) ReInit() {
	h.repl = false
	h.replType = ""
	h.init()
}

func (h *ScriptHandler) init() {
	luaVm := lua.NewState()
	luaVm.OpenLibs()

	jsVm := otto.New()

	h.jsDriver.vm = jsVm
	h.luaDriver.vm = luaVm

	helper := &scriptHelper{h}
	client := &httpHelper{}
	db := &dataHelper{make(map[string]interface{})}
	irc := &ircHelper{h.conn}

	h.luaVm = luaVm
	h.jsVm = jsVm
	h.helper = helper

	h.jsVm.Set("Http", client)
	h.jsVm.Set("Data", db)
	h.jsVm.Set("Irc", irc)
	h.jsVm.Set("Script", helper)

	h.luaVm.Register("typename", func(vm *lua.State) int {
		o := vm.Typename(int(vm.Type(1)))
		h.luaVm.PushString(o)
		return 1
	})
	h.luaVm.Register("setex", func(vm *lua.State) int {
		key := vm.ToString(1)
		value := vm.ToString(2)
		db.Set(key, value)
		return 0
	})
	h.luaVm.Register("getex", func(vm *lua.State) int {
		key := vm.ToString(1)
		value := db.Get(key)
		if value != nil {
			vm.PushString(value.(string))
			return 1
		}
		vm.PushNil()
		return 1
	})
	h.luaVm.Register("joinchan", func(vm *lua.State) int {
		channel := vm.ToString(1)
		irc.Join(channel)
		return 0
	})
	h.luaVm.Register("partchan", func(vm *lua.State) int {
		channel := vm.ToString(1)
		irc.Part(channel)
		return 0
	})
	h.luaVm.Register("privmsg", func(vm *lua.State) int {
		target := vm.ToString(1)
		message := vm.ToString(2)
		irc.Privmsg(target, message)
		return 0
	})
	h.luaVm.Register("httpget", func(vm *lua.State) int {
		url := vm.ToString(1)
		res := client.Get(url)
		vm.PushString(res)
		return 1
	})
	h.luaVm.Register("addhandler", func(vm *lua.State) int {
		typeName := vm.ToString(1)
		fnName := vm.ToString(2)
		helper.AddHandler(typeName, fnName)
		return 0
	})
	h.luaVm.Register("removehandler", func(vm *lua.State) int {
		typeName := vm.ToString(1)
		fnName := vm.ToString(2)
		helper.RemoveHandler(typeName, fnName)
		return 0
	})

	lisp.SetHandler("setex", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		key := vars[0].String()
		value := vars[1].String()
		db.Set(key, value)
		return lisp.Nil, nil
	})
	lisp.SetHandler("getex", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		key := vars[0].String()
		if val := db.Get(key); val != nil {
			return lisp.StringValue(val.(string)), nil
		}
		return lisp.Nil, nil
	})
	lisp.SetHandler("joinchan", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		channel := vars[0].String()
		irc.Join(channel)
		return lisp.Nil, nil
	})
	lisp.SetHandler("partchan", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		channel := vars[0].String()
		irc.Part(channel)
		return lisp.Nil, nil
	})
	lisp.SetHandler("privmsg", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		target := vars[0].String()
		message := vars[1].String()
		irc.Privmsg(target, message)
		return lisp.Nil, nil
	})
	lisp.SetHandler("httpget", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		url := vars[0].String()
		res := client.Get(url)
		return lisp.StringValue(res), nil
	})
	lisp.SetHandler("addhandler", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		typeName := vars[0].String()
		fnName := vars[1].String()
		helper.AddHandler(typeName, fnName)
		return lisp.Nil, nil
	})
	lisp.SetHandler("removehandler", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		typeName := vars[0].String()
		fnName := vars[1].String()
		helper.RemoveHandler(typeName, fnName)
		return lisp.Nil, nil
	})

	scripts := h.repo.FetchAll()
	for _, script := range scripts {
		fmt.Println("Running", script.Type, "script", script.Title)
		switch {
		case script.Type == scriptJavascript:
			runUnsafeJavascript(h.jsVm, script.Body)

		case script.Type == scriptLua:
			runUnsafeLua(h.luaVm, script.Body)

		case script.Type == scriptLisp:
			runUnsafeLisp(script.Body)
		}
	}
}

func newEventListenerScript(driver scriptDriver, eventCode string, fn string) *EventListenerScript {
	return &EventListenerScript{driver, eventCode, fn}
}

type EventListenerScript struct {
	driver    scriptDriver
	eventCode string
	fn        string
}

func (h *EventListenerScript) Id() string {
	return "listener-" + h.eventCode + "-" + h.driver.String() + "-" + h.fn
}

func (h *EventListenerScript) Matches(e *irc.Event) bool {
	return e.Code == h.eventCode
}

func (h *EventListenerScript) Handle(e *irc.Event) {
	h.driver.Handle(e, h.fn)
}
