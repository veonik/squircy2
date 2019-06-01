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

func (m *Manager) Configure(i inject.Injector) []error {
	var errs []error
	for _, p := range m.plugins {
		fp := filepath.Join(m.basePath, p+".so")
		pl, err := plugin.Open(fp)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "plugin configuration failed (%s)", p))
			continue
		}
		in, err := pl.Lookup("Initialize")
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "plugin configuration failed (%s)", p))
			continue
		}
		fn, ok := in.(InitializerFunc)
		if !ok {
			err := errors.Errorf("init function for plugin %s is of incorrect type %T, expected type InitializerFunc = func(inject.Injector) error", p, in)
			errs = append(errs, errors.Wrapf(err, "plugin configuration failed (%s)", p))
			continue
		}
		if err = fn(i); err != nil {
			errs = append(errs, errors.Wrapf(err, "plugin configuration failed (%s)", p))
			continue
		}
	}
	if len(errs) > 0 {
		return append([]error{}, errs...)
	}
	return nil
}
