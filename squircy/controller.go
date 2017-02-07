package squircy

import (
	"fmt"
	"html"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/secure"
	"github.com/nu7hatch/gouuid"
	"github.com/tyler-sommer/squircy2/squircy/config"
	"github.com/tyler-sommer/squircy2/squircy/event"
	"github.com/tyler-sommer/squircy2/squircy/eventsource"
	"github.com/tyler-sommer/squircy2/squircy/irc"
	"github.com/tyler-sommer/squircy2/squircy/script"
	"github.com/tyler-sommer/squircy2/squircy/webhook"
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

func configureWeb(manager *Manager, conf *config.Configuration) {
	manager.Handlers(
		newStaticHandler(),
		newStickHandler(),
		render.Renderer(),
		secure.Secure(secure.Options{
			BrowserXssFilter: true,
			FrameDeny:        true,
			SSLRedirect:      conf.RequireHTTPS,
			SSLHost:          conf.SSLHostPort,
			DisableProdCheck: true,
		}),
	)
	manager.NotFound(func(req *http.Request, r render.Render, l *log.Logger) {
		r.Error(404)
	})

	manager.Post("/webhooks/:webhook_id", webhookReceiveAction)

	// Admin web interface
	handlers := []martini.Handler{}
	if conf.HTTPAuth && len(conf.AuthUsername) > 0 && len(conf.AuthPassword) > 0 {
		handlers = append(handlers, auth.Basic(conf.AuthUsername, conf.AuthPassword))
	}
	manager.Group("", func(rm martini.Router) {
		rm.Get("/event", func(es *eventsource.Broker, w http.ResponseWriter, r *http.Request) {
			es.ServeHTTP(w, r)
		})
		rm.Get("/", indexAction)
		rm.Get("/status", statusAction)
		rm.Group("/manage", func(r martini.Router) {
			r.Get("", manageAction)
			r.Post("/update", manageUpdateAction)
			r.Post("/export-scripts", manageExportScriptsAction)
			r.Post("/import-scripts", manageImportScriptsAction)
		})
		rm.Post("/connect", connectAction)
		rm.Post("/disconnect", disconnectAction)
		rm.Group("/script", func(r martini.Router) {
			r.Get("", scriptAction)
			r.Post("/reinit", scriptReinitAction)
			r.Get("/new", newScriptAction)
			r.Post("/create", createScriptAction)
			r.Get("/:id/edit", editScriptAction)
			r.Post("/:id/update", updateScriptAction)
			r.Post("/:id/remove", removeScriptAction)
			r.Post("/:id/toggle", toggleScriptAction)
		})
		rm.Group("/repl", func(r martini.Router) {
			r.Get("", replAction)
			r.Post("/execute", replExecuteAction)
		})
		rm.Group("/webhook", func(r martini.Router) {
			r.Get("", webhookAction)
			r.Get("/new", newWebhookAction)
			r.Post("/create", createWebhookAction)
			r.Get("/:id/edit", editWebhookAction)
			r.Post("/:id/update", updateWebhookAction)
			r.Post("/:id/remove", removeWebhookAction)
			r.Post("/:id/toggle", toggleWebhookAction)
		})
	}, handlers...)
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
	conf.ScriptsAsFiles = request.FormValue("scripts_as_files") == "on"
	conf.ScriptsPath = request.FormValue("scripts_path")

	conf.TLS = request.FormValue("tls") == "on"
	conf.AutoConnect = request.FormValue("auto_connect") == "on"
	conf.Network = request.FormValue("network")
	conf.Nick = request.FormValue("nick")
	conf.Username = request.FormValue("username")
	conf.OwnerNick = request.FormValue("owner_nick")
	conf.OwnerHost = request.FormValue("owner_host")

	conf.WebInterface = request.FormValue("web_interface") == "on"
	conf.HTTPHostPort = request.FormValue("http_host_port")

	conf.RequireHTTPS = request.FormValue("require_https") == "on"
	conf.HTTPS = conf.RequireHTTPS || request.FormValue("https") == "on"
	conf.SSLHostPort = request.FormValue("ssl_host_port")
	conf.SSLCertFile = request.FormValue("ssl_cert_file")
	conf.SSLCertKey = request.FormValue("ssl_cert_key")

	conf.HTTPAuth = request.FormValue("http_auth") == "on"
	conf.AuthUsername = request.FormValue("auth_username")
	conf.AuthPassword = request.FormValue("auth_password")

	config.SaveConfig(database, conf)

	r.Redirect("/manage", 302)
}

func manageExportScriptsAction(r render.Render, m *script.ScriptManager, conf *config.Configuration, request *http.Request) {
	oldPath := conf.ScriptsPath
	defer func() {
		conf.ScriptsPath = oldPath
	}()
	conf.ScriptsPath = request.FormValue("scripts_export_path")
	err := m.Export()
	if err != nil {
		fmt.Println(err.Error())
	}
	r.Redirect("/manage", 302)
}

func manageImportScriptsAction(r render.Render, m *script.ScriptManager, conf *config.Configuration, request *http.Request) {
	oldPath := conf.ScriptsPath
	defer func() {
		conf.ScriptsPath = oldPath
	}()
	conf.ScriptsPath = request.FormValue("scripts_import_path")
	err := m.Import()
	if err != nil {
		fmt.Println(err.Error())
	}
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

func formatSignatureHeader(header string) string {
	// Format header in Camel case
	parts := strings.Split(header, "-")
	for i := 0; i < len(parts); i++ {
		if len(parts[i]) > 1 {
			first := strings.ToUpper(parts[i][0:1])
			last := strings.ToLower(parts[i][1:len(parts[i])])
			parts[i] = first + last
		}
	}
	res := strings.Join(parts, "-")
	return res
}

func createWebhookAction(r render.Render, repo webhook.WebhookRepository, request *http.Request) {
	title := request.FormValue("title")

	// Generate key value as an uuid
	key, err := uuid.NewV4()
	if err != nil {
		r.JSON(500, "Error generating key UUID")
	}
	signature := formatSignatureHeader(request.FormValue("signature"))

	hook := &webhook.Webhook{0, title, key.String(), signature, true}
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
	title := request.FormValue("title")
	signature := formatSignatureHeader(request.FormValue("signature"))

	hook := repo.Fetch(int(id))
	if hook == nil {
		r.Error(404)
		return
	}
	hook.Title = title
	hook.SignatureHeader = signature

	repo.Save(hook)

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
func webhookReceiveAction(render render.Render, evm event.EventManager, repo webhook.WebhookRepository, request *http.Request, params martini.Params) {
	// Find webhook by it's url
	webhookId, err := strconv.Atoi(params["webhook_id"])
	if err != nil {
		render.JSON(400, "Invalid ID")
		return
	}
	hook := repo.Fetch(webhookId)
	if hook == nil {
		render.JSON(404, "Webhook not found")
		return
	}
	// Get signature
	signature := request.Header.Get(hook.SignatureHeader)
	if signature == "" {
		err := "Signature header not found " + hook.SignatureHeader
		render.JSON(400, err)
		return
	}

	// Parse body
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Printf("error reading the request body. %+v\n", err)
		render.JSON(400, "Invalid data")
		return
	}
	// Process json
	contentType := request.Header.Get("Content-Type")

	// All is good
	evt := webhook.WebhookEvent{Body: body, Webhook: hook, ContentType: contentType, Signature: signature}
	err = evt.Process(evm)
	if err != nil {
		render.JSON(500, "An error occurred while processing the webhook.")
	}
	render.JSON(200, "OK")
}
