package config

import (
	"github.com/HouzuoGuo/tiedot/db"
)

type Configuration struct {
	ID int

	RootPath string

	Network     string // Hostname and port, format: hostname:1234
	TLS         bool   // Enable TLS/SSL for IRC
	AutoConnect bool
	Nick        string
	Username    string
	OwnerNick   string
	OwnerHost   string

	WebInterface bool
	HTTPHostPort string // Hostname and port, format: hostname:1234

	HTTPS        bool
	RequireHTTPS bool
	SSLCertFile  string
	SSLCertKey   string
	SSLHostPort  string // Hostname and port, format: hostname:1234

	HTTPAuth     bool
	AuthUsername string
	AuthPassword string
}

func NewConfiguration(rootPath string) *Configuration {
	return &Configuration{
		ID:           -1,
		RootPath:     rootPath,
		Network:      "irc.freenode.net:6667",
		TLS:          false,
		AutoConnect:  false,
		Nick:         "mrsquishy",
		Username:     "squishyj",
		OwnerNick:    "",
		OwnerHost:    "",
		WebInterface: true,
		HTTPHostPort: ":3000",
		HTTPS:        false,
		RequireHTTPS: false,
		SSLCertFile:  "",
		SSLCertKey:   "",
		SSLHostPort:  "",
		HTTPAuth:     false,
		AuthUsername: "",
		AuthPassword: "",
	}
}

func LoadConfig(database *db.DB, conf *Configuration) {
	repo := configRepository{database}
	repo.fetchInto(conf)
	SaveConfig(database, conf)
}

func SaveConfig(database *db.DB, conf *Configuration) {
	repo := configRepository{database}
	repo.save(conf)
}
