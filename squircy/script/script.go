package script

import (
	"fmt"
	"log"

	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
)

type ScriptManager struct {
	e            event.EventManager
	jsDriver     javascriptDriver
	httpHelper   httpHelper
	configHelper configHelper
	ircHelper    ircHelper
	dataHelper   dataHelper
	scriptHelper scriptHelper
	mathHelper   mathHelper
	osHelper     osHelper
	repo         ScriptRepository
	l            *log.Logger
}

func NewScriptManager(repo ScriptRepository, l *log.Logger, e event.EventManager, ircmanager *irc.IrcConnectionManager, config *config.Configuration) *ScriptManager {
	mgr := ScriptManager{
		e,
		javascriptDriver{},
		httpHelper{},
		configHelper{config},
		ircHelper{ircmanager},
		dataHelper{make(map[string]interface{})},
		scriptHelper{},
		mathHelper{},
		osHelper{},
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

	default:
		err = UnknownScriptType
		return
	}

	result, err = d.RunUnsafe(code)

	return
}

func (m *ScriptManager) ReInit() {
	close(m.jsDriver.vm.quit)
	m.init()
}

func (m *ScriptManager) init() {
	m.e.ClearAll()

	m.jsDriver.vm = newJavascriptVm(m)

	m.scriptHelper = scriptHelper{m.e, m.jsDriver, make(map[string]event.EventHandler, 0)}

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		if !script.Enabled {
			continue
		}
		m.l.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
