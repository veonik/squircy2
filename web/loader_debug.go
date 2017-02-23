// +build debug

package web

import (
	"os"

	"github.com/go-martini/martini"
	"github.com/tyler-sommer/stick"
)

var templateMapping = map[string]generatedTemplate{}

var rootDir string

func init() {
	r, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rootDir = r
}

func newTemplateLoader() stick.Loader {
	return stick.NewFilesystemLoader(rootDir + "/web/views")
}

func newStaticHandler() martini.Handler {
	return martini.Static(rootDir+"/web/public", martini.StaticOptions{
		SkipLogging: true,
	})
}
