package server

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gobuffalo/packr/v2"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

const templateFolder = "templates"

type Server struct {
	rootFolder   string
	homeTopic    string
	port         string
	viewTemplate *template.Template
	editTemplate *template.Template
	md           *goldmark.Markdown
	box          *packr.Box
	forwardLinks LinkMap
	reverseLinks LinkMap
}

var bootstrapFiles = []string{"bootstrap.min.css", "jquery.min.js",
	"bootstrap.bundle.min.js"}

func NewServer(rootFolder, homeTopic, port string) *Server {

	box := packr.New("connexus_box", "./templates")

	// ensure that the wiki folder has 'img' and 'css' sub folders
	subFolders := []string{"img", "css"}
	for _, f := range subFolders {
		subFolderPath := filepath.Join(rootFolder, f)
		if _, err := os.Stat(subFolderPath); os.IsNotExist(err) {
			log.Println("creating sub folder:", subFolderPath)
			err := os.Mkdir(subFolderPath, 0755)
			if err != nil {
				log.Fatalln("failed to create", subFolderPath)
			}
		} else {
			log.Printf("%s already present\n", subFolderPath)
		}
	}

	// ensure that 'css' subfolder in wiki contains bootstrap files
	for _, bf := range bootstrapFiles {
		bootstrapFile := filepath.Join(rootFolder, "css", bf)
		if _, err := os.Stat(bootstrapFile); os.IsNotExist(err) {
			log.Printf("bootstrap file %s not present. writing...\n", bootstrapFile)
			// read css content from box
			css, err := box.FindString("css/" + bf)
			if err != nil {
				log.Fatalf("could not find bootstrap file %s in box\n", bf)
			}
			// write file in css sub folder under wiki root
			err = ioutil.WriteFile(bootstrapFile, []byte(css), 0644)
			if err != nil {
				log.Fatalln("failed to write css file", bootstrapFile)
			}
		} else {
			log.Printf("bootstrap file %s already present\n", bootstrapFile)
		}
	}
	customCss := "style.css"
	// ensure that the css folder contains the latest copy of the custom css file
	cssFile := filepath.Join(rootFolder, "css", customCss)
	// read css content from box
	css, err := box.FindString("css/" + customCss)
	if err != nil {
		log.Fatalf("could not find custom css file %s in box\n", cssFile)
	}
	// write file in css sub folder under wiki root
	err = ioutil.WriteFile(cssFile, []byte(css), 0644)
	if err != nil {
		log.Fatalln("failed to write css file", cssFile)
	}
	log.Println("wrote custom css file", cssFile)

	// initialize markdown parser
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.TaskList,
			extension.Linkify, extension.DefinitionList, extension.Footnote,
			extension.Strikethrough, extension.Typographer,
			highlighting.Highlighting),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()))

	loadTemplate := func(tf string) *template.Template {
		content, err := box.FindString(tf)
		if err != nil {
			log.Fatalln("unable to find template", tf)
		}
		tmpl, err := template.New(tf).Parse(content)
		if err != nil {
			log.Fatalln("failed to parse template", tf)
		}
		return tmpl
	}

	viewTemplate := loadTemplate("view.html")
	editTemplate := loadTemplate("edit.html")

	sr := &Server{
		rootFolder:   rootFolder,
		homeTopic:    homeTopic,
		port:         port,
		viewTemplate: viewTemplate,
		editTemplate: editTemplate,
		md:           &md,
		box:          box,
	}

	err = sr.buildLinks()
	if err != nil {
		log.Fatalln(err)
	}

	return sr
}

func (sr *Server) Run() {

	log.Printf("Starting server with root: '%s' and home: '%s' on port:%s\n",
		sr.rootFolder, sr.homeTopic, sr.port)

	http.HandleFunc("/", sr.homePage)
	http.HandleFunc("/view/", makeHandler(sr.viewHandler))
	http.HandleFunc("/edit/", makeHandler(sr.editHandler))
	http.HandleFunc("/save/", makeHandler(sr.saveHandler))

	fs := http.FileServer(http.Dir(sr.rootFolder))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("localhost:%s", sr.port),
		nil))
}

var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9/_-]+)$")

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			log.Println("url path not found:", r.URL.Path)
			http.NotFound(w, r)
			return
		}
		//log.Println("md file name:", m[2])
		fn(w, r, m[2])
	}
}
