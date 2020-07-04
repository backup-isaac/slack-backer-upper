package main

import (
	"flag"
	"log"
	"slack-backer-upper/archive"
	"slack-backer-upper/server"
	"slack-backer-upper/storage"
)

var (
	dirname = flag.String("d", "", "a directory to import")
	zipname = flag.String("z", "", "a zip file to import")
)

func main() {
	flag.Parse()

	s, err := storage.New()
	if err != nil {
		log.Fatalf("Error initializing data storage: %v", err)
	}
	defer s.Close()

	as, err := storage.Archiver(s)
	if err != nil {
		log.Fatalf("Error initializing archive storage: %v", err)
	}
	defer as.Close()
	a := archive.New(as)

	if *zipname != "" {
		if err := a.ImportZipFile(*zipname); err != nil {
			log.Fatalf("Error unzipping file: %v\n", err)
		}
	} else if *dirname != "" {
		if err := a.ImportFolder(*dirname); err != nil {
			log.Fatalf("Error importing backup: %v\n", err)
		}
	} else {
		vs, err := storage.Viewer(s)
		if err != nil {
			log.Fatalf("Error initializing viewer storage: %v", err)
		}
		defer vs.Close()
		srv := server.New(&a, vs)

		if err := srv.Start(); err != nil {
			log.Fatal(err)
		}
	}
}
