package squircy

import (
	"bufio"
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"log"
	"os"
	"reflect"
	"time"
)

func (man *Manager) LoopCli() {
	man.Invoke(loopCli)
}

func getConnection(manager *Manager) *ircevent.Connection {
	return manager.Injector.Get(reflect.TypeOf((*ircevent.Connection)(nil))).Interface().(*ircevent.Connection)
}

func getConnectionManager(manager *Manager) *irc.IrcConnectionManager {
	return manager.Injector.Get(reflect.TypeOf((*irc.IrcConnectionManager)(nil))).Interface().(*irc.IrcConnectionManager)
}

func loopCli(l *log.Logger, manager *Manager) {
	help := func() {
		fmt.Println(`Commands:

exit		Quits IRC, if connected, and exits the program
debug		Toggles debug on or off`)
		mgr := getConnectionManager(manager)
		if mgr.Status() != irc.Disconnected {
			fmt.Println("disconnect	Disconnect from IRC\n")
		} else {
			fmt.Println("connect		Connect to IRC\n")
		}
	}

	help()

	bin := bufio.NewReader(os.Stdin)
	for {
		switch str, _ := bin.ReadString('\n'); {
		case str == "exit\n" || str == "quit\n":
			mgr := getConnectionManager(manager)
			go mgr.Quit()
			time.Sleep(2 * time.Second)
			l.Println("Exiting")
			return

		case str == "debug\n":
			mgr := getConnectionManager(manager)
			if mgr.Status() == irc.Disconnected {
				fmt.Println("Not connected")
			} else {
				conn := getConnection(manager)
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
			mgr := getConnectionManager(manager)
			if mgr.Status() != irc.Disconnected {
				fmt.Println("Disconnecting...")
				mgr.Quit()
			} else {
				fmt.Println("Connecting...")
				mgr.Connect()
			}

		default:
			fmt.Print("Unknown input. ")
			help()
		}
	}
}
