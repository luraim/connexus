package server

import "net/http"

func (sr *Server) editHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	p, err := sr.loadMarkdown(fileName)
	if err != nil {
		p = newPage(fileName, sr)
	}
	sr.renderTemplate(w, "edit", p)
}
