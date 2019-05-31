package script // import "github.com/veonik/squircy2/script"

import (
	"os"

	"github.com/veonik/squircy2/data"

	log "github.com/sirupsen/logrus"

	"github.com/veonik/squircy2/config"
	"github.com/veonik/squircy2/event"
	"github.com/veonik/squircy2/irc"
)

type InitVMFunc = func(m *ScriptManager)

type ScriptManager struct {
	database     *data.DB
	events       event.EventManager
	conf         *config.Configuration
	driver       javascriptDriver
	httpHelper   httpHelper
	configHelper configHelper
	ircHelper    ircHelper
	scriptHelper scriptHelper
	mathHelper   mathHelper
	osHelper     osHelper
	fileHelper   fileHelper
	repo         ScriptRepository
	logger       log.FieldLogger

	initFns []InitVMFunc
}

func NewScriptManager(repo ScriptRepository, l log.FieldLogger, e event.EventManager, ircmanager *irc.ConnectionManager, config *config.Configuration, database *data.DB) *ScriptManager {
	mgr := &ScriptManager{
		database:     database,
		events:       e,
		conf:         config,
		driver:       javascriptDriver{nil, l},
		httpHelper:   httpHelper{},
		configHelper: configHelper{config},
		ircHelper:    ircHelper{ircmanager},
		scriptHelper: scriptHelper{},
		mathHelper:   mathHelper{},
		osHelper:     osHelper{},
		fileHelper:   fileHelper{config},
		repo:         repo,
		logger:       l,
	}
	return mgr
}

func (m *ScriptManager) AddOnInit(fn InitVMFunc) {
	m.initFns = append(m.initFns, fn)
}

func (m *ScriptManager) RunUnsafe(t ScriptType, code string) (result interface{}, err error) {
	err = nil
	result = nil

	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
		if err != nil {
			m.logger.Infoln("An error occurred: ", err)
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

// Export copies all scripts stored in the internal database to the filesystem.
func (m *ScriptManager) Export() error {
	if _, err := os.Stat(m.conf.ScriptsPath); err != nil {
		return err
	}
	fileRepo := newFileRepository(m.conf, m.logger)
	dbRepo := newDBRepository(m.database, m.logger)
	scripts := dbRepo.FetchAll()
	for _, script := range scripts {
		fileRepo.Save(script)
	}
	return nil
}

// Import copies all scripts stored on the filesystem to the internal database.
func (m *ScriptManager) Import() error {
	if _, err := os.Stat(m.conf.ScriptsPath); err != nil {
		return err
	}
	fileRepo := newFileRepository(m.conf, m.logger)
	dbRepo := newDBRepository(m.database, m.logger)
	scripts := fileRepo.FetchAll()
	for _, script := range scripts {
		dbRepo.Save(script)
	}
	return nil
}

func (m *ScriptManager) ReInit() {
	if m.driver.VM != nil {
		select {
		case <-m.driver.VM.quit:
			break
		default:
			close(m.driver.VM.quit)
		}
	}
	m.init()
}

func (m *ScriptManager) GetVM(t ScriptType) (*jsVm, error) {
	if t != Javascript {
		return nil, UnknownScriptType
	}
	return m.driver.VM, nil
}

func (m *ScriptManager) init() {
	m.events.ClearAll()

	m.driver.VM = newJavascriptVm(m)
	for _, fn := range m.initFns {
		fn(m)
	}

	m.scriptHelper = scriptHelper{m.events, m.driver, make(map[string]event.EventHandler, 0)}

	scripts := m.repo.FetchAll()
	for _, script := range scripts {
		if !script.Enabled {
			continue
		}
		m.logger.Debugln("Running", script.Type, "script", script.Title)
		if _, err := m.RunUnsafe(script.Type, script.Body); err != nil {
			m.logger.Warnln("Error running script", script.Title, err)
		}
	}
}
