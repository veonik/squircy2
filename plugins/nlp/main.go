package main // import "github.com/veonik/squircy2/plugins/nlp"

import (
	"errors"

	"github.com/codegangsta/inject"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/render"
	log "github.com/sirupsen/logrus"
	_ "github.com/tyler-sommer/stick/parse"
	"github.com/veonik/squircy2/script"
	"github.com/veonik/squircy2/web"
	"gopkg.in/jdkato/prose.v2"
)

func Initialize(i inject.Injector) error {
	if err := web.Register(NewWithInjector); err != nil {
		return err
	}
	if _, err := i.Invoke(BindOnInitFunc); err != nil {
		return err
	}
	return nil
}

func BindOnInitFunc(m *script.ScriptManager, logger *log.Logger) {
	bn := &initBinding{logger: logger}
	m.AddOnInit(bn.AddFunctionsToScriptManager)
}

type initBinding struct {
	logger *log.Logger
}

func (b *initBinding) AddFunctionsToScriptManager(m *script.ScriptManager) {
	must := func(err error) {
		if err != nil {
			b.logger.Warnln("Error binding value", err)
		}
	}

	vm, err := m.GetVM(script.Javascript)
	if err != nil {
		must(err)
		return
	}
	if v, err := vm.Object("({})"); err == nil {
		must(v.Set("NewDocument", func(input string) *prose.Document {
			doc, err := prose.NewDocument(input)
			if err != nil {
				panic(err)
			}
			return doc
		}))
		must(vm.Set("NLP", v))
	} else {
		must(err)
	}
}

type webModule struct{}

func NewWithInjector(injector inject.Injector) (web.Module, error) {
	res, err := injector.Invoke(New)
	if err != nil {
		return nil, err
	}
	if m, ok := res[0].Interface().(web.Module); ok {
		return m, nil
	}
	return nil, errors.New("webhook: unable to create web webModule")
}

func New() *webModule {
	return &webModule{}
}

func (m *webModule) Configure(s *web.Server) error {
	s.Group("/nlp", func(r martini.Router) {
		r.Get("", m.indexAction)
	})

	return nil
}

func (m *webModule) indexAction(r render.Render) {
	r.JSON(200, map[string]interface{}{"test": "mate"})
}
