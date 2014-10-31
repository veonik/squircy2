package script

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	anko_parser "github.com/mattn/anko/parser"
	anko "github.com/mattn/anko/vm"
	"github.com/robertkrimen/otto"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	glisp "github.com/zhemao/glisp/interpreter"
	"log"
	"reflect"
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
	Anko                  = "Anko"
)

type ScriptManager struct {
	e            event.EventManager
	jsVm         *otto.Otto
	jsDriver     javascriptDriver
	luaVm        *lua.State
	luaDriver    luaDriver
	lispVm       *glisp.Glisp
	lispDriver   lispDriver
	ankoVm       *anko.Env
	ankoDriver   ankoDriver
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
		nil,
		lispDriver{},
		nil,
		ankoDriver{},
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

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()

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
		res, e := runUnsafeLisp(m.lispVm, code)
		if e != nil {
			err = e
			return
		}
		result = sexpToInterface(res)

	case t == Anko:
		res, e := runUnsafeAnko(m.ankoVm, code)
		if e != nil {
			err = e
			return
		}
		result = res.Interface()

	default:
		err = UnknownScriptType
	}

	return
}

func runUnsafeJavascript(vm *otto.Otto, unsafe string) (val otto.Value, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e == Halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			err = e.(error)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		vm.Interrupt <- func() {
			panic(Halt)
		}
	}()

	val, err = vm.Run(unsafe)

	return
}

func runUnsafeLua(vm *lua.State, unsafe string) (err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e == Halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			err = e.(error)
		}
	}()

	vm.SetExecutionLimit(maxExecutionTime * (1 << 26))
	err = vm.DoString(unsafe)

	return
}

func runUnsafeLisp(vm *glisp.Glisp, unsafe string) (val glisp.Sexp, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e.(error).Error() == "Execution limit exceeded" {
				fmt.Println("Some code took too long! Stopping after: ", duration)
				e = Halt
			}
			val = glisp.SexpNull
			err = e.(error)
		}
	}()

	vm.Clear()
	vm.LoadString(unsafe)
	val, err = vm.Run()

	return
}

func sexpToInterface(val glisp.Sexp) interface{} {
	switch t := val.(type) {
	case glisp.SexpInt:
		return int(t)

	case glisp.SexpFloat:
		return float64(t)

	case glisp.SexpStr:
		return string(t)

	default:
		return val.SexpString()
	}
}

func runUnsafeAnko(vm *anko.Env, unsafe string) (val reflect.Value, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e == Halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			val = anko.NilValue
			err = e.(error)
		}
	}()

	scanner := &anko_parser.Scanner{}
	scanner.Init(unsafe)
	stmts, err := anko_parser.Parse(scanner)
	if err != nil {
		val = anko.NilValue
		return
	}
	val, err = anko.Run(stmts, vm)

	return
}

func (m *ScriptManager) ReInit() {
	m.init()
}

func (m *ScriptManager) init() {
	m.e.ClearAll()
	m.e.Bind(irc.ConnectingEvent, func(mgr *irc.IrcConnectionManager) {
		m.ircHelper.conn = mgr.Connection()
	})

	m.e.Bind(irc.DisconnectEvent, func(ev event.Event) {
		m.ircHelper.conn = nil
	})

	jsVm := newJavascriptVm(m)
	luaVm := newLuaVm(m)
	lispVm := newLispVm(m)
	ankoVm := newAnkoVm(m)

	m.jsVm = jsVm
	m.jsDriver.vm = jsVm
	m.luaVm = luaVm
	m.luaDriver.vm = luaVm
	m.lispVm = lispVm
	m.lispDriver.vm = lispVm
	m.ankoVm = ankoVm
	m.ankoDriver.vm = ankoVm

	m.scriptHelper = scriptHelper{m.e, m.jsDriver, m.luaDriver, m.lispDriver, m.ankoDriver, make(map[string]event.EventHandler, 0)}

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		m.l.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
