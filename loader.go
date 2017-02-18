// +build !debug

package squircy2

import (
	"bytes"
	"io"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/staticbin"
	"github.com/tyler-sommer/squircy2/generated"
	"github.com/tyler-sommer/squircy2/generated/manage"
	"github.com/tyler-sommer/squircy2/generated/repl"
	"github.com/tyler-sommer/squircy2/generated/script"
	"github.com/tyler-sommer/squircy2/generated/webhook"
	"github.com/tyler-sommer/stick"
)

var templateMapping = map[string]generatedTemplate{
	"index.html.twig":         generated.TemplateIndexHtmlTwig,
	"webhook/edit.html.twig":  webhook.TemplateWebhookEditHtmlTwig,
	"webhook/index.html.twig": webhook.TemplateWebhookIndexHtmlTwig,
	"webhook/new.html.twig":   webhook.TemplateWebhookNewHtmlTwig,
	"script/edit.html.twig":   script.TemplateScriptEditHtmlTwig,
	"script/index.html.twig":  script.TemplateScriptIndexHtmlTwig,
	"script/new.html.twig":    script.TemplateScriptNewHtmlTwig,
	"repl/index.html.twig":    repl.TemplateReplIndexHtmlTwig,
	"manage/edit.html.twig":   manage.TemplateManageEditHtmlTwig,
}

type stringTemplate struct {
	name     string
	contents string
}

func (t *stringTemplate) Name() string {
	return t.name
}

func (t *stringTemplate) Contents() io.Reader {
	return bytes.NewBufferString(t.contents)
}

func newTemplateLoader() stick.Loader {
	return &assetLoader{}
}

type assetLoader struct{}

func (l *assetLoader) Load(name string) (stick.Template, error) {
	res, err := Asset("views/" + name)
	if err != nil {
		return nil, err
	}
	return &stringTemplate{name, string(res)}, nil
}

func newStaticHandler() martini.Handler {
	return staticbin.Static("public", Asset)
}
