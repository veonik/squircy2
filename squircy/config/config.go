package config

import (
	"github.com/HouzuoGuo/tiedot/db"
	"go/build"
	"os"
)

type Configuration struct {
	ID        int
	Network   string
	Nick      string
	Username  string
	OwnerNick string
	OwnerHost string
	RootPath  string
}

func NewDefaultConfiguration() (config *Configuration) {
	config = &Configuration{
		-1,
		"irc.freenode.net:6667",
		"mrsquishy",
		"mrjones",
		"",
		"",
		"",
	}
	compile(config)
	return
}

func compile(config *Configuration) {
	if len(config.RootPath) == 0 {
		config.RootPath = resolveRoot()
	}
	return
}

func resolveRoot() string {	
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return wd
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
