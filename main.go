package main

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/palantir/stacktrace"

	"gopkg.in/russross/blackfriday.v2"
)

const homeTopic = "test1"
const css = "style.css"
const templateFolder = "templates"

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9_-]+)$")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 2 {
		log.Fatalln("please provide root folder")
	}
	rootFolder := os.Args[1]
	err := exec.Command("cp", css, rootFolder).Run()
	if err != nil {
		log.Println(stacktrace.Propagate(err, "failed to copy %s to %s", css, rootFolder))
	}
	server := newServer(rootFolder)
	server.run()
}

type Server struct {
	rootFolder string
	templates  *template.Template
}

func newServer(rootFolder string) *Server {
	templateFiles := make([]string, 0)
	for _, tf := range []string{"edit.html", "view.html"} {
		templateFiles = append(templateFiles, filepath.Join(templateFolder, tf))
	}
	templates := template.Must(template.ParseFiles(templateFiles...))
	return &Server{
		rootFolder: rootFolder,
		templates:  templates,
	}
}

func (sr *Server) homePage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/"+homeTopic, http.StatusFound)
}

func (sr *Server) run() {
	http.HandleFunc("/", sr.homePage)
	http.HandleFunc("/view/", makeHandler(sr.viewHandler))
	http.HandleFunc("/edit/", makeHandler(sr.editHandler))
	http.HandleFunc("/save/", makeHandler(sr.saveHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

const PRE = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml">
<head>
<title></title>
<meta charset="utf-8" />
<link rel="stylesheet" type="text/css" href="style.css" />
</head>
<body> `

const POST = `
</body>
</html>
`

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
	m := validPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.NotFound(w, r)
		return "", stacktrace.NewError(
			"Invalid Page Title:%s", r.URL.Path)
	}
	return m[2], nil // The title is the second subexpression.
}

type Page struct {
	Title string
	Body  []byte
}

func (pg *Page) Render() template.HTML {
	output := blackfriday.Run(pg.Body)
	return template.HTML(PRE + string(output) + POST)
}

func (sr *Server) pagePath(title string) string {
	return filepath.Join(sr.rootFolder, title+".md")
}

func (sr *Server) savePage(p *Page) error {
	filename := sr.pagePath(p.Title)
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func (sr *Server) loadMarkdown(title string) (*Page, error) {
	filename := sr.pagePath(title)
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"error reading file: '%s'", filename)
	}
	return &Page{Title: title, Body: body}, nil
}

func (sr *Server) viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := sr.loadMarkdown(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	sr.renderTemplate(w, "view", p)
}

func (sr *Server) editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := sr.loadMarkdown(title)
	if err != nil {
		p = &Page{Title: title}
	}
	sr.renderTemplate(w, "edit", p)
}

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := sr.savePage(p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func (sr *Server) renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := sr.templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}
