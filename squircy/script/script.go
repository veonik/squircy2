package script

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/robertkrimen/otto"
	//ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/veonik/go-lisp/lisp"
	"log"
	"time"
)

const maxExecutionTime = 2 // in seconds
var Halt = errors.New("Execution limit exceeded")
var UnknownScriptType = errors.New("Unknown script type")

type ScriptType string

const (
	Javascript ScriptType = "Javascript"
	Lua                   = "Lua"
	Lisp                  = "Lisp"
)

type ScriptManager struct {
	e            event.EventManager
	jsVm         *otto.Otto
	jsDriver     javascriptDriver
	luaVm        *lua.State
	luaDriver    luaDriver
	lispDriver   lispDriver
	httpHelper   httpHelper
	ircHelper    ircHelper
	dataHelper   dataHelper
	scriptHelper scriptHelper
	repo         ScriptRepository
	l            *log.Logger
}

func NewScriptManager(repo ScriptRepository, l *log.Logger, e event.EventManager) ScriptManager {
	mgr := ScriptManager{
		e,
		nil,
		javascriptDriver{},
		nil,
		luaDriver{},
		lispDriver{},
		httpHelper{},
		ircHelper{},
		dataHelper{make(map[string]interface{})},
		scriptHelper{},
		repo,
		l,
	}
	mgr.init()

	return mgr
}

func (m *ScriptManager) RunUnsafe(t ScriptType, code string) (result interface{}, err error) {
	err = nil
	result = nil

	switch {
	case t == Javascript:
		res, e := runUnsafeJavascript(m.jsVm, code)
		if e != nil {
			err = e
			return
		}
		result, _ = res.Export()

	case t == Lua:
		err = runUnsafeLua(m.luaVm, code)

	case t == Lisp:
		res, e := runUnsafeLisp(code)
		if e != nil {
			err = e
			return
		}
		result = res.Inspect()

	default:
		err = UnknownScriptType
	}

	return
}

func runUnsafeJavascript(vm *otto.Otto, unsafe string) (otto.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == Halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		vm.Interrupt <- func() {
			panic(Halt)
		}
	}()

	return vm.Run(unsafe)
}

func runUnsafeLua(vm *lua.State, unsafe string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == Halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.SetExecutionLimit(maxExecutionTime * (1 << 26))
	err := vm.DoString(unsafe)

	if err != nil && err.Error() == "Lua execution quantum exceeded" {
		panic(Halt)
	}

	return err
}

func runUnsafeLisp(unsafe string) (lisp.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err.(error).Error() == "Execution limit exceeded" {
				fmt.Println("Some code took too long! Stopping after: ", duration)
				panic(Halt)
			}
			panic(err)
		}
	}()

	lisp.SetExecutionLimit(maxExecutionTime * (1 << 15))
	return lisp.EvalString(unsafe)
}

