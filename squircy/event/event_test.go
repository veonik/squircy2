package event

import (
	"github.com/codegangsta/inject"
	"testing"
)

const TestEvent EventType = "test"

func TestEventHandler(t *testing.T) {
	invoker := inject.New()
	manager := NewEventManager(&invoker)

	test := false
	manager.Bind(TestEvent, func(event Event) {
		test = true
	})

	manager.Trigger(TestEvent, nil)

	if !test {
		t.Error("Failed to trigger event")
	}
}

func TestInvalidHandler(t *testing.T) {
	failed := false
	defer func() {
		if err := recover(); err == InvalidHandler {
			failed = true
		}
	}()

	invoker := inject.New()
	manager := NewEventManager(&invoker)

	manager.Bind(TestEvent, nil)

	if !failed {
		t.Error("Failed to trigger error")
	}
}
