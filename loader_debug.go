// +build debug

package squircy2

import (
	"os"

	"github.com/go-martini/martini"
	"github.com/tyler-sommer/stick"
)

var rootDir string

func init() {
	r, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	rootDir = r
}

func newTemplateLoader() stick.Loader {
	return stick.NewFilesystemLoader(rootDir + "/views")
}

func newStaticHandler() martini.Handler {
	return martini.Static(rootDir+"/public", martini.StaticOptions{
		SkipLogging: true,
	})
}
