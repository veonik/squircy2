package squircy

import (
	"github.com/go-martini/martini"
)

type Manager struct {
	*martini.ClassicMartini
}

func NewManager() *Manager {
	return &Manager{martini.Classic()}
}