package main

import (
	"connexus/server"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

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

	server.NewServer(rootFolder, homeTopic).Run()

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
