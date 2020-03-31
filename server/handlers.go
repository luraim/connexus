package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"

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

func (sr *Server) editHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	p, err := sr.loadMarkdown(fileName)
	if err != nil {
		p = newPage(fileName, sr)
	}
	sr.renderTemplate(w, "edit", p)
}

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	body := r.FormValue("body")
	p := newPage(fileName, sr)
	p.Body = []byte(body)
	err := sr.savePage(p)
	if err != nil {
		httpErr(w, err)
		return
	}
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9/_-]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Println("url path not found:", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		log.Println("md file name:", m[2])
		fn(w, r, m[2])
	}
}

func (sr *Server) pagePath(fileName string) (string, error) {
	return filepath.Join(sr.rootFolder, fileName+".md"), nil
}

func (sr *Server) savePage(p *Page) error {
	filename, err := sr.pagePath(p.MdFileName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get page path")
	}
	log.Println("writing to file:", filename)
	return ioutil.WriteFile(filename, p.Body, 0640)
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
