package script

import (
	"fmt"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"log"
)

type ScriptManager struct {
	e            event.EventManager
	jsDriver     javascriptDriver
	luaDriver    luaDriver
	lispDriver   lispDriver
	ankoDriver   ankoDriver
	httpHelper   httpHelper
	configHelper configHelper
	ircHelper    ircHelper
	dataHelper   dataHelper
	scriptHelper scriptHelper
	repo         ScriptRepository
	l            *log.Logger
}

func NewScriptManager(repo ScriptRepository, l *log.Logger, e event.EventManager, ircmanager *irc.IrcConnectionManager, config *config.Configuration) *ScriptManager {
	mgr := ScriptManager{
		e,
		javascriptDriver{},
		luaDriver{},
		lispDriver{},
		ankoDriver{},
		httpHelper{},
		configHelper{config},
		ircHelper{ircmanager},
		dataHelper{make(map[string]interface{})},
		scriptHelper{},
		repo,
		l,
	}
	mgr.init()

	return &mgr
}

func (m *ScriptManager) RunUnsafe(t ScriptType, code string) (result interface{}, err error) {
	err = nil
	result = nil

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
			return
		}
	}()

	defer func() {
		if err != nil {
			fmt.Println("An error occurred: ", err)
		}
	}()
	var d scriptDriver
	switch {
	case t == Javascript:
		d = m.jsDriver

	case t == Lua:
		d = m.luaDriver

	case t == Lisp:
		d = m.lispDriver

	case t == Anko:
		d = m.ankoDriver

	default:
		err = UnknownScriptType
		return
	}

	result, err = d.RunUnsafe(code)

	return
}

func (m *ScriptManager) ReInit() {
	m.init()
}

func (m *ScriptManager) init() {
	m.e.ClearAll()

	m.jsDriver.vm   = newJavascriptVm(m)
	m.luaDriver.vm  = newLuaVm(m)
	m.lispDriver.vm = newLispVm(m)
	m.ankoDriver.vm = newAnkoVm(m)

	m.scriptHelper = scriptHelper{m.e, m.jsDriver, m.luaDriver, m.lispDriver, m.ankoDriver, make(map[string]event.EventHandler, 0)}

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		m.l.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
