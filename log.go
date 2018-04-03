package squircy2

import (
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/tyler-sommer/squircy2/event"
)

const (
	LogEvent event.EventType = "log.OUTPUT"
)

func newLogger(evm event.EventManager) *log.Logger {
	hooks := make(log.LevelHooks)
	hooks.Add(&triggerHook{evm})
	return &log.Logger{
		Out:       os.Stderr,
		Formatter: new(log.TextFormatter),
		Hooks:     hooks,
		Level:     log.InfoLevel,
	}
}

type triggerHook struct {
	evm event.EventManager
}

func (l *triggerHook) Levels() []log.Level {
	return log.AllLevels
}
func (l *triggerHook) Fire(e *log.Entry) error {
	data := make(map[string]interface{})
	for k, v := range e.Data {
		data[k] = v
	}
	l.evm.Trigger(LogEvent, map[string]interface{}{
		"Time":    e.Time.String(),
		"Level":   e.Level.String(),
		"Message": e.Message,
		"Data":    data,
	})
	return nil
}
