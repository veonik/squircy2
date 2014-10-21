package irc

import (
	"errors"
	"github.com/thoj/go-ircevent"
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
	conn *irc.Connection
	password string
	disabled bool
}

func NewNickservHandler(conn *irc.Connection, password string) (h *NickservHandler) {
	h = &NickservHandler{conn, password, false}
	
	return
}

func (h *NickservHandler) Id() string {
	return "nickserv"
}

func (h *NickservHandler) Matches(e *irc.Event) bool {
	return !h.disabled && strings.Contains(strings.ToLower(e.Message()), "identify") && e.User == "NickServ"
}

func (h *NickservHandler) Handle(e *irc.Event) {
	h.disabled = true
	h.conn.Privmsgf("NickServ", "IDENTIFY %s", h.password)
}