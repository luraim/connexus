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
	"runtime"
	"strings"

	mathjax "github.com/litao91/goldmark-mathjax"

	highlighting "github.com/yuin/goldmark-highlighting"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"github.com/palantir/stacktrace"
)

const templateFolder = "templates"

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9/_-]+)$")
var linkRe = regexp.MustCompile(`\[(.*?)\]\((.*?)\)`)

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

const pathSep = string(os.PathSeparator)

func (sr *Server) pagePath(fileName string) (string, error) {

	parts := strings.Split(fileName, pathSep)
	if len(parts) > 1 {
		path := strings.Join(parts[:len(parts)-1], pathSep)
		path = filepath.Join(sr.rootFolder, path)
		log.Println("creating directory:", path)
		err := os.MkdirAll(path, 0700)
		if err != nil {
			return "", stacktrace.Propagate(err,
				"failed to create path:%s", path)
		}
	}

	return filepath.Join(sr.rootFolder, fileName+".md"), nil
}

func (sr *Server) savePage(p *Page) error {
	filename, err := sr.pagePath(p.MdFileName)
	if err != nil {
		return stacktrace.Propagate(err, "failed to get page path")
	}
	log.Println("writing to file:", filename)
	return ioutil.WriteFile(filename, p.Body, 0640)
}

func (sr *Server) loadMarkdown(title string) (*Page, error) {
	filename, err := sr.pagePath(title)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get page path")
	}
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, stacktrace.Propagate(err,
			"error reading file: '%s'", filename)
	}

	page := newPage(title, sr)
	page.Body = body
	return page, nil
}

type Link struct {
	title    string
	fileName string
}

// getLinks parses markdown content to extract all links
func getLinks(content string) []*Link {
	ret := make([]*Link, 0)
	res := linkRe.FindAllStringSubmatch(content, -1)
	for _, m := range res {
		if len(m) != 3 {
			continue
		}
		title, fileName := m[1], m[2]
		ret = append(ret, &Link{title: title, fileName: fileName})
	}
	return ret
}

func (sr *Server) viewHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	p, err := sr.loadMarkdown(fileName)
	if err != nil {
		http.Redirect(w, r, "/edit/"+fileName, http.StatusFound)
		return
	}
	sr.renderTemplate(w, "view", p)
}

func (sr *Server) editHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	p, err := sr.loadMarkdown(fileName)
	if err != nil {
		p = newPage(fileName, sr)
	}
	sr.renderTemplate(w, "edit", p)
}

func (sr *Server) saveHandler(w http.ResponseWriter, r *http.Request, fileName string) {
	body := r.FormValue("body")
	p := newPage(fileName, sr)
	p.Body = []byte(body)
	err := sr.savePage(p)
	if err != nil {
		httpErr(w, err)
		return
	}
	http.Redirect(w, r, "/view/"+fileName, http.StatusFound)
}

func (sr *Server) renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := sr.templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		httpErr(w, err)
		return
	}
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Println("url path not found:", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		log.Println("md file name:", m[2])
		fn(w, r, m[2])
	}
}

type PageLinks struct {
	Incoming []*Link
	Outgoing []*Link
}

func newPageLinks() *PageLinks {
	return &PageLinks{
		Incoming: make([]*Link, 0),
		Outgoing: make([]*Link, 0),
	}
}

type PageLinksMap map[string]*PageLinks

func (sr *Server) buildLinks() {

}

func httpErr(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	caller := ""
	if ok && details != nil {
		caller = details.Name() + ":"
	}
	log.Println(caller, err.Error())
}
