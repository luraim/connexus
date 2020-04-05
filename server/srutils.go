package server

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"runtime"

	"github.com/palantir/stacktrace"
)

func httpErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	caller := ""
	if ok && details != nil {
		caller = details.Name() + ":"
	}
	log.Println(caller, err.Error())
}

func (sr *Server) renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	var t *template.Template
	switch tmpl {
	case "view":
		t = sr.viewTemplate
	case "edit":
		t = sr.editTemplate
	}

	if t == nil {
		httpErr(w, stacktrace.NewError("invalid template '%s'", tmpl))
		return
	}

	err := t.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		httpErr(w, err)
		return
	}
}

func (sr *Server) pagePath(fileName string) (string, error) {
	return filepath.Join(sr.rootFolder, fileName+".md"), nil
}
