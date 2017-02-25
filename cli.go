package squircy2

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/peterh/liner"
	"github.com/tyler-sommer/squircy2/irc"
)

func (m *Manager) LoopCLI() {
	hist := filepath.Join(m.conf.RootPath, ".history")

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
		if m.irc.Status() != irc.Disconnected {
			fmt.Println(`disconnect	Disconnect from IRC
reconnect	Force a reconnection to IRC`)
		} else {
			fmt.Println("connect		Connect to IRC")
		}
		fmt.Println()
	}

	help()

	oldLevel := m.logger.Level

	for {
		cmd, err := cli.Prompt("cmd> ")
		if err != nil {
			// TODO: do something useful
			continue
		}
		cli.AppendHistory(cmd)
		switch {
		case cmd == "exit" || cmd == "quit":
			m.irc.Quit()
			fmt.Println("Exiting")
			return

		case cmd == "listen":
			m.web.ListenAndServe()

		case cmd == "stop":
			err = m.web.StopListenAndServe()
			if err != nil {
				fmt.Println("Unable to stop HTTP server: ", err.Error())
				m.logger.Errorln(err.Error())
			}
			err = m.web.StopListenAndServeTLS()
			if err != nil {
				fmt.Println("Unable to stop HTTPS server: ", err.Error())
				m.logger.Errorln(err.Error())
			}

		case cmd == "debug":
			debugging := !m.irc.Debug()
			m.irc.SetDebug(debugging)
			if debugging {
				oldLevel = m.logger.Level
				m.logger.Level = log.DebugLevel
				m.logger.Infoln("Debug ENABLED")
			} else {
				m.logger.Level = oldLevel
				m.logger.Infoln("Debug DISABLED")
			}

		case cmd == "reload":
			m.logger.Infoln("Reloading...")
			m.scripts.ReInit()
			m.logger.Infoln("Reloaded.")

		case cmd == "repl":
			m.scripts.Repl()

		case cmd == "connect" || cmd == "disconnect":
			if m.irc.Status() != irc.Disconnected {
				m.logger.Infoln("Disconnecting...")
				m.irc.Quit()
			} else {
				m.logger.Infoln("Connecting...")
				m.irc.Connect()
			}

		case cmd == "reconnect":
			m.logger.Infoln("Reconnecting...")
			m.irc.Reconnect()

		default:
			fmt.Println("Unknown input. ")
			help()
		}
	}
}
