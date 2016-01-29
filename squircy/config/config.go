package config

import (
	"github.com/HouzuoGuo/tiedot/db"
	"go/build"
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

const basePkg = "github.com/tyler-sommer/squircy2"

func NewConfiguration(rootPath string) *Configuration {
	if len(rootPath) == 0 {
		rootPath = resolveRoot()
	}
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

func NewDefaultConfiguration() *Configuration {
	return NewConfiguration("")
}

func resolveRoot() string {
	p, err := build.Default.Import(basePkg, "", build.FindOnly)
	if err != nil {
		panic(err)
	}
	return p.Dir
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
