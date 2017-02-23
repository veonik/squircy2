package squircy2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/peterh/liner"
	"github.com/tyler-sommer/squircy2/config"
	"github.com/tyler-sommer/squircy2/event"
	"github.com/tyler-sommer/squircy2/irc"
	"github.com/tyler-sommer/squircy2/script"
)

const (
	OutputEvent event.EventType = "cli.OUTPUT"
	InputEvent                  = "cli.INPUT"
)

func (m *Manager) LoopCLI() {
	m.Invoke(loopCLI)
}

func loopCLI(conf *config.Configuration, l log.FieldLogger, ircmgr *irc.ConnectionManager, evm event.EventManager, scmgr *script.ScriptManager, m *Manager) {
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

	commands := []string{"exit", "reload", "repl", "debug", "connect", "reconnect", "disconnect", "listen", "stop"}
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
		fmt.Println(`Commands:

exit		Quits IRC, if connected, and exits the program
listen		Start HTTP (and HTTPS, if configured) server
stop		Stop HTTP and HTTPS server, if they are running
reload		Reload the scripting engine
repl		Start a JavaScript REPL
debug		Toggles debug on or off`)
		if ircmgr.Status() != irc.Disconnected {
			fmt.Println(`disconnect	Disconnect from IRC
reconnect	Force a reconnection to IRC`)
		} else {
			fmt.Println("connect		Connect to IRC")
		}
		fmt.Println()
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
			ircmgr.Quit()
			fmt.Println("Exiting")
			return

		case cmd == "listen":
			m.s.ListenAndServe()

		case cmd == "stop":
			err = m.s.StopListenAndServe()
			if err != nil {
				fmt.Println("Unable to stop HTTP server: ", err.Error())
				l.Errorln(err.Error())
			}
			err = m.s.StopListenAndServeTLS()
			if err != nil {
				fmt.Println("Unable to stop HTTPS server: ", err.Error())
				l.Errorln(err.Error())
			}

		case cmd == "debug":
			debugging := !ircmgr.Debug()
			ircmgr.SetDebug(debugging)
			if debugging {
				l.Infoln("Debug ENABLED")
			} else {
				l.Infoln("Debug DISABLED")
			}

		case cmd == "reload":
			l.Infoln("Reloading...")
			scmgr.ReInit()
			l.Infoln("Reloaded.")

		case cmd == "repl":
			scmgr.Repl()

		case cmd == "connect" || cmd == "disconnect":
			if ircmgr.Status() != irc.Disconnected {
				l.Infoln("Disconnecting...")
				ircmgr.Quit()
			} else {
				l.Infoln("Connecting...")
				ircmgr.Connect()
			}

		case cmd == "reconnect":
			l.Infoln("Reconnecting...")
			ircmgr.Reconnect()

		default:
			fmt.Println("Unknown input. ")
			help()
		}
	}
}
