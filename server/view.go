package server

import (
	"io/ioutil"
	"net/http"

	"github.com/palantir/stacktrace"
)

func (sr *Server) homePage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/"+sr.homeTopic, http.StatusFound)
}

func (sr *Server) viewHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	p, err := sr.loadMarkdown(fileName)
	if err != nil {
		http.Redirect(w, r, "/edit/"+fileName, http.StatusFound)
		return
	}
	sr.renderTemplate(w, "view", p)
}

func (sr *Server) loadMarkdown(title string) (*Page, error) {
	filename, err := sr.pagePath(title)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get page path")
	}
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"error reading file: '%s'", filename)
	}

	page := newPage(title, sr)
	page.Body = body
	return page, nil
}
