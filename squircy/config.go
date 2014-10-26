package squircy

import (
	"go/build"
)

const basePkg = "github.com/tyler-sommer/squircy2"

type Configuration struct {
	ID        int
	Network   string
	Nick      string
	Username  string
	Password  string
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
	p, err := build.Default.Import(basePkg, "", build.FindOnly)
	if err != nil {
		panic(err)
	}
	return p.Dir
}
