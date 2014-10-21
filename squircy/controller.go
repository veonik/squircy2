package squircy

import (
	"github.com/martini-contrib/render"
	"github.com/thoj/go-ircevent"
)

func indexAction(r render.Render) {
	r.HTML(200, "index", nil)
}

type appStatus struct {
	Connected bool
}

func statusAction(r render.Render, conn *irc.Connection) {
	status := false
	if conn.GetNick() != "" {
		status = true
	}

	r.JSON(200, appStatus{status})
}
