package web

import (
	"fmt"
	"html"
	"io"
	"net/http"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/martini-contrib/render"
	"github.com/martini-contrib/secure"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.com/tyler-sommer/stick"
)

var requestCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: "http_requests_total",
	Help: "Total web request count.",
}, []string{"proto", "method", "url"})

func init() {
	prometheus.MustRegister(requestCounter)
}

type generatedTemplate func(env *stick.Env, output io.Writer, ctx map[string]stick.Value)

type generatedEnv struct {
	*stick.Env
	mapping map[string]generatedTemplate
}

func (env *generatedEnv) Execute(tpl string, out io.Writer, ctx map[string]stick.Value) error {
	if fn, ok := env.mapping[tpl]; ok {
		fn(env.Env, out, ctx)
		return nil
	}
	return env.Env.Execute(tpl, out, ctx)
}

type StickHandler struct {
	env *generatedEnv
	res http.ResponseWriter
}

func (h *StickHandler) HTML(status int, name string, ctx map[string]stick.Value) {
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
	genv := &generatedEnv{env, templateMapping}
	return func(res http.ResponseWriter, req *http.Request, c martini.Context) {
		c.Map(&StickHandler{genv, res})
	}
}

func counterHandler(req *http.Request) {
	proto := "http"
	if req.TLS != nil {
		proto = "https"
	}
	requestCounter.With(prometheus.Labels{"proto": proto, "method": req.Method, "url": req.URL.Path}).Add(1)
}

func (s *Server) Configure() {
	s.Handlers(
		counterHandler,
		newStaticHandler(),
		newStickHandler(),
		render.Renderer(),
		secure.Secure(secure.Options{
			BrowserXssFilter: true,
			FrameDeny:        true,
			SSLRedirect:      s.conf.RequireHTTPS,
			SSLHost:          s.conf.SSLHostPort,
			DisableProdCheck: true,
		}),
	)
	s.NotFound(func(req *http.Request, r render.Render, l log.FieldLogger) {
		r.Error(404)
	})

	if s.conf.HTTPAuth && len(s.conf.AuthUsername) > 0 && len(s.conf.AuthPassword) > 0 {
		s.Use(auth.Basic(s.conf.AuthUsername, s.conf.AuthPassword))
	}

	err := Configure(s)
	if err != nil {
		s.l.Errorln(err)
	}
}
