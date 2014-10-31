package irc

import (
	"fmt"
	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/squircy/event"
)

const (
	PrivmsgEvent    event.EventType = "irc.PRIVMSG"
	NoticeEvent                     = "irc.NOTICE"
	ConnectEvent                    = "irc.CONNECT"
	ConnectingEvent                 = "irc.CONNECTING"
	DisconnectEvent                 = "irc.DISCONNECT"
	IrcEvent                        = "irc.WILDCARD"
)

func bindEvents(mgr *IrcConnectionManager, e event.EventManager) {
	mgr.conn.AddCallback("*", func(ev *ircevent.Event) {
		e.Trigger(IrcEvent, newEventData(ev))
		e.Trigger(event.EventType("irc."+ev.Code), newEventData(ev))
	})

	mgr.conn.AddCallback("001", func(ev *ircevent.Event) {
		fmt.Println("Connected")
		mgr.status = Connected
		e.Trigger(ConnectEvent, newEventData(ev))
	})

	mgr.conn.AddCallback("ERROR", func(ev *ircevent.Event) {
		if mgr.status != Disconnected {
			mgr.Quit()
		}
		e.Trigger(DisconnectEvent, newEventData(ev))
	})
}

func triggerConnecting(e event.EventManager) {
	e.Trigger(ConnectingEvent, nil)
}

func newEventData(ev *ircevent.Event) map[string]interface{} {
	return map[string]interface{}{
		"Event":   ev,
		"Code":    ev.Code,
		"Message": ev.Message(),
		"Nick":    ev.Nick,
		"Target":  ev.Arguments[0],
	}
}
