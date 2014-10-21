package squircy

import (
	"errors"
	"github.com/thoj/go-ircevent"
	"log"
	"strings"
)

const maxExecutionTime = 2 // in seconds
var halt = errors.New("Execution limit exceeded")

func replyTarget(e *irc.Event) string {
	if strings.HasPrefix(e.Arguments[0], "#") {
		return e.Arguments[0]
	} else {
		return e.Nick
	}
}

func parseCommand(msg string) (string, []string) {
	fields := strings.Fields(msg)
	if len(fields) < 1 {
		return "", nil
	}

	command := fields[0][1:]
	args := fields[1:]

	return command, args
}

type NickservHandler struct {
	conn     *irc.Connection
	log      *log.Logger
	handlers *HandlerCollection
	config   *Configuration
}

func newNickservHandler(conn *irc.Connection, log *log.Logger, handlers *HandlerCollection, config *Configuration) (h *NickservHandler) {
	h = &NickservHandler{conn, log, handlers, config}

	return
}

func (h *NickservHandler) Id() string {
	return "nickserv"
}

func (h *NickservHandler) Matches(e *irc.Event) bool {
	return strings.Contains(strings.ToLower(e.Message()), "identify") && e.User == "NickServ"
}

func (h *NickservHandler) Handle(e *irc.Event) {
	h.conn.Privmsgf("NickServ", "IDENTIFY %s", h.config.Password)
	h.log.Println("Identified with Nickserv")
	h.handlers.Remove(h)
}
