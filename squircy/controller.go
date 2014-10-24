package squircy

import (
	"github.com/fzzy/radix/redis"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/thoj/go-ircevent"
	"net/http"
	"strconv"
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

func scriptAction(r render.Render, client *redis.Client) {
	repo := scriptRepository{client}
	scripts := repo.Fetch()

	r.HTML(200, "script/index", map[string]interface{}{"scripts": scripts})
}

func scriptReinitAction(r render.Render, h *ScriptHandler) {
	h.ReInit()

	r.JSON(200, nil)
}

func newScriptAction(r render.Render) {
	r.HTML(200, "script/new", nil)
}

func createScriptAction(r render.Render, client *redis.Client, request *http.Request) {
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo := scriptRepository{client}
	repo.Save(-1, persistentScript{scriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func editScriptAction(r render.Render, client *redis.Client, params martini.Params) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)

	repo := scriptRepository{client}
	script := repo.FetchIndex(int(index))

	r.HTML(200, "script/edit", map[string]interface{}{"Index": index, "Script": script})
}

func updateScriptAction(r render.Render, client *redis.Client, params martini.Params, request *http.Request) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo := scriptRepository{client}
	repo.Save(int(index), persistentScript{scriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func removeScriptAction(r render.Render, client *redis.Client, params martini.Params) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)

	repo := scriptRepository{client}
	repo.Delete(int(index))

	r.JSON(200, nil)
}

func connectAction(r render.Render, conn *irc.Connection, config *Configuration, h *HandlerCollection) {
	err := conn.Connect(config.Network)
	if err != nil {
		r.JSON(503, err)
	}

	h.bind(conn)

	r.JSON(200, nil)
}

func disconnectAction(r render.Render, conn *irc.Connection) {
	conn.Quit()

	r.JSON(200, nil)
}
