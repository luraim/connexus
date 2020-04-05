package server

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/palantir/stacktrace"
)

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	body := r.FormValue("body")
	p := newPage(fileName, sr)
	p.Body = []byte(body)
	err := sr.savePage(p)
	if err != nil {
		httpErr(w, err)
		return
	}
	sr.buildLinks()
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

func (sr *Server) savePage(p *Page) error {
	filename, err := sr.pagePath(p.Topic)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get page path")
	}
	log.Println("writing to file:", filename)
	return ioutil.WriteFile(filename, p.Body, 0640)
}
