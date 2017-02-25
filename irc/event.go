package irc

import (
	"time"

	ircevent "github.com/thoj/go-ircevent"
	"github.com/tyler-sommer/squircy2/event"
)

const (
	PrivmsgEvent    event.EventType = "irc.PRIVMSG"
	NoticeEvent                     = "irc.NOTICE"
	ConnectEvent                    = "irc.CONNECT"
	ConnectingEvent                 = "irc.CONNECTING"
	DisconnectEvent                 = "irc.DISCONNECT"
	IrcEvent                        = "irc.WILDCARD"
)

func bindEvents(mgr *ConnectionManager, e event.EventManager) {
	mgr.conn.AddCallback("*", func(ev *ircevent.Event) {
		e.Trigger(IrcEvent, newEventData(ev))
		e.Trigger(event.EventType("irc."+ev.Code), newEventData(ev))
	})

	mgr.conn.AddCallback("001", func(ev *ircevent.Event) {
		mgr.status = Connected
		e.Trigger(ConnectEvent, newEventData(ev))
	})

	mgr.conn.AddCallback("ERROR", func(ev *ircevent.Event) {
		if mgr.status != Disconnected {
			mgr.Quit()
		}
		// TODO: Triggers disconnect twice, but once with the error details.
		e.Trigger(DisconnectEvent, newEventData(ev))
	})
	mgr.conn.AddCallback("PONG", func(ev *ircevent.Event) {
		mgr.lastPong = time.Now()
	})
}

func triggerConnecting(e event.EventManager) {
	e.Trigger(ConnectingEvent, nil)
}

func triggerDisconnected(e event.EventManager) {
	e.Trigger(DisconnectEvent, nil)
}

func newEventData(ev *ircevent.Event) map[string]interface{} {
	return map[string]interface{}{
		"User":    ev.User,
		"Host":    ev.Host,
		"Source":  ev.Source,
		"Code":    ev.Code,
		"Message": ev.Message(),
		"Nick":    ev.Nick,
		"Target":  ev.Arguments[0],
		"Raw":     ev.Raw,
		"Args":    ev.Arguments,
	}
}
