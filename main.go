package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/palantir/stacktrace"

	"gopkg.in/russross/blackfriday.v2"
)

const homeMdFile = "test1.md"
const css = "style.css"

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
}

func newServer(rootFolder string) *Server {
	return &Server{rootFolder: rootFolder}
}

func (sr *Server) handler(w http.ResponseWriter, r *http.Request) {
	if len(r.URL.Path) == 1 {
		log.Println("rendering home page:", homeMdFile)
		fmt.Fprintf(w, sr.md2html(homeMdFile))
	} else {
		fileName := r.URL.Path[1:]
		fmt.Println("file:", fileName)
		if strings.HasSuffix(fileName, ".md") {
			log.Println("rendering md file:", fileName)
			fmt.Fprintf(w, sr.md2html(fileName))
		} else {
			contents, err := ioutil.ReadFile(fileName)
			if err != nil {
				log.Println(stacktrace.Propagate(err, "error reading %s", fileName))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			fmt.Fprintf(w, string(contents))
		}
	}
	//fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func (sr *Server) md2html(mdFile string) string {
	input, err := ioutil.ReadFile(filepath.Join(sr.rootFolder, mdFile))
	if err != nil {
		log.Fatalln(stacktrace.Propagate(err, "error reading md file"))
	}

	output := blackfriday.Run(input)
	return PRE + string(output) + POST
}

func (sr *Server) run() {
	http.HandleFunc("/", sr.handler)
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
