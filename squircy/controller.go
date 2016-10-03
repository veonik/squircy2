package squircy

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/antage/eventsource"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"github.com/tyler-sommer/stick"
)

type stickHandler struct {
	env *stick.Env
	res http.ResponseWriter
}

func (h *stickHandler) HTML(status int, name string, ctx map[string]stick.Value) {
	h.res.WriteHeader(200)
	err := h.env.Execute(name, h.res, ctx)
	if err != nil {
		fmt.Println(err)
	}
}

func newStickHandler() martini.Handler {
	env := stick.New(newTemplateLoader())
	env.Functions["escape"] = func(ctx stick.Context, args ...stick.Value) stick.Value {
		if len(args) < 1 {
			return nil
		}
		return html.EscapeString(stick.CoerceString(args[0]))
	}
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		c.Map(&stickHandler{env, res})
	}
}

func configureWeb(manager *Manager) {
	manager.Handlers(
		newStaticHandler(),
		newStickHandler(),
		render.Renderer(),
	)
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
		r.Post("/:id/toggle", toggleScriptAction)
	})
	manager.Group("/repl", func(r martini.Router) {
		r.Get("", replAction)
		r.Post("/execute", replExecuteAction)
	})
}

func indexAction(s *stickHandler, t *eventTracer) {
	s.HTML(200, "index.html.twig", map[string]stick.Value{
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

func scriptAction(s *stickHandler, repo script.ScriptRepository) {
	scripts := repo.FetchAll()

	s.HTML(200, "script/index.html.twig", map[string]stick.Value{"scripts": scripts})
}

func scriptReinitAction(r render.Render, mgr *script.ScriptManager) {
	mgr.ReInit()

	r.JSON(200, nil)
}

func newScriptAction(s *stickHandler) {
	s.HTML(200, "script/new.html.twig", nil)
}

func createScriptAction(r render.Render, repo script.ScriptRepository, request *http.Request) {
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(&script.Script{0, script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func editScriptAction(s *stickHandler, repo script.ScriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := repo.Fetch(int(id))

	s.HTML(200, "script/edit.html.twig", map[string]stick.Value{"script": script})
}

func updateScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	repo.Save(&script.Script{int(id), script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func removeScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	repo.Delete(int(id))

	r.JSON(200, nil)
}

func toggleScriptAction(r render.Render, repo script.ScriptRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := repo.Fetch(int(id))
	script.Enabled = !script.Enabled
	repo.Save(script)

	r.JSON(200, nil)
}

func replAction(s *stickHandler) {
	s.HTML(200, "repl/index.html.twig", nil)
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

func manageAction(s *stickHandler, config *config.Configuration) {
	s.HTML(200, "manage/edit.html.twig", map[string]stick.Value{
		"config": config,
	})
}

func manageUpdateAction(r render.Render, database *db.DB, conf *config.Configuration, request *http.Request) {
	conf.Network = request.FormValue("network")
	conf.Nick = request.FormValue("nick")
	conf.Username = request.FormValue("username")
	conf.OwnerNick = request.FormValue("owner_nick")
	conf.OwnerHost = request.FormValue("owner_host")
	conf.TLS = (request.FormValue("tls") == "on")

	config.SaveConfig(database, conf)

	r.Redirect("/manage", 302)
}
