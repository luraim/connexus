package server

import (
	"bytes"
	"html/template"

	"github.com/palantir/stacktrace"
)

type Page struct {
	Topic  string
	Body   []byte
	Server *Server
}

func newPage(topicName string, sr *Server) *Page {
	return &Page{
		Topic:  topicName,
		Body:   make([]byte, 0),
		Server: sr,
	}
}

func (pg *Page) Render() template.HTML {
	var buf bytes.Buffer
	md := *pg.Server.md
	if err := md.Convert(pg.Body, &buf); err != nil {
		return template.HTML(stacktrace.Propagate(err,
			"error converting md %s to html", pg.Topic).Error())
	}

	return template.HTML(buf.String())
}

func (pg *Page) OutgoingLinks() []string {
	return pg.Server.outgoingLinks(pg.Topic)
}

func (pg *Page) PageExists(topic string) bool {
	return pg.Server.PageExists(topic)
}

func (pg *Page) IncomingLinks() []string {
	return pg.Server.incomingLinks(pg.Topic)
}

func (pg *Page) HomePage() string {
	return pg.Server.homeTopic
}
