package squircy

import (
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/antage/eventsource"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"net/http"
	"strconv"
)

func configureWeb(manager *Manager, conf *config.Configuration) {
	manager.Handlers(
		martini.Static(conf.RootPath+"/public", martini.StaticOptions{
			SkipLogging: true,
		}),
		render.Renderer(render.Options{
			Directory:  conf.RootPath + "/views",
			Layout:     "layout",
			Extensions: []string{".tmpl", ".html"},
		}))
	manager.Get("/event", func(es eventsource.EventSource, w http.ResponseWriter, r *http.Request) {
		es.ServeHTTP(w, r)
	})
	manager.Get("/", indexAction)
	manager.Get("/status", statusAction)
	manager.Group("/manage", func(r martini.Router) {
		r.Get("", manageAction)
		r.Post("/update", manageUpdateAction)
	})
	manager.Post("/connect", connectAction)
	manager.Post("/disconnect", disconnectAction)
	manager.Group("/script", func(r martini.Router) {
		r.Get("", scriptAction)
		r.Post("/reinit", scriptReinitAction)
		r.Get("/new", newScriptAction)
		r.Post("/create", createScriptAction)
		r.Get("/:id/edit", editScriptAction)
		r.Post("/:id/update", updateScriptAction)
		r.Post("/:id/remove", removeScriptAction)
	})
	manager.Group("/repl", func(r martini.Router) {
		r.Get("", replAction)
		r.Post("/execute", replExecuteAction)
	})
}

func indexAction(r render.Render, t *eventTracer) {
	r.HTML(200, "index", map[string]interface{}{
		"terminal": t.History(OutputEvent),
		"irc":      t.History(irc.IrcEvent),
	})
}

type appStatus struct {
	Status irc.ConnectionStatus
}

func statusAction(r render.Render, mgr *irc.IrcConnectionManager) {
	r.JSON(200, appStatus{mgr.Status()})
}

func scriptAction(r render.Render, repo script.ScriptRepository) {
	scripts := repo.FetchAll()

	r.HTML(200, "script/index", map[string]interface{}{"scripts": scripts})
}

func scriptReinitAction(r render.Render, mgr *script.ScriptManager) {
	mgr.ReInit()

	r.JSON(200, nil)
}

func newScriptAction(r render.Render) {
	r.HTML(200, "script/new", nil)
}

func createScriptAction(r render.Render, repo script.ScriptRepository, request *http.Request) {
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(script.Script{0, script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func editScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := repo.Fetch(int(id))

	r.HTML(200, "script/edit", map[string]interface{}{"Script": script})
}

func updateScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(script.Script{int(id), script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func removeScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	repo.Delete(int(id))

	r.JSON(200, nil)
}

func replAction(r render.Render) {
	r.HTML(200, "repl/index", nil)
}

func replExecuteAction(r render.Render, manager *script.ScriptManager, request *http.Request) {
	code := request.FormValue("script")
	sType := script.ScriptType(request.FormValue("scriptType"))

	res, err := manager.RunUnsafe(sType, code)
	var errStr interface{}
	if err != nil {
		errStr = err.Error()
	}
	r.JSON(200, map[string]interface{}{
		"res": res,
		"err": errStr,
	})
}

func connectAction(r render.Render, mgr *irc.IrcConnectionManager) {
	mgr.Connect()

	r.JSON(200, nil)
}

func disconnectAction(r render.Render, mgr *irc.IrcConnectionManager) {
	mgr.Quit()

	r.JSON(200, nil)
}

func manageAction(r render.Render, config *config.Configuration) {
	r.HTML(200, "manage/edit", map[string]interface{}{
		"Config": config,
	})
}

func manageUpdateAction(r render.Render, database *db.DB, conf *config.Configuration, request *http.Request) {
	conf.Network = request.FormValue("network")
	conf.Nick = request.FormValue("nick")
	conf.Username = request.FormValue("username")
	conf.OwnerNick = request.FormValue("owner_nick")
	conf.OwnerHost = request.FormValue("owner_host")

	config.SaveConfig(database, conf)

	r.Redirect("/manage", 302)
}
