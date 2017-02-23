package squircy2

import (
	"io"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/tyler-sommer/squircy2/event"
)

func newLogger(evm event.EventManager) *log.Logger {
	hist := &triggerLogger{evm}
	out := io.MultiWriter(os.Stderr, hist)
	return &log.Logger{
		Out:       out,
		Formatter: new(log.TextFormatter),
		Hooks:     make(log.LevelHooks),
		Level:     log.InfoLevel,
	}
}

type triggerLogger struct {
	evm event.EventManager
}

func (l *triggerLogger) Write(p []byte) (n int, err error) {
	l.evm.Trigger(OutputEvent, map[string]interface{}{
		"Message": string(p),
	})

	return len(p), nil
}
