package main

import (
	"flag"
	"log"
	"os"
	"slack-backer-upper/files"
)

var (
	dirname = flag.String("d", "", "a directory to import")
	zipname = flag.String("z", "", "a zip file to import")
)

func main() {
	flag.Parse()
	if *zipname != "" {
		if err := files.Unzip(*zipname, "_archive"); err != nil {
			log.Fatalf("Error unzipping file: %v\n", err)
		}
		if err := files.ImportFolder("_archive"); err != nil {
			log.Printf("Error importing backup: %v\n", err)
		}
		if err := os.RemoveAll("_archive"); err != nil {
			log.Fatalf("Error removing temp dir: %v\n", err)
		}
	} else if *dirname != "" {
		if err := files.ImportFolder(*dirname); err != nil {
			log.Printf("Error importing backup: %v\n", err)
		}
	} else {
		log.Printf("Starting HTTP server...")
	}
}
