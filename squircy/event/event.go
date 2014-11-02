package event

import (
	"errors"
	"github.com/codegangsta/inject"
	"reflect"
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
	e.events = make(map[EventType][]reflect.Value, 0)
}

func (e *eventManager) Trigger(eventName EventType, data map[string]interface{}) {
	handlers, ok := e.events[eventName]
	if !ok {
		return
	}

	event := Event{eventName, data}

	c := inject.New()
	c.SetParent(e.injector)
	c.Map(event)

	for _, handler := range handlers {
		c.Invoke(handler.Interface())
	}

	if wildcardHandlers, ok := e.events[AllEvents]; ok {
		for _, handler := range wildcardHandlers {
			c.Invoke(handler.Interface())
		}
	}
}
