package webhook

import (
	"fmt"
	"github.com/tyler-sommer/squircy2/squircy/irc"
)

type WebhookEvent struct {
	Message string
}

// Process a webhook event
func (e *WebhookEvent) Process(mgr *irc.IrcConnectionManager) error {
	conn := mgr.Connection()
	fmt.Println("Processing webhook with ", e.Message)
	conn.Notice("chamal", e.Message)
	return nil
}
