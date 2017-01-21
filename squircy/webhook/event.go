package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"log"
	"strings"
)

type WebhookEvent struct {
	Body        []byte
	ContentType string
	Signature   string
	Webhook     Webhook
}

// Process a webhook event
func (e *WebhookEvent) Process(mgr *irc.IrcConnectionManager) error {
	_, err := e.CheckPayloadSignature()
	if err != nil {
		log.Printf("Check payload failed, %s", err)
		return err
	}
	// Parse JSON
	var payload map[string]interface{}

	if strings.Contains(e.ContentType, "json") {
		decoder := json.NewDecoder(strings.NewReader(string(e.Body)))
		decoder.UseNumber()

		err := decoder.Decode(&payload)

		if err != nil {
			return errors.New("Invalid JSON")
		}
	} else {
		return errors.New("Invalid Content-Type")
	}
	log.Printf("Processing webhook with payload %+v", payload)
	// conn := mgr.Connection()
	// conn.Notice("chamal", e.Message)
	return nil
}

// SignatureError describes an invalid payload signature passed to Hook.
type SignatureError struct {
	Signature string
}

func (e *SignatureError) Error() string {
	if e == nil {
		return "<nil>"
	}
	log.Printf("invalid payload signature %s", e.Signature)
	return ""
}

// CheckPayloadSignature calculates and verifies SHA1 signature of the given payload
func (e *WebhookEvent) CheckPayloadSignature() (string, error) {
	if strings.HasPrefix(e.Signature, "sha1=") {
		signature := e.Signature[5:]

		mac := hmac.New(sha1.New, []byte(e.Webhook.Key))
		_, err := mac.Write(e.Body)
		if err != nil {
			return "", err
		}
		expectedMAC := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
			return expectedMAC, &SignatureError{signature}
		}
		return expectedMAC, err
	}
	return "", errors.New("Only sha1 signature handled")
}
