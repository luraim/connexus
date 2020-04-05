package server

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/palantir/stacktrace"
)

var linkRe = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

type LinkMap map[string]map[string]bool

func (lm LinkMap) String() string {
	b := &strings.Builder{}
	for _, topic := range lm.topics() {
		fmt.Fprintf(b, "%s:\n", topic)
		for _, link := range lm.links(topic) {
			fmt.Fprintf(b, "\t%s\n", link)
		}
	}
	return b.String()
}

func newLinkMap() LinkMap {
	return make(LinkMap)
}

func (lm LinkMap) add(from, to string) {
	links, ok := lm[from]
	if !ok {
		links = make(map[string]bool)
	}
	links[to] = true
	lm[from] = links
}

func (lm LinkMap) remove(from, to string) {
	links, ok := lm[from]
	if !ok {
		return
	}
	delete(links, to)
	lm[from] = links
}

func (lm LinkMap) topics() []string {
	ret := make([]string, 0)
	for topic := range lm {
		ret = append(ret, topic)
	}
	sort.Strings(ret)
	return ret
}

func (lm LinkMap) links(topic string) []string {
	ret := make([]string, 0)
	for link := range lm[topic] {
		ret = append(ret, link)
	}
	sort.Strings(ret)
	return ret
}

func (sr *Server) buildLinks() error {
	forwardLinks := newLinkMap()
	reverseLinks := newLinkMap()
	fis, err := ioutil.ReadDir(sr.rootFolder)
	if err != nil {
		return stacktrace.Propagate(err,
			"error reading contents of root folder:%s", sr.rootFolder)
	}

	// build forward links going through each file
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		fname := fi.Name()
		ext := filepath.Ext(fname)
		if ext == ".md" {
			topicName := strings.TrimSuffix(fname, ext)
			content, err := ioutil.ReadFile(filepath.Join(sr.rootFolder, fname))
			if err != nil {
				log.Println(stacktrace.Propagate(err, "%s", fname))
				continue
			}
			links := parseLinks(string(content))
			for _, link := range links {
				forwardLinks.add(topicName, link)
			}
		}
	}

	// build reverse links
	for _, topicName := range forwardLinks.topics() {
		for _, link := range forwardLinks.links(topicName) {
			reverseLinks.add(link, topicName)
		}
	}

	sr.forwardLinks = forwardLinks
	sr.reverseLinks = reverseLinks
	log.Printf("links rebuilt: %d forward, %d reverse\n",
		len(forwardLinks), len(reverseLinks))
	return nil
}

// parseLinks parses markdown content to extract all links
func parseLinks(content string) []string {
	ret := make([]string, 0)
	res := linkRe.FindAllStringSubmatch(content, -1)
	for _, m := range res {
		if len(m) != 3 {
			continue
		}
		_, fileName := m[1], m[2]
		if strings.HasPrefix(fileName, "/static") {
			// link to static content - skip
			continue
		}
		ret = append(ret, fileName)
	}
	return ret
}

func (sr *Server) outgoingLinks(topic string) []string {
	return sr.forwardLinks.links(topic)
}

func (sr *Server) incomingLinks(topic string) []string {
	return sr.reverseLinks.links(topic)
}
