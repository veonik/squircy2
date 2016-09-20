package script

import (
	"crypto/sha1"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
)

// timer represents a function call to be performed after a delay.
type timer struct {
	t      *time.Timer
	dur    time.Duration
	repeat bool
	call   otto.FunctionCall
}

// jsVM is a small wrapper around Otto VM, facilitating the event loop
// and setTimeout/setInterval functionality.
type jsVm struct {
	*otto.Otto

	registry map[*timer]*timer
	ready    chan *timer
	quit     chan struct{}
}

func newJavascriptVm(m *ScriptManager) *jsVm {
	jsVm := &jsVm{otto.New(), make(map[*timer]*timer), make(chan *timer), make(chan struct{})}
	getFnName := func(fn otto.Value) (name string) {
		if fn.IsFunction() {
			name = fmt.Sprintf("__Handler%x", sha1.Sum([]byte(fn.String())))
		} else {
			name = fn.String()
		}

		return
	}
	newTimer := func(call otto.FunctionCall, repeat bool) (*timer, otto.Value, error) {
		delay, _ := call.Argument(1).ToInteger()
		if delay <= 0 {
			delay = 1
		}

		res := &timer{nil, time.Duration(delay) * time.Millisecond, repeat, call}
		jsVm.registry[res] = res

		res.t = time.AfterFunc(res.dur, func() {
			jsVm.ready <- res
		})

		val, err := jsVm.ToValue(res)
		if err != nil {
			return nil, otto.UndefinedValue(), err
		}

		return res, val, nil
	}
	clearTimer := func(call otto.FunctionCall) otto.Value {
		ti, _ := call.Argument(0).Export()
		if ti, ok := ti.(*timer); ok {
			ti.t.Stop()
			delete(jsVm.registry, ti)
		}
		return otto.UndefinedValue()
	}

	jsVm.Set("setTimeout", func(call otto.FunctionCall) otto.Value {
		_, v, err := newTimer(call, false)
		if err != nil {
			return otto.UndefinedValue()
		}
		return v
	})
	jsVm.Set("setInterval", func(call otto.FunctionCall) otto.Value {
		_, v, err := newTimer(call, true)
		if err != nil {
			return otto.UndefinedValue()
		}
		return v
	})
	jsVm.Set("clearTimeout", clearTimer)
	jsVm.Set("clearInterval", clearTimer)
	jsVm.Set("Http", &m.httpHelper)
	jsVm.Set("Config", &m.configHelper)
	jsVm.Set("Data", &m.dataHelper)
	jsVm.Set("Irc", &m.ircHelper)
	jsVm.Set("Os", &m.osHelper)
	jsVm.Set("Math", &m.mathHelper)
	jsVm.Set("bind", func(call otto.FunctionCall) otto.Value {
		eventType := call.Argument(0).String()
		fn := call.Argument(1)
		fnName := getFnName(fn)
		if fn.IsFunction() {
			m.jsDriver.vm.Set(fnName, func(call otto.FunctionCall) otto.Value {
				fn.Call(call.This, call.ArgumentList)
				return otto.UndefinedValue()
			})
		}
		m.scriptHelper.Bind(Javascript, event.EventType(eventType), fnName)
		val, _ := otto.ToValue(fnName)
		return val
	})
	jsVm.Set("unbind", func(call otto.FunctionCall) otto.Value {
		eventType := call.Argument(0).String()
		fnName := getFnName(call.Argument(1))
		m.scriptHelper.Unbind(Javascript, event.EventType(eventType), fnName)
		return otto.UndefinedValue()
	})
	jsVm.Set("trigger", func(call otto.FunctionCall) otto.Value {
		eventType := call.Argument(0).String()
		dat, _ := call.Argument(1).Export()
		if dat == nil {
			dat = make(map[string]interface{}, 0)
		}
		m.scriptHelper.Trigger(event.EventType(eventType), dat.(map[string]interface{}))
		return otto.UndefinedValue()
	})
	jsVm.Set("use", func(call otto.FunctionCall) otto.Value {
		coll := call.Argument(0).String()

		// Todo: get the Database properly
		db := data.NewGenericRepository(m.repo.database, coll)
		obj, _ := jsVm.Object("({})")
		obj.Set("Save", func(call otto.FunctionCall) otto.Value {
			exp, _ := call.Argument(0).Export()
			var model data.GenericModel
			switch t := exp.(type) {
			case data.GenericModel:
				model = t

			case map[string]interface{}:
				model = data.GenericModel(t)
			}
			switch t := model["ID"].(type) {
			case int64:
				model["ID"] = int(t)

			case int:
				model["ID"] = t

			case float64:
				model["ID"] = int(t)
			}
			db.Save(model)

			id, _ := jsVm.ToValue(model["ID"])

			return id
		})
		obj.Set("Delete", func(call otto.FunctionCall) otto.Value {
			i, _ := call.Argument(0).ToInteger()
			db.Delete(int(i))

			res, _ := jsVm.ToValue(true)
			return res
		})
		obj.Set("Fetch", func(call otto.FunctionCall) otto.Value {
			i, _ := call.Argument(0).ToInteger()
			val := db.Fetch(int(i))
			v, err := jsVm.ToValue(val)

			if err != nil {
				panic(err)
			}

			return v
		})
		obj.Set("FetchAll", func(call otto.FunctionCall) otto.Value {
			vals := db.FetchAll()
			v, err := jsVm.ToValue(vals)

			if err != nil {
				m.l.Println("An error occurred: ", err)
			}

			return v
		})
		obj.Set("Index", func(call otto.FunctionCall) otto.Value {
			exp, _ := call.Argument(0).Export()
			cols := make([]string, 0)
			for _, val := range exp.([]interface{}) {
				cols = append(cols, val.(string))
			}
			db.Index(cols)

			return otto.UndefinedValue()
		})
		obj.Set("Query", func(call otto.FunctionCall) otto.Value {
			qry, _ := call.Argument(0).Export()
			vals := db.Query(qry)
			v, err := jsVm.ToValue(vals)

			if err != nil {
				m.l.Println("An error occurred: ", err)
			}

			return v
		})

		return obj.Value()
	})

	go jsVm.Loop()

	return jsVm
}

// Loop kicks off the VM's event loop.
func (vm *jsVm) Loop() {
	for {
		select {
		case ti := <-vm.ready:
			var args []interface{}
			if len(ti.call.ArgumentList) > 2 {
				tmp := ti.call.ArgumentList[2:]
				// args[1] will end up being "this" when the function is invoked,
				// so we need to offset each actual argument by one
				args = make([]interface{}, 2+len(tmp))
				for i, value := range tmp {
					args[i+2] = value
				}
			} else {
				args = make([]interface{}, 1)
			}
			args[0] = ti.call.ArgumentList[0]
			// Since we are calling "Function.call.call", we pass the function
			// to be called in as the first argument, which means args[1] will
			// end up being the function's "this" binding, and any further values
			// in args will be passed to the function as arguments.
			_, err := vm.Call("Function.call.call", nil, args...)
			if err == nil && ti.repeat {
				ti.t.Reset(ti.dur)
			} else {
				delete(vm.registry, ti)
			}
		case <-vm.quit:
			return
		}
	}
}
