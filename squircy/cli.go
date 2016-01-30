package squircy

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
)

const (
	OutputEvent event.EventType = "cli.OUTPUT"
	InputEvent                  = "cli.INPUT"
)

func (man *Manager) LoopCli() {
	man.Invoke(loopCli)
}

func loopCli(l *log.Logger, ircmgr *irc.IrcConnectionManager, evm event.EventManager) {
	help := func() {
		fmt.Println(`Commands:

exit		Quits IRC, if connected, and exits the program
debug		Toggles debug on or off`)
		if ircmgr.Status() != irc.Disconnected {
			fmt.Println("disconnect	Disconnect from IRC\n")
		} else {
			fmt.Println("connect		Connect to IRC\n")
		}
	}

	help()

	bin := bufio.NewReader(os.Stdin)
	for {
		str, _ := bin.ReadString('\n')
		evm.Trigger(InputEvent, map[string]interface{}{
			"Message": str,
		})
		switch {
		case str == "exit\n" || str == "quit\n":
			go ircmgr.Quit()
			time.Sleep(2 * time.Second)
			l.Println("Exiting")
			return

		case str == "debug\n":
			if ircmgr.Status() == irc.Disconnected {
				fmt.Println("Not connected")
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

		case str == "connect\n" || str == "disconnect\n":
			if ircmgr.Status() != irc.Disconnected {
				fmt.Println("Disconnecting...")
				ircmgr.Quit()
			} else {
				fmt.Println("Connecting...")
				ircmgr.Connect()
			}

		default:
			fmt.Print("Unknown input. ")
			help()
		}
	}
}

func configureLog(manager *Manager, evm event.EventManager) {
	hist := &triggerLogger{evm}
	out := io.MultiWriter(os.Stdout, hist)
	logger := log.New(out, "", log.Ltime)
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
