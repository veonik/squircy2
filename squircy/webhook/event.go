package webhook

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"log"
	"strings"
)

type WebhookEvent struct {
	Body []byte
}

// Process a webhook event
func (e *WebhookEvent) Process(mgr *irc.IrcConnectionManager) error {
	log.Printf("Processing webhook with %+v", e.Body)
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
func (e *WebhookEvent) CheckPayloadSignature(key string, signature string) (string, error) {
	if strings.HasPrefix(signature, "sha1=") {
		signature = signature[5:]

		mac := hmac.New(sha1.New, []byte(key))
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
