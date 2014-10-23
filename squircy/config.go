package squircy

import (
	"encoding/json"
	"go/build"
	"os"
)

const basePkg = "github.com/tyler-sommer/squircy2"

type Configuration struct {
	Network       string
	Nick          string
	Username      string
	Password      string
	Channel       string
	OwnerNick     string
	OwnerHost     string
	RedisHost     string
	RedisDatabase int
	RootPath      string
}

func NewConfiguration(fname string) (config *Configuration) {
	config = &Configuration{}

	file, err := os.Open(fname)
	if err != nil {
		panic("Could not open configuration: " + err.Error())
	}

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(config); err != nil {
		panic("Could not decode configuration: " + err.Error())
	}
	compile(config)
	return
}

func NewDefaultConfiguration() (config *Configuration) {
	config = &Configuration{
		"irc.freenode.net:6667",
		"mrsquishy",
		"mrjones",
		"",
		"#squishyslab",
		"",
		"",
		"127.0.0.1:6379",
		0,
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
