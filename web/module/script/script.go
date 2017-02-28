package script

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/squircy2/script"
	"github.com/tyler-sommer/squircy2/web"
	"github.com/tyler-sommer/stick"
)

func init() {
	web.Register(NewWithInjector)
}

type module struct {
	repo    script.ScriptRepository
	manager *script.ScriptManager
}

func NewWithInjector(injector inject.Injector) (web.Module, error) {
	res, err := injector.Invoke(New)
	if err != nil {
		return nil, err
	}
	if m, ok := res[0].Interface().(web.Module); ok {
		return m, nil
	}
	return nil, errors.New("script: unable to create web module")
}

func New(repo script.ScriptRepository, manager *script.ScriptManager) *module {
	return &module{repo, manager}
}

func (m *module) Configure(s *web.Server) error {
	s.Group("/script", func(r martini.Router) {
		r.Get("", m.scriptAction)
		r.Post("/reinit", m.scriptReinitAction)
		r.Get("/new", m.newScriptAction)
		r.Post("/create", m.createScriptAction)
		r.Get("/:id/edit", m.editScriptAction)
		r.Post("/:id/update", m.updateScriptAction)
		r.Post("/:id/remove", m.removeScriptAction)
		r.Post("/:id/toggle", m.toggleScriptAction)
	})
	s.Group("/repl", func(r martini.Router) {
		r.Get("", m.replAction)
		r.Post("/execute", m.replExecuteAction)
	})

	return nil
}

func (m *module) scriptAction(s *web.StickHandler) {
	s.HTML(200, "script/index.html.twig", map[string]stick.Value{"scripts": m.repo.FetchAll()})
}

func (m *module) scriptReinitAction(r render.Render) {
	m.manager.ReInit()

	r.JSON(200, nil)
}

func (m *module) newScriptAction(s *web.StickHandler) {
	s.HTML(200, "script/new.html.twig", nil)
}

func (m *module) createScriptAction(r render.Render, request *http.Request) {
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	m.repo.Save(&script.Script{0, script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func (m *module) editScriptAction(s *web.StickHandler, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := m.repo.Fetch(int(id))

	s.HTML(200, "script/edit.html.twig", map[string]stick.Value{"script": script})
}

func (m *module) updateScriptAction(r render.Render, params martini.Params, request *http.Request) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)
	sType := request.FormValue("type")
	title := request.FormValue("title")
	body := request.FormValue("body")

	m.repo.Save(&script.Script{int(id), script.ScriptType(sType), title, body, true})

	r.Redirect("/script", 302)
}

func (m *module) removeScriptAction(r render.Render, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	m.repo.Delete(int(id))

	r.JSON(200, nil)
}

func (m *module) toggleScriptAction(r render.Render, params martini.Params) {
	id, _ := strconv.ParseInt(params["id"], 0, 64)

	script := m.repo.Fetch(int(id))
	script.Enabled = !script.Enabled
	m.repo.Save(script)

	r.JSON(200, nil)
}

func (m *module) replAction(s *web.StickHandler) {
	s.HTML(200, "repl/index.html.twig", nil)
}

func (m *module) replExecuteAction(r render.Render, request *http.Request) {
	code := request.FormValue("script")
	sType := script.ScriptType(request.FormValue("scriptType"))

	res, err := m.manager.RunUnsafe(sType, code)
	var errStr interface{}
	if err != nil {
		errStr = err.Error()
	}
	r.JSON(200, map[string]interface{}{
		"res": res,
		"err": errStr,
	})
}
