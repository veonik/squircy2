package webhook

import (
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"log"
)

type WebhookEvent struct {
	Payload map[string]interface{}
}

// Process a webhook event
func (e *WebhookEvent) Process(mgr *irc.IrcConnectionManager) error {
	log.Printf("Processing webhook with %+v", e.Payload)
	// conn := mgr.Connection()
	// conn.Notice("chamal", e.Message)
	return nil
}
