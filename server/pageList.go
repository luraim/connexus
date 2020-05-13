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

	topicMap := make(map[string]bool)
	// include all md file names, so that we can cover orphaned topics as well
	for _, md := range mdFiles {
		_, f := filepath.Split(md)
		topicMap[f[:len(f)-3]] = true

	}
	// include all outgoing links, so that we can cover unpopulated links
	for _, destinations := range sr.forwardLinks {
		for d := range destinations {
			topicMap[d] = true
		}
	}
	// collect unique topic names, and sort alphabetically
	topicNames := make([]string, 0)
	for t := range topicMap {
		topicNames = append(topicNames, t)
	}
	sort.Strings(topicNames)

	err = sr.pageListTemplate.ExecuteTemplate(w, "pageList.html",
		TopicList{sr.homeTopic, topicNames, sr})
	if err != nil {
		httpErr(w, err)
		return
	}
}

type TopicList struct {
	HomePage string
	Names    []string
	Server   *Server
}
