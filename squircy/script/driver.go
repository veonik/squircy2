package script

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	anko_parser "github.com/mattn/anko/parser"
	anko "github.com/mattn/anko/vm"
	"github.com/robertkrimen/otto"
	"github.com/stevedonovan/luar"
	"github.com/tyler-sommer/squircy2/squircy/event"
	glisp "github.com/zhemao/glisp/interpreter"
	"strconv"
	"strings"
	"time"
)

const maxExecutionTime = 2 // in seconds
var Halt = errors.New("Execution limit exceeded")
var UnknownScriptType = errors.New("Unknown script type")

type scriptDriver interface {
	Handle(e event.Event, fnName string)
	RunUnsafe(code string) (interface{}, error)
	String() string
}

type javascriptDriver struct {
	vm *otto.Otto
}

func (d javascriptDriver) Handle(e event.Event, fnName string) {
	d.vm.Interrupt = make(chan func(), 1)
	data, err := d.vm.ToValue(e.Data)
	if err != nil {
		fmt.Println("An error occurred while creating event data", err)
		return
	}
	_, err = d.vm.Call(fnName, otto.NullValue(), data)
	if err != nil {
		fmt.Println("An error occurred while executing the Javascript handler", err)
	}
}

func (d javascriptDriver) RunUnsafe(unsafe string) (val interface{}, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e == Halt {
				fmt.Println("Some Javascript code took too long! Stopping after: ", duration)
			}
			err = e.(error)
		}
	}()

	d.vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		d.vm.Interrupt <- func() {
			panic(Halt)
		}
	}()

	v, err := d.vm.Run(unsafe)
	val, _ = v.Export()

	return
}

func (d javascriptDriver) String() string {
	return "js"
}

type luaDriver struct {
	vm *lua.State
}

func (d luaDriver) Handle(e event.Event, fnName string) {
	o := luar.NewLuaObjectFromName(d.vm, fnName)
	_, err := o.Call(e.Data)
	if err != nil {
		fmt.Println("An error occurred while executing the Lua handler", err)
	}
}

func (d luaDriver) RunUnsafe(unsafe string) (val interface{}, err error) {
	val = nil // TODO: Lua does not return a value
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if e := recover(); e != nil {
			if e == Halt {
				fmt.Println("Some Lua code took too long! Stopping after: ", duration)
			}
			err = e.(error)
		}
	}()
	d.vm.Register("res", func(vm *lua.State) int {
		val = luar.LuaToGo(vm, nil, 0)
		return 0
	})
	d.vm.SetExecutionLimit(maxExecutionTime * (1 << 26))
	err = d.vm.DoString(unsafe)

	return
}

func (d luaDriver) String() string {
	return "lua"
}

type lispDriver struct {
	vm *glisp.Glisp
}

func (d lispDriver) Handle(e event.Event, fnName string) {
	_, err := d.RunUnsafe(fmt.Sprintf("(%s \"%s\" \"%s\" \"%s\" %s)", fnName, e.Data["Code"], e.Data["Target"], e.Data["Nick"], strconv.Quote(e.Data["Message"].(string))))
	if err != nil {
		fmt.Println("An error occurred while executing the Lisp handler", err)
	}
}

func (d lispDriver) RunUnsafe(unsafe string) (val interface{}, err error) {
	halted := false
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if halted {
			fmt.Println("Some Lisp code took too long! Stopping after: ", duration)
			err = Halt
			val = glisp.SexpNull
			return
		}

		if e := recover(); e != nil {
			val = glisp.SexpNull
			err = e.(error)
		}
	}()

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		d.vm.Clear()
		halted = true
	}()

	d.vm.Clear()
	d.vm.LoadString(unsafe)
	v, err := d.vm.Run()
	val = exportSexp(v)

	return
}

func exportSexp(val glisp.Sexp) interface{} {
	switch t := val.(type) {
	case glisp.SexpSymbol:
		return t.Name()

	case glisp.SexpInt:
		return int(t)

	case glisp.SexpFloat:
		return float64(t)

	case glisp.SexpStr:
		return string(t)

	case glisp.SexpBool:
		return bool(t)

	case glisp.SexpChar:
		return rune(t)

	case glisp.SexpArray:
		res := make([]interface{}, 0)
		for _, sexp := range t {
			res = append(res, exportSexp(sexp))
		}
		return res

	case glisp.SexpHash:
		res := make(map[string]interface{}, 0)
		for _, pairs := range t {
			for _, sexp := range pairs {
				p := glisp.SexpPair(sexp)
				symbol := p.Head()
				tail := p.Tail()

				res[exportSexp(symbol).(string)] = exportSexp(tail)
			}
		}
		return res

	default:
		return val.SexpString()
	}
}

func sexpToString(val glisp.Sexp) string {
	return exportSexp(val).(string)
}

func (d lispDriver) String() string {
	return "lisp"
}

type ankoDriver struct {
	vm *anko.Env
}

func (d ankoDriver) Handle(e event.Event, fnName string) {
	_, err := d.RunUnsafe(fmt.Sprintf("%s(\"%s\", \"%s\", \"%s\", %s)", fnName, e.Data["Code"], e.Data["Target"], e.Data["Nick"], strconv.Quote(e.Data["Message"].(string))))
	if err != nil {
		fmt.Println("An error occurred while executing the Anko handler", err)
	}
}

func (d ankoDriver) RunUnsafe(unsafe string) (val interface{}, err error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err == Halt {
			fmt.Println("Some Anko code took too long! Stopping after: ", duration)
			return
		}
		if e := recover(); e != nil {
			val = anko.NilValue
			err = e.(error)
		}
	}()

	// Anko chokes on carriage returns
	unsafe = strings.Replace(unsafe, "\r", "", -1)

	scanner := &anko_parser.Scanner{}
	scanner.Init(unsafe)
	stmts, e := anko_parser.Parse(scanner)
	if e != nil {
		val = nil
		err = e
		return
	}

	done := make(chan bool)
	go func() {
		v, e := anko.Run(stmts, d.vm)
		select {
		case _ = <-done:
			return
		default:
			done <- true
			val = v.Interface()
			err = e
		}
	}()

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		select {
		case _ = <-done:
			return
		default:
			done <- true
			val = nil
			err = Halt
		}
	}()

	_ = <-done
	close(done)

	return
}

func (d ankoDriver) String() string {
	return "anko"
}
