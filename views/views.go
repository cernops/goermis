package views

import (
	"errors"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"
)

var (
	baseDir    = filepath.Join(*bootstrap.HomeFlag, "/templates/base.html")
	layoutsDir = filepath.Join(*bootstrap.HomeFlag, "/templates/layouts")
	formsDir   = filepath.Join(*bootstrap.HomeFlag, "/templates/forms")
)

// TemplateRegistry defines the template registry struct
type TemplateRegistry struct {
	templates map[string]*template.Template
}

// Render Implements e.Renderer interface
func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		err := errors.New("Template not found -> " + name)
		log.Warn("Template not found")
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

func readCurrentDir(dir string) []string {
	file, err := os.Open(dir)
	if err != nil {
		log.Fatal("Failed opening directory")

	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	return list
}

//InitViews initializes the GUI
func InitViews(e *echo.Echo) {
	staticfiles := filepath.Join(*bootstrap.HomeFlag, "/staticfiles")
	e.Static("/staticfiles", staticfiles)
	templates := make(map[string]*template.Template)
	for _, file := range readCurrentDir(layoutsDir) {
		templates[file] = template.Must(template.ParseFiles(baseDir, layoutsDir+"/"+file))
		templates[file].ParseGlob(formsDir + "/" + "*.html")

	}

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

}
