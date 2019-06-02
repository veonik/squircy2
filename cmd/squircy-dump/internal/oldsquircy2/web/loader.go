// +build !debug

package web

import (
	"bytes"
	"github.com/veonik/squircy2/web/generated"
	"io"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/staticbin"
	"github.com/tyler-sommer/stick"
)

var templateMapping = map[string]generatedTemplate{}

type stringTemplate struct {
	name     string
	contents string
}

func (t *stringTemplate) Name() string {
	return t.name
}

func (t *stringTemplate) Contents() io.Reader {
	return bytes.NewBufferString(t.contents)
}

func newTemplateLoader() stick.Loader {
	return &assetLoader{}
}

type assetLoader struct{}

func (l *assetLoader) Load(name string) (stick.Template, error) {
	res, err := generated.Asset("web/views/" + name)
	if err != nil {
		return nil, err
	}
	return &stringTemplate{name, string(res)}, nil
}

func newStaticHandler() martini.Handler {
	return staticbin.Static("public", generated.Asset)
}
