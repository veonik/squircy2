// v2 of the great example of SSE in go by @ismasan.
// includes fixes:
//    * infinite loop ending in panic
//    * closing a client twice
//    * potentially blocked listen() from closing a connection during multiplex step.
package eventsource

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

// the amount of time to wait when pushing a message to
// a slow client or a client that closed after `range clients` started.
const patience time.Duration = time.Second * 1

type Message struct {
	ID    string
	Event string
	Data  string
}

func (m *Message) String() string {
	var data bytes.Buffer
	if len(m.ID) > 0 {
		data.WriteString(fmt.Sprintf("id: %s\n", strings.Replace(m.ID, "\n", "", -1)))
	}
	if len(m.Event) > 0 {
		data.WriteString(fmt.Sprintf("event: %s\n", strings.Replace(m.Event, "\n", "", -1)))
	}
	if len(m.Data) > 0 {
		lines := strings.Split(m.Data, "\n")
		for _, line := range lines {
			data.WriteString(fmt.Sprintf("data: %s\n", line))
		}
	}
	data.WriteString("\n")
	return data.String()
}

type Broker struct {
	*log.Logger

	messages chan *Message

	newClients     chan chan *Message
	closingClients chan chan *Message
	clients        map[chan *Message]bool
}

func New() *Broker {
	broker := &Broker{
		Logger:         nil,
		messages:       make(chan *Message, 10),
		newClients:     make(chan chan *Message),
		closingClients: make(chan chan *Message),
		clients:        make(map[chan *Message]bool),
	}
	go broker.listen()
	return broker
}

func (broker *Broker) Notify(m *Message) {
	broker.messages <- m
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")

	client := make(chan *Message)
	defer func() {
		broker.closingClients <- client
	}()
	broker.newClients <- client

	notify := rw.(http.CloseNotifier).CloseNotify()

	for {
		select {
		case <-notify:
			return
		default:
			m := <-client
			rw.Write([]byte(m.String()))
			flusher.Flush()
		}
	}

}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			broker.clients[s] = true
			if broker.Logger != nil {
				broker.Printf("Client added. %d registered clients.\n", len(broker.clients))
			}

		case s := <-broker.closingClients:
			delete(broker.clients, s)
			if broker.Logger != nil {
				broker.Printf("Removed client. %d registered clients.\n", len(broker.clients))
			}

		case event := <-broker.messages:
			wg := &sync.WaitGroup{}
			wg.Add(len(broker.clients))
			for client := range broker.clients {
				go func(client chan *Message) {
					select {
					case client <- event:
					case <-time.After(patience):
						if broker.Logger != nil {
							broker.Println("Skipping client.")
						}
					}
					wg.Done()
				}(client)
			}
			wg.Wait()
		}
	}
}
