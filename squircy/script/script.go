package script

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/robertkrimen/otto"
	//"github.com/thoj/go-ircevent"
	"github.com/veonik/go-lisp/lisp"
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
}

func NewScriptManager(repo ScriptRepository) *ScriptManager {
	mgr := &ScriptManager{
		nil,
		javascriptDriver{},
		nil,
		luaDriver{},
		lispDriver{},
		httpHelper{},
		ircHelper{nil},
		dataHelper{make(map[string]interface{})},
		scriptHelper{},
		repo,
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
	luaVm := lua.NewState()
	luaVm.OpenLibs()

	jsVm := otto.New()

	m.luaVm = luaVm
	m.jsVm = jsVm

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		fmt.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
