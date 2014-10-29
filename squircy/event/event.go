package event

import (
	"errors"
	"github.com/codegangsta/inject"
	"reflect"
)

var InvalidHandler = errors.New("Invalid handler")

type EventType string

type Event struct {
	Type EventType
	Data map[string]interface{}
}

type EventHandler interface{}

type EventManager interface {
	Bind(eventName EventType, handler EventHandler)
	Trigger(eventName EventType, data map[string]interface{})
	Clear(eventName EventType)
}

type eventManager struct {
	injector inject.Injector
	events   map[EventType][]EventHandler
}

func NewEventManager(injector inject.Injector) EventManager {
	return EventManager(&eventManager{injector, make(map[EventType][]EventHandler, 0)})
}

func (e *eventManager) Bind(eventName EventType, handler EventHandler) {
	if reflect.ValueOf(handler).Kind() != reflect.Func {
		panic(InvalidHandler)
	}

	handlers, ok := e.events[eventName]
	if !ok {
		handlers = make([]EventHandler, 0)
	}

	e.events[eventName] = append(handlers, handler)
}

func (e *eventManager) Clear(eventName EventType) {
	e.events[eventName] = make([]EventHandler, 0)
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
		c.Invoke(handler)
	}
}
