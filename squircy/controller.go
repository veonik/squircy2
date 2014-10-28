package squircy

import (
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"net/http"
	"strconv"
)

func indexAction(r render.Render) {
	r.HTML(200, "index", nil)
}

type appStatus struct {
	Connected  bool
	Connecting bool
}

func statusAction(r render.Render, mgr *IrcConnectionManager) {
	r.JSON(200, appStatus{mgr.Connected(), mgr.Connecting()})
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
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := repo.Fetch(int(id))

	r.HTML(200, "script/edit", map[string]interface{}{"Script": script})
}

func updateScriptAction(r render.Render, repo scriptRepository, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(persistentScript{int(id), scriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func removeScriptAction(r render.Render, repo scriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	repo.Delete(int(id))

	r.JSON(200, nil)
}

func executeScriptAction(r render.Render, repo scriptRepository, handler *ScriptHandler, params martini.Params) {
	index, _ := strconv.ParseInt(params["id"], 0, 64)

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

func connectAction(r render.Render, mgr *IrcConnectionManager) {
	mgr.Connect()

	r.JSON(200, nil)
}

func disconnectAction(r render.Render, mgr *IrcConnectionManager) {
	mgr.Quit()

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
