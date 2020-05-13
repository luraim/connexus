package main

import (
	"connexus/server"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if len(os.Args) != 4 {
		log.Fatalln("please provide root folder, main topic and port number")
	}

	rootFolder, err := filepath.Abs(os.Args[1])
	if err != nil {
		log.Fatalln("error getting absolute path to root folder")
	}

	homeTopic := os.Args[2]
	port := os.Args[3]

	server.NewServer(rootFolder, homeTopic, port).Run()

}

func main2() {
	rootFolder, err := filepath.Abs("testwiki")
	if err != nil {
		log.Fatalln("error getting absolute path to root folder")
	}
	homeTopic := "test1"
	port := "46723"
	s := server.NewServer(rootFolder, homeTopic, port)
	fmt.Println(s.PageExists(homeTopic))
}
