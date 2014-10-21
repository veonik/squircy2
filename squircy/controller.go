package squircy

import (
	"github.com/martini-contrib/render"
)

func indexAction(r render.Render) {
	r.HTML(200, "index", nil)
}
