package slack // import "github.com/veonik/squircy2/slack"

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/nlopes/slack"
	"github.com/veonik/squircy2/event"
)

type Config struct {
	APIToken string
}

type Manager struct {
	client *slack.Client
	events event.EventManager
	logger *log.Logger
}

func New(c Config, events event.EventManager, l *log.Logger) *Manager {
	m := &Manager{
		client: slack.New(c.APIToken),
		events: events,
		logger: l,
	}
	go func() {
		fmt.Println(m.Connect())
	}()
	return m
}

func (m *Manager) Connect() error {
	rtm := m.client.NewRTM()
	go rtm.ManageConnection()

	chs, err := m.client.GetChannels(true)
	if err != nil {
		return err
	}

	var chID string
	for _, ch := range chs {
		if ch.Name == "general" {
			chID = ch.ID
		}
		fmt.Println("Channel:", ch.Name, "  ID:", ch.ID)
	}

	for msg := range rtm.IncomingEvents {
		fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.HelloEvent:
			// Ignore hello

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			rtm.SendMessage(rtm.NewOutgoingMessage("Hello world", chID))

		case *slack.MessageEvent:
			fmt.Printf("Message: %v\n", ev)

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return nil

		default:
			fmt.Printf("Unexpected %T: %s - %v\n", msg, msg.Type, msg.Data)
		}
	}
	return nil
}

func (m *Manager) Disconnect() error {
	return nil
}
