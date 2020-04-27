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
var todoRe = regexp.MustCompile(`(?i)TODO:(.*)`)

// source topic -> destination topic set
// inner map is used as a set
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
	todoLinks := make(map[string]string)

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
			links, todos := parseLinks(string(content))
			for _, link := range links {
				forwardLinks.add(topicName, link)
			}
			for _, todo := range todos {
				todoLinks[todo] = topicName
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
	sr.todoLinks = todoLinks
	log.Printf("links rebuilt: %d forward, %d reverse, %d todos\n",
		len(forwardLinks), len(reverseLinks), len(todoLinks))

	todosPage := newPage("todos", sr)
	todosBody := &strings.Builder{}
	for todo, topic := range sr.todoLinks {
		fmt.Fprintf(todosBody, "- [%s](%s) : %s\n", topic, topic, todo)
	}
	todosPage.Body = []byte(todosBody.String())
	sr.savePage(todosPage)

	return nil
}

// parseLinks parses markdown content to extract all links
func parseLinks(content string) ([]string, []string) {
	links := make([]string, 0)
	res := linkRe.FindAllStringSubmatch(content, -1)
	for _, m := range res {
		if len(m) != 3 {
			continue
		}
		_, fileName := m[1], m[2]
		if strings.HasPrefix(fileName, "/static") ||
			strings.HasPrefix(fileName, "http://") ||
			strings.HasPrefix(fileName, "https://") {
			// link to static content - skip
			continue
		}
		links = append(links, fileName)
	}
	sort.Strings(links)

	todos := make([]string, 0)
	res = todoRe.FindAllStringSubmatch(content, -1)
	for _, m := range res {
		if len(m) != 2 {
			continue
		}
		todos = append(todos, m[1])
	}

	return links, todos
}

func (sr *Server) linksChanged(page *Page) bool {
	links, todos := parseLinks(string(page.Body))
	existingLinks := sr.outgoingLinks(page.Topic)
	//fmt.Println(links, existingLinks)
	linksChanged := !compareStringSlices(links, existingLinks)

	todosChanged := false
	for _, todo := range todos {
		if topicName, ok := sr.todoLinks[todo]; !ok {
			if topicName == page.Topic {
				// found a new todo
				todosChanged = true
				break
			}
		}
	}
	return linksChanged || todosChanged
}

func compareStringSlices(a, b []string) bool {

	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func (sr *Server) outgoingLinks(topic string) []string {
	return sr.forwardLinks.links(topic)
}

func (sr *Server) incomingLinks(topic string) []string {
	return sr.reverseLinks.links(topic)
}
