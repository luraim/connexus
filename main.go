package main

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	mathjax "github.com/litao91/goldmark-mathjax"

	highlighting "github.com/yuin/goldmark-highlighting"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/palantir/stacktrace"
)

const templateFolder = "templates"

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9_-]+)$")

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 3 {
		log.Fatalln("please provide root folder and main topic")
	}
	rootFolder, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln("error getting absolute path to root folder")
	}

	homeTopic := os.Args[2]

	imgFolder := filepath.Join(rootFolder, "img")
	exec.Command("mkdir", imgFolder).Run()

	server := newServer(rootFolder, homeTopic)
	log.Printf("Starting server with root: '%s' and home: '%s'\n",
		server.rootFolder, server.homeTopic)
	server.run()
}

type Server struct {
	rootFolder string
	homeTopic  string
	templates  *template.Template
	md         *goldmark.Markdown
}

func newServer(rootFolder, homeTopic string) *Server {
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.TaskList,
			extension.Linkify, extension.DefinitionList, extension.Footnote,
			extension.Strikethrough, extension.Typographer,
			highlighting.Highlighting, mathjax.MathJax),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	templateFiles := make([]string, 0)
	for _, tf := range []string{"edit.html", "view.html"} {
		templateFiles = append(templateFiles, filepath.Join(templateFolder, tf))
	}
	templates := template.Must(template.ParseFiles(templateFiles...))
	return &Server{
		rootFolder: rootFolder,
		homeTopic:  homeTopic,
		templates:  templates,
		md:         &md,
	}
}

func (sr *Server) homePage(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/view/"+sr.homeTopic, http.StatusFound)
}

func (sr *Server) run() {
	http.HandleFunc("/", sr.homePage)
	http.HandleFunc("/view/", makeHandler(sr.viewHandler))
	http.HandleFunc("/edit/", makeHandler(sr.editHandler))
	http.HandleFunc("/save/", makeHandler(sr.saveHandler))
	fs := http.FileServer(http.Dir(filepath.Join(sr.rootFolder, "img")))
	http.Handle("/img/", http.StripPrefix("/img/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

type Page struct {
	Title  string
	Body   []byte
	Server *Server
}

func newPage(title string, sr *Server) *Page {
	return &Page{
		Title:  title,
		Body:   make([]byte, 0),
		Server: sr,
	}
}

func (pg *Page) Render() template.HTML {
	var buf bytes.Buffer
	md := *pg.Server.md
	if err := md.Convert(pg.Body, &buf); err != nil {
		return template.HTML(stacktrace.Propagate(err,
			"error converting md %s to html", pg.Title).Error())
	}

	return template.HTML(buf.String())
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
	page := newPage(title, sr)
	page.Body = body
	return page, nil
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
		p = newPage(title, sr)
	}
	sr.renderTemplate(w, "edit", p)
}

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := newPage(title, sr)
	p.Body = []byte(body)
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
