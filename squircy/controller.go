package squircy

import (
	"encoding/json"
	"fmt"
	"github.com/HouzuoGuo/tiedot/db"
	"github.com/antage/eventsource"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"github.com/tyler-sommer/squircy2/squircy/webhook"
	"github.com/tyler-sommer/stick"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
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

type webhookHandler struct {
	env *stick.Env
	res http.ResponseWriter
}

func newWebhookHandler() martini.Handler {
	env := stick.New(newTemplateLoader())
	env.Functions["escape"] = func(ctx stick.Context, args ...stick.Value) stick.Value {
		if len(args) < 1 {
			return nil
		}
		return html.EscapeString(stick.CoerceString(args[0]))
	}
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		c.Map(&webhookHandler{env, res})
	}
}

func configureWeb(manager *Manager) {
	manager.Handlers(
		newStaticHandler(),
		newStickHandler(),
		render.Renderer(),
		newWebhookHandler(),
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
	manager.Group("/webhook", func(r martini.Router) {
		r.Get("", webhookAction)
		r.Get("/new", newWebhookAction)
		r.Post("/create", createWebhookAction)
		r.Get("/:id/edit", editWebhookAction)
		r.Post("/:id/update", updateWebhookAction)
		r.Post("/:id/remove", removeWebhookAction)
		r.Post("/:id/toggle", toggleWebhookAction)
	})
	manager.Post("/webhooks", webhookReceiveAction)
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

// Manage webhook definitions
func webhookAction(s *stickHandler, repo webhook.WebhookRepository) {
	webhooks := repo.FetchAll()

	s.HTML(200, "webhook/index.html.twig", map[string]stick.Value{"webhooks": webhooks})
}

func newWebhookAction(s *stickHandler) {
	s.HTML(200, "webhook/new.html.twig", nil)
}

func createWebhookAction(r render.Render, repo webhook.WebhookRepository, request *http.Request) {
	sType := request.FormValue("type")
	body := request.FormValue("body")
	title := request.FormValue("title")
	url := request.FormValue("url")
	key := request.FormValue("key")
	sign := request.FormValue("signature")
	hook := &webhook.Webhook{0, script.ScriptType(sType), body, title, url, key, sign, true}
	repo.Save(hook)
	log.Printf("Created webhook %d", hook.ID)
	r.Redirect("/webhook", 302)
}

func editWebhookAction(s *stickHandler, repo webhook.WebhookRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	webhook := repo.Fetch(int(id))

	s.HTML(200, "webhook/edit.html.twig", map[string]stick.Value{"webhook": webhook})
}

func updateWebhookAction(r render.Render, repo webhook.WebhookRepository, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")
	url := request.FormValue("url")
	key := request.FormValue("key")
	sign := request.FormValue("signature")

	repo.Save(&webhook.Webhook{int(id), script.ScriptType(sType), body, title, url, key, sign, true})

	r.Redirect("/webhook", 302)
}

func removeWebhookAction(r render.Render, repo webhook.WebhookRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	repo.Delete(int(id))

	r.JSON(200, nil)
}

func toggleWebhookAction(r render.Render, repo webhook.WebhookRepository, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	webhook := repo.Fetch(int(id))
	webhook.Enabled = !webhook.Enabled
	repo.Save(webhook)

	r.JSON(200, nil)
}

// Manage webhook events
func webhookReceiveAction(render render.Render, mgr *irc.IrcConnectionManager, request *http.Request) {
	// Parse body
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("error reading the request body. %+v\n", err)
		render.JSON(400, "Invalid data")
	}
	// Process json
	contentType := request.Header.Get("Content-Type")

	var payload map[string]interface{}

	if strings.Contains(contentType, "json") {
		decoder := json.NewDecoder(strings.NewReader(string(body)))
		decoder.UseNumber()

		err := decoder.Decode(&payload)

		if err != nil {
			log.Printf("error parsing JSON payload %+v\n", err)
			render.JSON(400, "Invalid json")
		}
	} else {
		render.JSON(400, "Invalid content-type")
	}
	// All is good
	hook := webhook.WebhookEvent{Body: body}
	err = hook.Process(mgr)
	render.JSON(200, nil)
}
