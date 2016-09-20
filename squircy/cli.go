package squircy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"strings"

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

func loopCli(l *log.Logger, ircmgr *irc.IrcConnectionManager, evm event.EventManager, scmgr *script.ScriptManager) {
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

	bin := bufio.NewReader(os.Stdin)
	for {
		cmd, _ := bin.ReadString('\n')
		cmd = strings.TrimSuffix(cmd, "\n")
		evm.Trigger(InputEvent, map[string]interface{}{
			"Message": cmd,
		})
		switch {
		case cmd == "exit" || cmd == "quit":
			go ircmgr.Quit()
			time.Sleep(2 * time.Second)
			l.Println("Exiting")
			return

		case cmd == "debug":
			if ircmgr.Status() == irc.Disconnected {
				l.Println("Not connected")
			} else {
				conn := ircmgr.Connection()
				debugging := !conn.Debug
				conn.Debug = debugging
				conn.VerboseCallbackHandler = debugging
				if debugging {
					l.Println("Debug ENABLED")
				} else {
					l.Println("Debug DISABLED")
				}
			}

		case cmd == "reload":
			l.Println("Reloading...")
			scmgr.ReInit()
			l.Println("Reloaded.")

		case cmd == "repl":
			cursorHandler := func(ev event.Event) {
				fmt.Print("> ")
			}
			evm.Bind(OutputEvent, cursorHandler)
			fmt.Println("Starting javascript REPL...")
			fmt.Println("Type 'exit' and hit enter to exit the REPL.")
			for {
				fmt.Print("> ")
				str, _ := bin.ReadString('\n')
				if str == "exit\n" {
					evm.Unbind(OutputEvent, cursorHandler)
					fmt.Println("Closing REPL...")
					break
				}
				v, _ := scmgr.RunUnsafe(script.Javascript, str)
				fmt.Println(v)
			}

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
	manager.Map(hist)
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
