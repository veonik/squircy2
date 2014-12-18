package script

import (
	"crypto/sha1"
	"fmt"
	"github.com/robertkrimen/otto"
	"github.com/tyler-sommer/squircy2/squircy/data"
	"github.com/tyler-sommer/squircy2/squircy/event"
)

func newJavascriptVm(m *ScriptManager) *otto.Otto {
	getFnName := func(fn otto.Value) (name string) {
		if fn.IsFunction() {
			name = fmt.Sprintf("__Handler%x", sha1.Sum([]byte(fn.String())))
		} else {
			name = fn.String()
		}

		return
	}

	jsVm := otto.New()
	jsVm.Set("Http", &m.httpHelper)
	jsVm.Set("Config", &m.configHelper)
	jsVm.Set("Data", &m.dataHelper)
	jsVm.Set("Irc", &m.ircHelper)
	jsVm.Set("Os", &m.osHelper)
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
		data, _ := call.Argument(1).Export()
		if data == nil {
			data = make(map[string]interface{}, 0)
		}
		m.scriptHelper.Trigger(event.EventType(eventType), data.(map[string]interface{}))
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
			}
			db.Save(model)

			id, _ := jsVm.ToValue(model["ID"])

			return id
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
				panic(err)
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
				panic(err)
			}

			return v
		})

		return obj.Value()
	})

	return jsVm
}
