package squircy

import (
	"github.com/go-martini/martini"
)

type Manager struct {
	*martini.ClassicMartini
	config *Configuration
}

func NewManager() *Manager {
	return &Manager{martini.Classic(),NewConfiguration("config.json")}
}