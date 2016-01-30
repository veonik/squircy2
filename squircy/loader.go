// +build !debug

package squircy

import (
	"github.com/go-martini/martini"
	"github.com/martini-contrib/staticbin"
	"github.com/tyler-sommer/stick"
)

func newTemplateLoader() stick.Loader {
	return &assetLoader{}
}

type assetLoader struct{}

func (l *assetLoader) Load(name string) (string, error) {
	res, err := Asset("views/" + name)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

func newStaticHandler() martini.Handler {
	return staticbin.Static("public", Asset)
}
