package main

import (
	"log"
	"os"
	"slack-backer-upper/files"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("usage: go run main.go <slack archive> [-z]")
	}
	if len(os.Args) > 2 {
		if os.Args[2] == "-z" {
			if err := files.Unzip(os.Args[1], "_archive"); err != nil {
				log.Fatalf("Error unzipping file: %v\n", err)
			}
			if err := files.ImportFolder("_archive"); err != nil {
				log.Printf("Error importing backup: %v\n", err)
			}
			if err := os.RemoveAll("_archive"); err != nil {
				log.Fatalf("Error removing temp dir: %v\n", err)
			}
		} else {
			log.Fatalln("usage: go run main.go <slack archive> [-z]")
		}
	} else if err := files.ImportFolder(os.Args[1]); err != nil {
		log.Printf("Error importing backup: %v\n", err)
	}
}
