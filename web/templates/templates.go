package templates

import (
	"embed"
	"html/template"
	"io"

	"github.com/gin-gonic/gin"
)

//go:embed *.html
var tmplFS embed.FS

type Template struct {
	templates *template.Template
}

func New() *Template {
	tmpl := template.Must(template.New("").ParseFS(tmplFS, "*.html"))

	return &Template{templates: tmpl}
}

func (t *Template) Render(w io.Writer, name string, data any) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(tmplFS, name))

	return tmpl.ExecuteTemplate(w, name, data)
}

func (t *Template) AutoRender(ctx *gin.Context, route string, extraData gin.H) {
	data := gin.H{}

	if userData, exists := ctx.Get("userData"); exists {
		data["User"] = userData
	}

	for k, v := range extraData {
		data[k] = v
	}

	t.Render(ctx.Writer, route, data)
}
