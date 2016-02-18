package config

import (
	"github.com/HouzuoGuo/tiedot/db"
)

type Configuration struct {
	ID        int
	Network   string
	Nick      string
	Username  string
	OwnerNick string
	OwnerHost string
	RootPath  string
	TLS       bool
}

func NewConfiguration(rootPath string) *Configuration {
	config := &Configuration{
		-1,
		"irc.freenode.net:6667",
		"mrsquishy",
		"squishyj",
		"",
		"",
		rootPath,
		false,
	}
	return config
}

func LoadConfig(database *db.DB, config *Configuration) {
	repo := configRepository{database}
	repo.fetchInto(config)
	SaveConfig(database, config)
}

func SaveConfig(database *db.DB, config *Configuration) {
	repo := configRepository{database}
	repo.save(config)
}
