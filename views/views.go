package views

import (
	"errors"
	"html/template"
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gitlab.cern.ch/lb-experts/goermis/bootstrap"

	"github.com/labstack/echo/v4"
)

var (
	baseDir    = filepath.Join(bootstrap.HomeFlag, "/templates/base.html")
	layoutsDir = filepath.Join(bootstrap.HomeFlag, "/templates/layouts")
	formsDir   = filepath.Join(bootstrap.HomeFlag, "/templates/forms")
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
		return err
	}
	return tmpl.ExecuteTemplate(w, "base.html", data)
}

func readCurrentDir(dir string) []string {
	file, err := os.Open(dir)
	if err != nil {
		log.Fatalf("failed opening directory: %s", err)
	}
	defer file.Close()

	list, _ := file.Readdirnames(0) // 0 to read all files and folders
	return list
}

//InitViews initializes the GUI
func InitViews(e *echo.Echo) {
	staticfiles := filepath.Join(bootstrap.HomeFlag, "/staticfiles")
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
