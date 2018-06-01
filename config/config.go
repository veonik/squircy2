package config // import "github.com/veonik/squircy2/config"

import (
	"github.com/HouzuoGuo/tiedot/db"
)

type Configuration struct {
	ID int

	RootPath string

	ScriptsAsFiles bool   // Store scripts on the filesystem
	ScriptsPath    string // Path to script storage
	EnableFileAPI  bool   // Enable the filesystem API in scripts
	FileAPIRoot    string // Only allow filesystem ops in this directory

	Network     string // Hostname and port, format: hostname:1234
	TLS         bool   // Enable TLS/SSL for IRC
	AutoConnect bool
	Nick        string
	Username    string

	SASL         bool // Enable SASL authentication
	SASLUsername string
	SASLPassword string

	OwnerNick string
	OwnerHost string

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
		ID:             -1,
		RootPath:       rootPath,
		ScriptsAsFiles: false,
		ScriptsPath:    "",
		EnableFileAPI:  false,
		FileAPIRoot:    "",
		Network:        "irc.freenode.net:6667",
		TLS:            false,
		AutoConnect:    false,
		Nick:           "mrsquishy",
		Username:       "squishyj",
		SASL:           false,
		SASLUsername:   "",
		SASLPassword:   "",
		OwnerNick:      "",
		OwnerHost:      "",
		WebInterface:   true,
		HTTPHostPort:   ":3000",
		HTTPS:          false,
		RequireHTTPS:   false,
		SSLCertFile:    "",
		SSLCertKey:     "",
		SSLHostPort:    "",
		HTTPAuth:       false,
		AuthUsername:   "",
		AuthPassword:   "",
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
