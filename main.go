package main

import (
	"flag"
	"fmt"
	"log"
	"slack-backer-upper/archive"
	"slack-backer-upper/server"
	"slack-backer-upper/storage"
)

var (
	dirname = flag.String("d", "", "a directory to import")
	zipname = flag.String("z", "", "a zip file to import")
)

func slackBackerUpper() error {
	flag.Parse()

	s, err := storage.New()
	if err != nil {
		return fmt.Errorf("Error initializing data storage: %v", err)
	}
	defer s.Close()

	as, err := storage.Archiver(s)
	if err != nil {
		return fmt.Errorf("Error initializing archive storage: %v", err)
	}
	defer as.Close()
	a := archive.New(as)

	if *zipname != "" {
		if err = a.ImportZipFile(*zipname); err != nil {
			return fmt.Errorf("Error importing zip file: %v", err)
		}
		return nil
	}
	if *dirname != "" {
		if err = a.ImportFolder(*dirname); err != nil {
			return fmt.Errorf("Error importing folder: %v", err)
		}
		return nil
	}
	vs, err := storage.Viewer(s)
	if err != nil {
		return fmt.Errorf("Error initializing viewer storage: %v", err)
	}
	defer vs.Close()
	srv := server.New(&a, vs)
	return srv.Start()
}

func main() {
	if err := slackBackerUpper(); err != nil {
		log.Fatal(err)
	}
}
