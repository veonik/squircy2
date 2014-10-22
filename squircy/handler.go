package squircy

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/fzzy/radix/redis"
	"github.com/robertkrimen/otto"
	"github.com/thoj/go-ircevent"
	"github.com/veonik/go-lisp/lisp"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const maxExecutionTime = 2 // in seconds
var halt = errors.New("Execution limit exceeded")

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

func replTypePretty(replType string) string {
	switch {
	case replType == "lua":
		return "Lua"

	case replType == "js":
		return "Javascript"

	case replType == "lisp":
		return "Lisp"
	}

	return "Unknown"
}

func scriptRecoveryHandler(conn *irc.Connection, e *irc.Event) {
	if err := recover(); err != nil {
		fmt.Println("An error occurred", err)
		if err == halt {
			conn.Privmsgf(replyTarget(e), "Script halted")
		}
	}
}

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

type ScriptDatastore map[string]string

type ScriptHandler struct {
	conn     *irc.Connection
	handlers *HandlerCollection
	config   *Configuration
	luaVm    *lua.State
	jsVm     *otto.Otto
	client   *redis.Client
	repl     bool
	replType string
	data     ScriptDatastore
}

func newScriptHandler(conn *irc.Connection, handlers *HandlerCollection, config *Configuration, client *redis.Client) *ScriptHandler {
	h := &ScriptHandler{conn, handlers, config, nil, nil, client, false, "", make(ScriptDatastore)}

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
			h.conn.Privmsgf(replyTarget(e), "%s REPL session ended.", replTypePretty(h.replType))
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

		switch {
		case args[0] == "js":
			handler := newJavascriptScript(h.conn, h.jsVm, args[1])
			h.handlers.Remove(handler)
			h.handlers.Add(handler)

		case args[0] == "lua":
			handler := newLuaScript(h.conn, h.luaVm, args[1])
			h.handlers.Remove(handler)
			h.handlers.Add(handler)

		case args[0] == "lisp":
			handler := newLispScript(h.conn, args[1])
			h.handlers.Remove(handler)
			h.handlers.Add(handler)
		}

	case command == "unregister":
		if len(args) != 2 && args[0] != "js" && args[0] != "lua" && args[0] != "lisp" {
			h.conn.Privmsgf(replyTarget(e), "Invalid syntax. Usage: !unregister <js|lua|lisp> <fn name>")

			return
		}

		switch {
		case args[0] == "js":
			h.conn.Privmsgf(replyTarget(e), "Unregistered Javsacript handler "+args[1])
			h.handlers.RemoveId("js-" + args[1])

		case args[0] == "lua":
			h.conn.Privmsgf(replyTarget(e), "Unregistered Lua handler "+args[1])
			h.handlers.RemoveId("lua-" + args[1])

		case args[0] == "lisp":
			h.conn.Privmsgf(replyTarget(e), "Unregistered Lisp handler "+args[1])
			h.handlers.RemoveId("lisp-" + args[1])
		}

	case command == "repl":
		if len(args) != 1 && args[0] != "js" && args[0] != "lua" && args[0] != "lisp" {
			h.conn.Privmsgf(replyTarget(e), "Invalid syntax. Usage: !repl <js|lua|lisp>")
			return
		}

		h.repl = true
		h.replType = args[0]
		h.conn.Privmsgf(replyTarget(e), "%s REPL session started.", replTypePretty(h.replType))
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

	h.luaVm = luaVm
	h.jsVm = jsVm

	client := &httpHelper{}
	cres, _ := h.jsVm.ToValue(client)
	h.jsVm.Set("Http", cres)
	db := &dataHelper{make(map[string]interface{})}
	dres, _ := h.jsVm.ToValue(db)
	h.jsVm.Set("Data", dres)
	irc := &ircHelper{h.conn}
	ires, _ := h.jsVm.ToValue(irc)
	h.jsVm.Set("Irc", ires)

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

	repo := scriptRepository{h.client}
	scripts := repo.Fetch()
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

func newJavascriptScript(conn *irc.Connection, vm *otto.Otto, fn string) *JavascriptScript {
	return &JavascriptScript{conn, vm, fn}
}

type JavascriptScript struct {
	conn *irc.Connection
	vm   *otto.Otto
	fn   string
}

func (h *JavascriptScript) Id() string {
	return "js-" + h.fn
}

func (h *JavascriptScript) Matches(e *irc.Event) bool {
	return true
}

func runUnsafeJavascript(vm *otto.Otto, unsafe string) (otto.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		vm.Interrupt <- func() {
			panic(halt)
		}
	}()

	return vm.Run(unsafe)
}

