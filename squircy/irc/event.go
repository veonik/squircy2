package irc

import (
	"github.com/thoj/go-ircevent"
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
	e.Clear(PrivmsgEvent)
	e.Clear(NoticeEvent)
	e.Clear(ConnectEvent)
	e.Clear(ConnectingEvent)
	e.Clear(DisconnectEvent)
	e.Clear(IrcEvent)

	mgr.conn.AddCallback("*", func(ev *irc.Event) {
		e.Trigger(IrcEvent, newEventData(mgr.conn, ev))
	})

	mgr.conn.AddCallback("001", func(ev *irc.Event) {
		mgr.status = Connected
		e.Trigger(ConnectEvent, newEventData(mgr.conn, ev))
	})

	mgr.conn.AddCallback("ERROR", func(ev *irc.Event) {
		if mgr.status != Disconnected {
			mgr.Quit()
		}
		e.Trigger(DisconnectEvent, newEventData(mgr.conn, ev))
	})
}

func newEventData(conn *irc.Connection, ev *irc.Event) map[string]interface{} {
	return map[string]interface{}{
		"Connection": conn,
		"Event":      ev,
		"Code":       ev.Code,
		"Message":    ev.Message(),
		"Nick":       ev.Nick,
		"Target":     ev.Arguments[0],
	}
}