func (m *ScriptManager) init() {
	m.e.Clear(irc.PrivmsgEvent)
	m.e.Clear(irc.NoticeEvent)
	m.e.Clear(irc.ConnectEvent)
	m.e.Clear(irc.ConnectingEvent)
	m.e.Clear(irc.DisconnectEvent)
	m.e.Clear(irc.IrcEvent)
	m.e.Bind(irc.ConnectingEvent, func(mgr *irc.IrcConnectionManager) {
		m.ircHelper.conn = mgr.Connection()
	})

	m.e.Bind(irc.DisconnectEvent, func(ev event.Event) {
		m.ircHelper.conn = nil
	})

	jsVm := otto.New()

	luaVm := lua.NewState()
	luaVm.OpenLibs()
	luaVm.Register("typename", func(vm *lua.State) int {
		o := vm.Typename(int(vm.Type(1)))
		vm.PushString(o)
		return 1
	})
	luaVm.Register("setex", func(vm *lua.State) int {
		key := vm.ToString(1)
		value := vm.ToString(2)
		m.dataHelper.Set(key, value)
		return 0
	})
	luaVm.Register("getex", func(vm *lua.State) int {
		key := vm.ToString(1)
		value := m.dataHelper.Get(key)
		if value != nil {
			vm.PushString(value.(string))
			return 1
		}
		vm.PushNil()
		return 1
	})
	luaVm.Register("joinchan", func(vm *lua.State) int {
		channel := vm.ToString(1)
		m.ircHelper.Join(channel)
		return 0
	})
	luaVm.Register("partchan", func(vm *lua.State) int {
		channel := vm.ToString(1)
		m.ircHelper.Part(channel)
		return 0
	})
	luaVm.Register("privmsg", func(vm *lua.State) int {
		target := vm.ToString(1)
		message := vm.ToString(2)
		m.ircHelper.Privmsg(target, message)
		return 0
	})
	luaVm.Register("httpget", func(vm *lua.State) int {
		url := vm.ToString(1)
		res := m.httpHelper.Get(url)
		vm.PushString(res)
		return 1
	})
	luaVm.Register("on", func(vm *lua.State) int {
		typeName := vm.ToString(1)
		eventType := vm.ToString(2)
		fnName := vm.ToString(3)
		m.scriptHelper.On(typeName, eventType, fnName)
		return 0
	})
	luaVm.Register("addhandler", func(vm *lua.State) int {
		typeName := vm.ToString(1)
		fnName := vm.ToString(2)
		m.scriptHelper.AddHandler(typeName, fnName)
		return 0
	})
	luaVm.Register("removehandler", func(vm *lua.State) int {
		typeName := vm.ToString(1)
		fnName := vm.ToString(2)
		m.scriptHelper.RemoveHandler(typeName, fnName)
		return 0
	})

	lisp.SetHandler("setex", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		key := vars[0].String()
		value := vars[1].String()
		m.dataHelper.Set(key, value)
		return lisp.Nil, nil
	})
	lisp.SetHandler("getex", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		key := vars[0].String()
		if val := m.dataHelper.Get(key); val != nil {
			return lisp.StringValue(val.(string)), nil
		}
		return lisp.Nil, nil
	})
	lisp.SetHandler("joinchan", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		channel := vars[0].String()
		m.ircHelper.Join(channel)
		return lisp.Nil, nil
	})
	lisp.SetHandler("partchan", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		channel := vars[0].String()
		m.ircHelper.Part(channel)
		return lisp.Nil, nil
	})
	lisp.SetHandler("privmsg", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		target := vars[0].String()
		message := vars[1].String()
		m.ircHelper.Privmsg(target, message)
		return lisp.Nil, nil
	})
	lisp.SetHandler("httpget", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 1 {
			return lisp.Nil, nil
		}
		url := vars[0].String()
		res := m.httpHelper.Get(url)
		return lisp.StringValue(res), nil
	})
	lisp.SetHandler("on", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 3 {
			return lisp.Nil, nil
		}
		typeName := vars[0].String()
		eventType := vars[1].String()
		fnName := vars[2].String()
		m.scriptHelper.On(typeName, eventType, fnName)
		return lisp.Nil, nil
	})
	lisp.SetHandler("addhandler", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		typeName := vars[0].String()
		fnName := vars[1].String()
		m.scriptHelper.AddHandler(typeName, fnName)
		return lisp.Nil, nil
	})
	lisp.SetHandler("removehandler", func(vars ...lisp.Value) (lisp.Value, error) {
		if len(vars) != 2 {
			return lisp.Nil, nil
		}
		typeName := vars[0].String()
		fnName := vars[1].String()
		m.scriptHelper.RemoveHandler(typeName, fnName)
		return lisp.Nil, nil
	})

	m.jsVm = jsVm
	m.jsDriver.vm = jsVm
	m.luaVm = luaVm
	m.luaDriver.vm = luaVm

	m.scriptHelper = scriptHelper{m.e, m.jsDriver, m.luaDriver, m.lispDriver}

	jsVm.Set("Http", &m.httpHelper)
	jsVm.Set("Data", &m.dataHelper)
	jsVm.Set("Irc", &m.ircHelper)
	jsVm.Set("Script", &m.scriptHelper)

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		m.l.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
