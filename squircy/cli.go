package squircy

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/peterh/liner"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
)

const (
	OutputEvent event.EventType = "cli.OUTPUT"
	InputEvent                  = "cli.INPUT"
)

func (man *Manager) LoopCli() {
	man.Invoke(loopCli)
}

func loopCli(conf *config.Configuration, l *log.Logger, ircmgr *irc.IrcConnectionManager, evm event.EventManager, scmgr *script.ScriptManager) {
	hist := filepath.Join(conf.RootPath, ".history")

	cli := liner.NewLiner()
	defer func() {
		if f, err := os.Create(hist); err == nil {
			cli.WriteHistory(f)
			f.Close()
		}
		cli.Close()
	}()

	if f, err := os.Open(hist); err == nil {
		cli.ReadHistory(f)
		f.Close()
	}

	commands := []string{"exit", "reload", "repl", "debug", "connect", "reconnect", "disconnect"}
	cli.SetCompleter(func(line string) []string {
		res := []string{}
		for _, v := range commands {
			if strings.HasPrefix(v, line) {
				res = append(res, v)
			}
		}
		return res
	})

	help := func() {
		l.Println(`Commands:

exit		Quits IRC, if connected, and exits the program
reload		Reload the scripting engine
repl		Start a JavaScript REPL
debug		Toggles debug on or off`)
		if ircmgr.Status() != irc.Disconnected {
			l.Println(`disconnect	Disconnect from IRC
reconnect	Force a reconnection to IRC`)
		} else {
			l.Println("connect		Connect to IRC")
		}
		l.Println()
	}

	help()

	for {
		cmd, err := cli.Prompt("cmd> ")
		if err != nil {
			// TODO: do something useful
			continue
		}
		evm.Trigger(InputEvent, map[string]interface{}{
			"Message": cmd,
		})
		cli.AppendHistory(cmd)
		switch {
		case cmd == "exit" || cmd == "quit":
			go ircmgr.Quit()
			time.Sleep(2 * time.Second)
			l.Println("Exiting")
			return

		case cmd == "debug":
			debugging := !ircmgr.Debug()
			ircmgr.SetDebug(debugging)
			if debugging {
				l.Println("Debug ENABLED")
			} else {
				l.Println("Debug DISABLED")
			}

		case cmd == "reload":
			l.Println("Reloading...")
			scmgr.ReInit()
			l.Println("Reloaded.")

		case cmd == "repl":
			scmgr.Repl()

		case cmd == "connect" || cmd == "disconnect":
			if ircmgr.Status() != irc.Disconnected {
				l.Println("Disconnecting...")
				ircmgr.Quit()
			} else {
				l.Println("Connecting...")
				ircmgr.Connect()
			}

		case cmd == "reconnect":
			l.Println("Reconnecting...")
			ircmgr.Reconnect()

		default:
			l.Print("Unknown input. ")
			help()
		}
	}
}

func configureLog(manager *Manager, evm event.EventManager) {
	hist := &triggerLogger{evm}
	out := io.MultiWriter(os.Stdout, hist)
	logger := log.New(out, "", 0)
	manager.Map(logger)
}

type triggerLogger struct {
	evm event.EventManager
}

func (l *triggerLogger) Write(p []byte) (n int, err error) {
	l.evm.Trigger(OutputEvent, map[string]interface{}{
		"Message": string(p),
	})

	return
}
