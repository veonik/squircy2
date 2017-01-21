package webhook

import (
	"log"

	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/irc"
)

type WebhookManager struct {
	e    event.EventManager
	conf *config.Configuration
	repo WebhookRepository
	l    *log.Logger
}

func NewWebhookManager(repo WebhookRepository, l *log.Logger, e event.EventManager, ircmanager *irc.IrcConnectionManager, config *config.Configuration) *WebhookManager {
	mgr := WebhookManager{
		e,
		config,
		repo,
		l,
	}
	return &mgr
}
