package server

import (
	"net/http"
	"path/filepath"
	"sort"

	"github.com/palantir/stacktrace"
)

func (sr *Server) pageListHandler(w http.ResponseWriter, r *http.Request) {
	mdFiles, err := filepath.Glob(sr.rootFolder + "/" + "*.md")
	if err != nil {
		httpErr(w, stacktrace.Propagate(err, "error listing md files in %s", sr.rootFolder))
		return
	}

	topicNames := make([]string, 0)
	for _, md := range mdFiles {
		_, f := filepath.Split(md)
		topicNames = append(topicNames, f[:len(f)-3])
	}
	sort.Strings(topicNames)
	err = sr.pageListTemplate.ExecuteTemplate(w, "pageList.html", struct {
		HomePage string
		Names    []string
	}{sr.homeTopic, topicNames})
	if err != nil {
		httpErr(w, err)
		return
	}
}
