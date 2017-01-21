package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tyler-sommer/squircy2/squircy/event"
)

type WebhookEvent struct {
	Body        []byte
	ContentType string
	Signature   string
	Webhook     *Webhook
}

// Process a webhook event
func (e *WebhookEvent) Process(evt event.EventManager) error {
	err := e.CheckPayloadSignature()
	if err != nil {
		log.Printf("Check payload failed, %s", err)
		return err
	}
	d := map[string]interface{}{
		"Body":        string(e.Body),
		"ContentType": e.ContentType,
		"Signature":   e.Signature,
		"Webhook":     e.Webhook.ID,
	}
	evt.Trigger(event.EventType(fmt.Sprintf("hook.%d", e.Webhook.ID)), d)
	evt.Trigger(event.EventType("hook.WILDCARD"), d)
	return nil
}

// CheckPayloadSignature calculates and verifies SHA1 signature of the given payload
func (e *WebhookEvent) CheckPayloadSignature() error {
	if strings.HasPrefix(e.Signature, "sha1=") {
		signature := e.Signature[5:]

		mac := hmac.New(sha1.New, []byte(e.Webhook.Key))
		_, err := mac.Write(e.Body)
		if err != nil {
			return err
		}
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return errors.New("Signature does not match")
		}
		return nil
	}
	return errors.New("Only sha1 signature handled")
}
