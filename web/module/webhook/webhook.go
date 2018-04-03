package irc

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/nu7hatch/gouuid"
	log "github.com/sirupsen/logrus"
	"github.com/tyler-sommer/squircy2/event"
	"github.com/tyler-sommer/squircy2/web"
	"github.com/tyler-sommer/squircy2/webhook"
	"github.com/tyler-sommer/stick"
)

func init() {
	web.Register(NewWithInjector)
}

type module struct {
	logger log.FieldLogger
	repo   webhook.WebhookRepository
	events event.EventManager
}

func NewWithInjector(injector inject.Injector) (web.Module, error) {
	res, err := injector.Invoke(New)
	if err != nil {
		return nil, err
	}
	if m, ok := res[0].Interface().(web.Module); ok {
		return m, nil
	}
	return nil, errors.New("webhook: unable to create web module")
}

func New(logger log.FieldLogger, repo webhook.WebhookRepository, events event.EventManager) *module {
	return &module{logger, repo, events}
}

func (m *module) Configure(s *web.Server) error {

	s.Post("/webhooks/:webhook_id", m.webhookReceiveAction)

	s.Group("/webhook", func(r martini.Router) {
		r.Get("", m.webhookAction)
		r.Get("/new", m.newWebhookAction)
		r.Post("/create", m.createWebhookAction)
		r.Get("/:id/edit", m.editWebhookAction)
		r.Post("/:id/update", m.updateWebhookAction)
		r.Post("/:id/remove", m.removeWebhookAction)
		r.Post("/:id/toggle", m.toggleWebhookAction)
	})

	return nil
}

// Manage webhook definitions
func (m *module) webhookAction(s *web.StickHandler) {
	webhooks := m.repo.FetchAll()

	s.HTML(200, "webhook/index.html.twig", map[string]stick.Value{"webhooks": webhooks})
}

func (m *module) newWebhookAction(s *web.StickHandler) {
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

func (m *module) createWebhookAction(r render.Render, request *http.Request) {
	title := request.FormValue("title")

	// Generate key value as an uuid
	key, err := uuid.NewV4()
	if err != nil {
		r.JSON(500, "Error generating key UUID")
	}
	signature := formatSignatureHeader(request.FormValue("signature"))

	hook := &webhook.Webhook{0, title, key.String(), signature, true}
	m.repo.Save(hook)
	log.Debugln("Created webhook %d", hook.ID)
	r.Redirect("/webhook", 302)
}

func (m *module) editWebhookAction(s *web.StickHandler, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	webhook := m.repo.Fetch(int(id))

	s.HTML(200, "webhook/edit.html.twig", map[string]stick.Value{"webhook": webhook})
}

func (m *module) updateWebhookAction(r render.Render, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	title := request.FormValue("title")
	signature := formatSignatureHeader(request.FormValue("signature"))

	hook := m.repo.Fetch(int(id))
	if hook == nil {
		r.Error(404)
		return
	}
	hook.Title = title
	hook.SignatureHeader = signature

	m.repo.Save(hook)

	r.Redirect("/webhook", 302)
}

func (m *module) removeWebhookAction(r render.Render, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	m.repo.Delete(int(id))

	r.JSON(200, nil)
}

func (m *module) toggleWebhookAction(r render.Render, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	webhook := m.repo.Fetch(int(id))
	webhook.Enabled = !webhook.Enabled
	m.repo.Save(webhook)

	r.JSON(200, nil)
}

// Manage webhook events
func (m *module) webhookReceiveAction(r render.Render, request *http.Request, params martini.Params) {
	// Find webhook by it's url
	webhookId, err := strconv.Atoi(params["webhook_id"])
	if err != nil {
		r.JSON(400, "Invalid ID")
		return
	}
	hook := m.repo.Fetch(webhookId)
	if hook == nil {
		r.JSON(404, "Webhook not found")
		return
	}
	// Get signature
	signature := request.Header.Get(hook.SignatureHeader)
	if signature == "" {
		err := "Signature header not found " + hook.SignatureHeader
		r.JSON(400, err)
		return
	}

	// Parse body
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Debugln("error reading the request body. %+v\n", err)
		r.JSON(400, "Invalid data")
		return
	}
	// Process json
	contentType := request.Header.Get("Content-Type")

	// All is good
	evt := webhook.WebhookEvent{Body: body, Webhook: hook, ContentType: contentType, Signature: signature}
	err = evt.Process(m.events)
	if err != nil {
		r.JSON(500, "An error occurred while processing the webhook.")
	}
	r.JSON(200, "OK")
}
