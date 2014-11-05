package script

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	anko_core "github.com/mattn/anko/builtins"
	anko_encoding "github.com/mattn/anko/builtins/encoding"
	anko_flag "github.com/mattn/anko/builtins/flag"
	anko_io "github.com/mattn/anko/builtins/io"
	anko_math "github.com/mattn/anko/builtins/math"
	anko_net "github.com/mattn/anko/builtins/net"
	anko_os "github.com/mattn/anko/builtins/os"
	anko_path "github.com/mattn/anko/builtins/path"
	anko_regexp "github.com/mattn/anko/builtins/regexp"
	anko_sort "github.com/mattn/anko/builtins/sort"
	anko_strings "github.com/mattn/anko/builtins/strings"
	anko_term "github.com/mattn/anko/builtins/term"
	anko "github.com/mattn/anko/vm"
	"github.com/robertkrimen/otto"
	"github.com/tyler-sommer/squircy2/squircy/event"
	glispext "github.com/zhemao/glisp/extensions"
	glisp "github.com/zhemao/glisp/interpreter"
	"reflect"
	"strconv"
)

func newJavascriptVm(m *ScriptManager) *otto.Otto {
	jsVm := otto.New()
	jsVm.Set("Http", &m.httpHelper)
	jsVm.Set("Data", &m.dataHelper)
	jsVm.Set("Irc", &m.ircHelper)
	jsVm.Set("bind", func(call otto.FunctionCall) otto.Value {
		eventType := call.Argument(0).String()
		fnName := call.Argument(1).String()
		m.scriptHelper.Bind(Javascript, event.EventType(eventType), fnName)
		return otto.UndefinedValue()
	})
	jsVm.Set("unbind", func(call otto.FunctionCall) otto.Value {
		eventType := call.Argument(0).String()
		fnName := call.Argument(1).String()
		m.scriptHelper.Unbind(Javascript, event.EventType(eventType), fnName)
		return otto.UndefinedValue()
	})

	return jsVm
}

func newLuaVm(m *ScriptManager) *lua.State {
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
	luaVm.Register("bind", func(vm *lua.State) int {
		eventType := vm.ToString(1)
		fnName := vm.ToString(2)
		m.scriptHelper.Bind(Lua, event.EventType(eventType), fnName)
		return 0
	})
	luaVm.Register("unbind", func(vm *lua.State) int {
		eventType := vm.ToString(1)
		fnName := vm.ToString(2)
		m.scriptHelper.Unbind(Lua, event.EventType(eventType), fnName)
		return 0
	})

	return luaVm
}

func newLispVm(m *ScriptManager) *glisp.Glisp {
	lispVm := glisp.NewGlisp()
	lispVm.ImportEval()
	glispext.ImportRandom(lispVm)
	glispext.ImportTime(lispVm)
	glispext.ImportChannels(lispVm)
	glispext.ImportCoroutines(lispVm)

	lispVm.AddFunction("setex", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		key := sexpToString(args[0])
		val := exportSexp(args[1])
		m.dataHelper.Set(key, val)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("getex", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		key := sexpToString(args[0])
		val := m.dataHelper.Get(key).(string)

		return glisp.SexpStr(val), nil
	})
	lispVm.AddFunction("joinchan", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		channel := sexpToString(args[0])
		m.ircHelper.Join(channel)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("partchan", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		channel := sexpToString(args[0])
		m.ircHelper.Part(channel)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("privmsg", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		target := sexpToString(args[0])
		message := sexpToString(args[1])
		m.ircHelper.Privmsg(target, message)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("httpget", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		url := sexpToString(args[0])
		resp := m.httpHelper.Get(url)

		return glisp.SexpStr(resp), nil
	})
	lispVm.AddFunction("bind", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		eventType := sexpToString(args[0])
		fnName := sexpToString(args[1])
		m.scriptHelper.Bind(Lisp, event.EventType(eventType), fnName)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("unbind", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 2 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		eventType := sexpToString(args[0])
		fnName := sexpToString(args[1])
		m.scriptHelper.Unbind(Lisp, event.EventType(eventType), fnName)

		return glisp.SexpNull, nil
	})
	lispVm.AddFunction("parse-integer", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		switch t := args[0].(type) {
		case glisp.SexpStr:
			val, err := strconv.ParseInt(string(t), 0, 64)
			if err != nil {
				return glisp.SexpNull, err
			}
			return glisp.SexpInt(int(val)), nil

		default:
			return glisp.SexpNull, errors.New(fmt.Sprintf("cannot convert %v to int", t))
		}
	})
	lispVm.AddFunction("write-to-string", func(vm *glisp.Glisp, name string, args []glisp.Sexp) (glisp.Sexp, error) {
		if len(args) != 1 {
			return glisp.SexpNull, errors.New("incorrect number of arguments")
		}

		switch t := args[0].(type) {
		case glisp.SexpStr:
			return t, nil

		case glisp.SexpInt:
			return glisp.SexpStr(strconv.Itoa(int(t))), nil

		default:
			return glisp.SexpStr(fmt.Sprintf("%v", t)), nil
		}
	})

	return lispVm
}

func newAnkoVm(m *ScriptManager) *anko.Env {
	ankoVm := anko.NewEnv()
	anko_core.Import(ankoVm)
	anko_flag.Import(ankoVm)
	anko_net.Import(ankoVm)
	anko_encoding.Import(ankoVm)
	anko_os.Import(ankoVm)
	anko_io.Import(ankoVm)
	anko_math.Import(ankoVm)
	anko_path.Import(ankoVm)
	anko_regexp.Import(ankoVm)
	anko_sort.Import(ankoVm)
	anko_strings.Import(ankoVm)
	anko_term.Import(ankoVm)

	mod := ankoVm.NewModule("data")
	mod.Define("Get", reflect.ValueOf(m.dataHelper.Get))
	mod.Define("Set", reflect.ValueOf(m.dataHelper.Set))

	mod = ankoVm.NewModule("irc")
	mod.Define("Join", reflect.ValueOf(func(channel string) {
		m.ircHelper.Join(channel)
	}))
	mod.Define("Part", reflect.ValueOf(func(channel string) {
		m.ircHelper.Part(channel)
	}))
	mod.Define("Privmsg", reflect.ValueOf(func(target, message string) {
		m.ircHelper.Privmsg(target, message)
	}))

	mod = ankoVm.NewModule("strconv")
	mod.Define("ParseInt", reflect.ValueOf(func(s string) int {
		i, _ := strconv.ParseInt(s, 0, 64)
		return int(i)
	}))

	ankoVm.Define("bind", reflect.ValueOf(func(eventType, fnName string) {
		m.scriptHelper.Bind(Anko, event.EventType(eventType), fnName)
	}))

	ankoVm.Define("unbind", reflect.ValueOf(func(eventType, fnName string) {
		m.scriptHelper.Unbind(Anko, event.EventType(eventType), fnName)
	}))

	return ankoVm
}
