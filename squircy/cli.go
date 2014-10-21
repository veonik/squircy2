package squircy

import (
	"bufio"
	"fmt"
	"github.com/thoj/go-ircevent"
	"log"
	"os"
	"reflect"
	"time"
)

func (man *Manager) LoopCli() {
	l := man.Injector.Get(reflect.TypeOf((*log.Logger)(nil))).Interface().(*log.Logger)
	conn := man.Injector.Get(reflect.TypeOf((*irc.Connection)(nil))).Interface().(*irc.Connection)

	bin := bufio.NewReader(os.Stdin)
	for {
		switch str, _ := bin.ReadString('\n'); {
		case str == "exit\n" || str == "quit\n":
			conn.Quit()
			time.Sleep(2 * time.Second)
			l.Println("Exiting")
			return

		case str == "debug\n":
			debugging := !conn.Debug
			conn.Debug = debugging
			conn.VerboseCallbackHandler = debugging
			if debugging {
				l.Println("Debug ENABLED")
			} else {
				l.Println("Debug DISABLED")
			}

		default:
			fmt.Println(`Unknown input. Commands:

exit		Quits IRC and exits the program
debug		Toggles debug
`)
		}
	}
}