func (h *JavascriptScript) Handle(e *irc.Event) {
	defer scriptRecoveryHandler(h.conn, e)

	h.vm.Set("print", func(call otto.FunctionCall) otto.Value {
		message, _ := call.Argument(0).ToString()
		h.conn.Privmsgf(replyTarget(e), message)
		return otto.Value{}
	})
	_, err := runUnsafeJavascript(h.vm, fmt.Sprintf("%s(\"%s\", \"%s\", \"%s\")", h.fn, e.Arguments[0], e.Nick, e.Message()))
	if err != nil {
		h.conn.Privmsgf(replyTarget(e), err.Error())

		return
	}
}

func newLuaScript(conn *irc.Connection, vm *lua.State, fn string) *LuaScript {
	return &LuaScript{conn, vm, fn}
}

type LuaScript struct {
	conn *irc.Connection
	vm   *lua.State
	fn   string
}

func (h *LuaScript) Id() string {
	return "lua-" + h.fn
}

func (h *LuaScript) Matches(e *irc.Event) bool {
	return true
}

func runUnsafeLua(vm *lua.State, unsafe string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.SetExecutionLimit(maxExecutionTime * (1 << 26))
	err := vm.DoString(unsafe)

	if err != nil && err.Error() == "Lua execution quantum exceeded" {
		panic(halt)
	}

	return err
}

func (h *LuaScript) Handle(e *irc.Event) {
	defer scriptRecoveryHandler(h.conn, e)

	h.vm.Register("print", func(vm *lua.State) int {
		o := vm.ToString(1)
		h.conn.Privmsgf(replyTarget(e), o)
		return 0
	})
	err := runUnsafeLua(h.vm, fmt.Sprintf("%s(\"%s\", \"%s\", \"%s\")", h.fn, e.Arguments[0], e.Nick, e.Message()))
	if err != nil {
		h.conn.Privmsgf(replyTarget(e), err.Error())
	}
}

func newLispScript(conn *irc.Connection, fn string) *LispScript {
	return &LispScript{conn, fn}
}

type LispScript struct {
	conn *irc.Connection
	fn   string
}

func (h *LispScript) Id() string {
	return "lisp-" + h.fn
}

func (h *LispScript) Matches(e *irc.Event) bool {
	return true
}

func runUnsafeLisp(unsafe string) (lisp.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err.(error).Error() == "Execution limit exceeded" {
				fmt.Println("Some code took too long! Stopping after: ", duration)
				panic(halt)
			}
			panic(err)
		}
	}()

	lisp.SetExecutionLimit(maxExecutionTime * (1 << 15))
	return lisp.EvalString(unsafe)
}

func (h *LispScript) Handle(e *irc.Event) {
	defer scriptRecoveryHandler(h.conn, e)

	lisp.SetHandler("print", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) == 1 {
			h.conn.Privmsgf(replyTarget(e), vars[0].String())
		}
		return lisp.Nil, nil
	})
	_, err := runUnsafeLisp(fmt.Sprintf("(%s \"%s\" \"%s\" \"%s\")", h.fn, e.Arguments[0], e.Nick, e.Message()))

	if err == halt {
		panic(err)

	} else if err != nil {
		h.conn.Privmsgf(replyTarget(e), err.Error())

		return
	}
}
