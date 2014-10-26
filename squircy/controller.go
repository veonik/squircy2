package squircy

import (
	"github.com/HouzuoGuo/tiedot/db"
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

func scriptAction(r render.Render, repo scriptRepository) {
	scripts := repo.FetchAll()

	r.HTML(200, "script/index", map[string]interface{}{"scripts": scripts})
}

func scriptReinitAction(r render.Render, h *ScriptHandler) {
	h.ReInit()

	r.JSON(200, nil)
}

func newScriptAction(r render.Render) {
	r.HTML(200, "script/new", nil)
}

func createScriptAction(r render.Render, repo scriptRepository, request *http.Request) {
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(persistentScript{0, scriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func editScriptAction(r render.Render, repo scriptRepository, params martini.Params) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)

	script := repo.Fetch(int(index))

	r.HTML(200, "script/edit", map[string]interface{}{"Index": script.ID, "Script": script})
}

func updateScriptAction(r render.Render, repo scriptRepository, params martini.Params, request *http.Request) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(persistentScript{int(index), scriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func removeScriptAction(r render.Render, repo scriptRepository, params martini.Params) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)

	repo.Delete(int(index))

	r.JSON(200, nil)
}

func executeScriptAction(r render.Render, repo scriptRepository, handler *ScriptHandler, params martini.Params) {
	index, _ := strconv.ParseInt(params["index"], 0, 32)

	script := repo.Fetch(int(index))

	switch {
	case script.Type == scriptJavascript:
		res, err := runUnsafeJavascript(handler.jsVm, script.Body)
		exres, _ := res.Export()
		r.JSON(200, map[string]interface{}{
			"res": exres,
			"err": err,
		})

	default:
		r.JSON(503, "Unsupported script type")
	}

}

func replAction(r render.Render) {
	r.HTML(200, "repl/index", nil)
}

func replExecuteAction(r render.Render, handler *ScriptHandler, request *http.Request) {
	script := request.FormValue("script")

	res, err := runUnsafeJavascript(handler.jsVm, script)
	exres, _ := res.Export()
	r.JSON(200, map[string]interface{}{
		"res": exres,
		"err": err,
	})
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

func manageAction(r render.Render, config *Configuration) {
	r.HTML(200, "manage/edit", map[string]interface{}{
		"Config": config,
	})
}

func manageUpdateAction(r render.Render, database *db.DB, config *Configuration, request *http.Request) {
	config.Network = request.FormValue("network")
	config.Nick = request.FormValue("nick")
	config.Username = request.FormValue("username")
	config.OwnerNick = request.FormValue("owner_nick")
	config.OwnerHost = request.FormValue("owner_host")

	saveConfig(database, config)

	r.Redirect("/manage", 302)
}
