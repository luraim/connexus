package server

import (
	"bytes"
	"html/template"

	"github.com/palantir/stacktrace"
)

type Page struct {
	MdFileName string
	Body       []byte
	Server     *Server
}

func newPage(mdFileName string, sr *Server) *Page {
	return &Page{
		MdFileName: mdFileName,
		Body:       make([]byte, 0),
		Server:     sr,
	}
}

func (pg *Page) Render() template.HTML {
	var buf bytes.Buffer
	md := *pg.Server.md
	if err := md.Convert(pg.Body, &buf); err != nil {
		return template.HTML(stacktrace.Propagate(err,
			"error converting md %s to html", pg.MdFileName).Error())
	}

	return template.HTML(buf.String())
}
