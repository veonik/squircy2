package manage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/HouzuoGuo/tiedot/db"
	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	"github.com/tyler-sommer/stick"
	"github.com/veonik/squircy2/config"
	"github.com/veonik/squircy2/script"
	"github.com/veonik/squircy2/web"
)

func init() {
	web.Register(NewWithInjector)
}

type module struct {
	conf     *config.Configuration
	manager  *script.ScriptManager
	database *db.DB
}

func NewWithInjector(injector inject.Injector) (web.Module, error) {
	res, err := injector.Invoke(New)
	if err != nil {
		return nil, err
	}
	if m, ok := res[0].Interface().(web.Module); ok {
		return m, nil
	}
	return nil, errors.New("manage: unable to create web module")
}

func New(conf *config.Configuration, manager *script.ScriptManager, database *db.DB) *module {
	return &module{conf, manager, database}
}

func (m *module) Configure(s *web.Server) error {
	s.Group("/manage", func(r martini.Router) {
		r.Get("", m.manageAction)
		r.Post("/update", m.manageUpdateAction)
		r.Post("/export-scripts", m.manageExportScriptsAction)
		r.Post("/import-scripts", m.manageImportScriptsAction)
	})

	return nil
}

func (m *module) manageAction(s *web.StickHandler) {
	s.HTML(200, "manage/edit.html.twig", map[string]stick.Value{
		"config": m.conf,
	})
}

func (m *module) manageUpdateAction(r render.Render, request *http.Request) {
	m.conf.ScriptsAsFiles = request.FormValue("scripts_as_files") == "on"
	m.conf.ScriptsPath = request.FormValue("scripts_path")
	m.conf.EnableFileAPI = request.FormValue("enable_file_api") == "on"
	m.conf.FileAPIRoot = request.FormValue("file_api_root")

	m.conf.TLS = request.FormValue("tls") == "on"
	m.conf.AutoConnect = request.FormValue("auto_connect") == "on"
	m.conf.Network = request.FormValue("network")
	m.conf.Nick = request.FormValue("nick")
	m.conf.Username = request.FormValue("username")

	m.conf.SASL = request.FormValue("enable_sasl") == "on"
	m.conf.SASLUsername = request.FormValue("sasl_username")
	m.conf.SASLPassword = request.FormValue("sasl_password")

	m.conf.OwnerNick = request.FormValue("owner_nick")
	m.conf.OwnerHost = request.FormValue("owner_host")

	m.conf.WebInterface = request.FormValue("web_interface") == "on"
	m.conf.HTTPHostPort = request.FormValue("http_host_port")

	m.conf.RequireHTTPS = request.FormValue("require_https") == "on"
	m.conf.HTTPS = m.conf.RequireHTTPS || request.FormValue("https") == "on"
	m.conf.SSLHostPort = request.FormValue("ssl_host_port")
	m.conf.SSLCertFile = request.FormValue("ssl_cert_file")
	m.conf.SSLCertKey = request.FormValue("ssl_cert_key")

	m.conf.HTTPAuth = request.FormValue("http_auth") == "on"
	m.conf.AuthUsername = request.FormValue("auth_username")
	m.conf.AuthPassword = request.FormValue("auth_password")

	config.SaveConfig(m.database, m.conf)

	r.Redirect("/manage", 302)
}

func (m *module) manageExportScriptsAction(r render.Render, request *http.Request) {
	oldPath := m.conf.ScriptsPath
	defer func() {
		m.conf.ScriptsPath = oldPath
	}()
	m.conf.ScriptsPath = request.FormValue("scripts_export_path")
	err := m.manager.Export()
	if err != nil {
		fmt.Println(err.Error())
	}
	r.Redirect("/manage", 302)
}

func (m *module) manageImportScriptsAction(r render.Render, request *http.Request) {
	oldPath := m.conf.ScriptsPath
	defer func() {
		m.conf.ScriptsPath = oldPath
	}()
	m.conf.ScriptsPath = request.FormValue("scripts_import_path")
	err := m.manager.Import()
	if err != nil {
		fmt.Println(err.Error())
	}
	r.Redirect("/manage", 302)
}
