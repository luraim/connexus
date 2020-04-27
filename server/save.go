package server

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/palantir/stacktrace"
)

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	// read textarea contents, making sure to eliminate carriage returns if present
	body := strings.Replace(r.FormValue("body"), "\r", "", -1)
	p := newPage(fileName, sr)
	p.Body = []byte(body)
	err := sr.savePage(p)
	if err != nil {
		httpErr(w, err)
		return
	}
	if sr.linksChanged(p) {
		log.Println("links changed - rebuilding links")
		sr.buildLinks()
	}
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

func (sr *Server) savePage(p *Page) error {
	filename, err := sr.pagePath(p.Topic)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get page path")
	}
	log.Println("writing to file:", filename)
	err = ioutil.WriteFile(filename, p.Body, 0640)
	if err != nil {
		return stacktrace.Propagate(err, "error writing page:%s", p.Topic)
	}

	return nil
}
