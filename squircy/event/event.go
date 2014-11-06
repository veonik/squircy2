package event

import (
	"errors"
	"github.com/codegangsta/inject"
	"reflect"
	"sync"
)

var InvalidHandler = errors.New("Invalid handler")

type EventType string

const AllEvents EventType = "event.WILDCARD"

type Event struct {
	Type EventType
	Data map[string]interface{}
}

type EventHandler interface{}

type EventManager interface {
	Bind(eventName EventType, handler EventHandler)
	Unbind(eventName EventType, handler EventHandler)
	Trigger(eventName EventType, data map[string]interface{})
	Clear(eventName EventType)
	ClearAll()
}

type eventManager struct {
	sync.Mutex
	injector inject.Injector
	events   map[EventType][]reflect.Value
}

func NewEventManager(injector inject.Injector) *eventManager {
	evm := new(eventManager)
	evm.events = make(map[EventType][]reflect.Value, 0)
	evm.injector = injector

	return evm
}

func (e *eventManager) Bind(eventName EventType, handler EventHandler) {
	fn := reflect.ValueOf(handler)
	if fn.Kind() != reflect.Func {
		panic(InvalidHandler)
	}

	handlers, ok := e.events[eventName]
	if !ok {
		handlers = make([]reflect.Value, 0)
	}

	e.events[eventName] = append(handlers, fn)
}

func (e *eventManager) Unbind(eventName EventType, handler EventHandler) {
	fn := reflect.ValueOf(handler)
	if fn.Kind() != reflect.Func {
		panic(InvalidHandler)
	}

	if handlers, ok := e.events[eventName]; ok {
		for i, other := range handlers {
			if other == fn {
				e.events[eventName] = append(e.events[eventName][:i], e.events[eventName][i+1:]...)
			}
		}
	}
}

func (e *eventManager) Clear(eventName EventType) {
	e.events[eventName] = make([]reflect.Value, 0)
}

func (e *eventManager) ClearAll() {
	wildcardHandlers, ok := e.events[AllEvents]

	e.events = make(map[EventType][]reflect.Value, 0)
	if ok {
		e.events[AllEvents] = wildcardHandlers
	}
}

func (e *eventManager) Trigger(eventName EventType, data map[string]interface{}) {
	e.Lock()
	defer e.Unlock()
	handlers, ok := e.events[eventName]
	wildcardHandlers, wok := e.events[AllEvents]
	if !ok && !wok {
		return
	}

	event := Event{eventName, data}

	c := inject.New()
	c.SetParent(e.injector)
	c.Map(event)

	if ok {
		for _, handler := range handlers {
			c.Invoke(handler.Interface())
		}
	}

	if wok {
		for _, handler := range wildcardHandlers {
			c.Invoke(handler.Interface())
		}
	}
}
