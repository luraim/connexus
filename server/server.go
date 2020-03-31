package server

import (
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	viewTemplate *template.Template
	editTemplate *template.Template
	md           *goldmark.Markdown
	box          *packr.Box
}

const cssFileName = "style.css"

func NewServer(rootFolder, homeTopic string) *Server {

	box := packr.New("connexus_box", "./templates")

	// ensure that the wiki folder has 'img' and 'css' sub folders
	subFolders := []string{"img", "css"}
	for _, f := range subFolders {
		subFolderPath := filepath.Join(rootFolder, f)
		if _, err := os.Stat(subFolderPath); os.IsNotExist(err) {
			log.Println("creating sub folder:", subFolderPath)
			os.Mkdir(subFolderPath, 0755)
		}
	}

	// ensure that 'css' subfolder in wiki contains css file
	cssFile := filepath.Join(rootFolder, "css", cssFileName)
	if _, err := os.Stat(cssFile); os.IsNotExist(err) {
		log.Printf("css file %s not present. writing...\n", cssFile)
		// read css content from box
		css, err := box.FindString(cssFileName)
		if err != nil {
			log.Fatalln("could not find css file in box")
		}
		// write css file in css sub folder under wiki root
		err = ioutil.WriteFile(cssFile, []byte(css), 0644)
		if err != nil {
			log.Fatalln("failed to write css file", cssFile)
		}
	} else {
		log.Printf("css file %s present\n", cssFile)
	}

	// initialize markdown parser
	md := goldmark.New(
		goldmark.WithExtensions(extension.Table, extension.TaskList,
			extension.Linkify, extension.DefinitionList, extension.Footnote,
			extension.Strikethrough, extension.Typographer,
			highlighting.Highlighting),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithHardWraps(), html.WithXHTML()))

	//// preload templates
	////templateFiles := make([]string, 0)
	//for _, tf := range
	//	//templateFiles = append(templateFiles, filepath.Join(templateFolder, tf))
	//	tmplContent :=
	//}
	//templates := template.Must(template.ParseFiles(templateFiles...))

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

	return &Server{
		rootFolder:   rootFolder,
		homeTopic:    homeTopic,
		viewTemplate: viewTemplate,
		editTemplate: editTemplate,
		md:           &md,
		box:          box,
	}
}

func (sr *Server) Run() {
	log.Printf("Starting server with root: '%s' and home: '%s'\n",
		sr.rootFolder, sr.homeTopic)
	http.HandleFunc("/", sr.homePage)
	http.HandleFunc("/view/", makeHandler(sr.viewHandler))
	http.HandleFunc("/edit/", makeHandler(sr.editHandler))
	http.HandleFunc("/save/", makeHandler(sr.saveHandler))

	fs := http.FileServer(http.Dir(sr.rootFolder))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func (sr *Server) buildLinks() {

}
