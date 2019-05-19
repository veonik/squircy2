package plugin

import (
	"path/filepath"
	"plugin"

	"github.com/codegangsta/inject"
	"github.com/pkg/errors"
)

type InitializerFunc = func(inject.Injector) error

type Manager struct {
	basePath string
	plugins  []string
}

func NewManager(basePath string, plugins ...string) *Manager {
	return &Manager{
		basePath: basePath,
		plugins:  plugins,
	}
}

func (m *Manager) Configure(i inject.Injector) error {
	for _, p := range m.plugins {
		fp := filepath.Join(m.basePath, p+".so")
		pl, err := plugin.Open(fp)
		if err != nil {
			return err
		}
		in, err := pl.Lookup("Initialize")
		if err != nil {
			return err
		}
		fn, ok := in.(InitializerFunc)
		if !ok {
			return errors.Errorf("init function for plugin %s is of incorrect type %T, expected type InitializerFunc = func(inject.Injector) error", p, in)
		}
		if err = fn(i); err != nil {
			return err
		}
	}
	return nil
}
