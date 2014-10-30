package script

import (
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/robertkrimen/otto"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"strconv"
)

type scriptDriver interface {
	Handle(e event.Event, fnName string)
	String() string
}

type javascriptDriver struct {
	vm *otto.Otto
}

func (d javascriptDriver) Handle(e event.Event, fnName string) {
	d.vm.Interrupt = make(chan func(), 1)
	d.vm.Call(fnName, otto.NullValue(), e.Data["Code"], e.Data["Target"], e.Data["Nick"], e.Data["Message"])
}

func (d javascriptDriver) String() string {
	return "js"
}

type luaDriver struct {
	vm *lua.State
}

func (d luaDriver) Handle(e event.Event, fnName string) {
	d.vm.GetGlobal(fnName)
	d.vm.PushString(e.Data["Code"].(string))
	d.vm.PushString(e.Data["Target"].(string))
	d.vm.PushString(e.Data["Nick"].(string))
	d.vm.PushString(e.Data["Message"].(string))
	d.vm.Call(4, 0)
}

func (d luaDriver) String() string {
	return "lua"
}

type lispDriver struct{}

func (d lispDriver) Handle(e event.Event, fnName string) {
	runUnsafeLisp(fmt.Sprintf("(%s \"%s\" \"%s\" \"%s\" %s)", fnName, e.Data["Code"], e.Data["Target"], e.Data["Nick"], strconv.Quote(e.Data["Message"].(string))))
}

func (d lispDriver) String() string {
	return "lisp"
}
