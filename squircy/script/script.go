package script

import (
	"log"

	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
)

type ScriptManager struct {
	events       event.EventManager
	conf         *config.Configuration
	driver       javascriptDriver
	httpHelper   httpHelper
	configHelper configHelper
	ircHelper    ircHelper
	scriptHelper scriptHelper
	mathHelper   mathHelper
	osHelper     osHelper
	repo         ScriptRepository
	logger       *log.Logger
}

func NewScriptManager(repo ScriptRepository, l *log.Logger, e event.EventManager, ircmanager *irc.IrcConnectionManager, config *config.Configuration) *ScriptManager {
	mgr := ScriptManager{
		events:       e,
		conf:         config,
		driver:       javascriptDriver{nil, l},
		httpHelper:   httpHelper{},
		configHelper: configHelper{config},
		ircHelper:    ircHelper{ircmanager},
		scriptHelper: scriptHelper{},
		mathHelper:   mathHelper{},
		osHelper:     osHelper{},
		repo:         repo,
		logger:       l,
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
		if err != nil {
			m.logger.Println("An error occurred: ", err)
		}
	}()

	var d scriptDriver
	switch {
	case t == Javascript:
		d = m.driver

	default:
		err = UnknownScriptType
		return
	}

	result, err = d.RunUnsafe(code)

	return
}

func (m *ScriptManager) ReInit() {
	close(m.driver.vm.quit)
	m.init()
}

func (m *ScriptManager) init() {
	m.events.ClearAll()

	m.driver.vm = newJavascriptVm(m)

	m.scriptHelper = scriptHelper{m.events, m.driver, make(map[string]event.EventHandler, 0)}

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		if !script.Enabled {
			continue
		}
		m.logger.Println("Running", script.Type, "script", script.Title)
		m.RunUnsafe(script.Type, script.Body)
	}
}
