package squircy

import (
	"errors"
	"fmt"
	"github.com/aarzilli/golua/lua"
	"github.com/robertkrimen/otto"
	"github.com/veonik/go-lisp/lisp"
	"time"
)

const maxExecutionTime = 2 // in seconds
var halt = errors.New("Execution limit exceeded")

func runUnsafeJavascript(vm *otto.Otto, unsafe string) (otto.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.Interrupt = make(chan func(), 1)

	go func() {
		time.Sleep(maxExecutionTime * time.Second)
		vm.Interrupt <- func() {
			panic(halt)
		}
	}()

	return vm.Run(unsafe)
}

func runUnsafeLua(vm *lua.State, unsafe string) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if err := recover(); err != nil {
			if err == halt {
				fmt.Println("Some code took too long! Stopping after: ", duration)
			}
			panic(err)
		}
	}()

	vm.SetExecutionLimit(maxExecutionTime * (1 << 26))
	err := vm.DoString(unsafe)

	if err != nil && err.Error() == "Lua execution quantum exceeded" {
		panic(halt)
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
				panic(halt)
			}
			panic(err)
		}
	}()

	lisp.SetExecutionLimit(maxExecutionTime * (1 << 15))
	return lisp.EvalString(unsafe)
}
