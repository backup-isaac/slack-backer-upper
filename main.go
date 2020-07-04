package main

import (
	"flag"
	"log"
	"slack-backer-upper/archive"
	"slack-backer-upper/server"
)

var (
	dirname = flag.String("d", "", "a directory to import")
	zipname = flag.String("z", "", "a zip file to import")
)

func main() {
	flag.Parse()
	if *zipname != "" {
		if err := archive.ImportZipFile(*zipname); err != nil {
			log.Fatalf("Error unzipping file: %v\n", err)
		}
	} else if *dirname != "" {
		if err := archive.ImportFolder(*dirname); err != nil {
			log.Printf("Error importing backup: %v\n", err)
		}
	} else if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
