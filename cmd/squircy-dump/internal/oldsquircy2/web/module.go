package web

import (
	"github.com/codegangsta/inject"
)

var modules = []ModuleFactory{}

func Register(m ModuleFactory) error {
	modules = append(modules, m)
	return nil
}

func MustRegister(m ModuleFactory) {
	if err := Register(m); err != nil {
		panic(err)
	}
}

func Configure(s *Server) error {
	for _, m := range modules {
		mod, err := m(s.Injector)
		if err != nil {
			return err
		}
		if err := mod.Configure(s); err != nil {
			return err
		}
	}
	return nil
}

type ModuleFactory func(injector inject.Injector) (Module, error)

type Module interface {
	Configure(s *Server) error
}
