package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aaronland/mcn/event"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		abs_path, err := filepath.Abs(path)

		if err != nil {
			log.Fatalf("Failed to derive absolute path for %s, %v", path, err)
		}

		body, err := os.ReadFile(abs_path)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", abs_path, err)
		}

		e, err := event.ParseEventPage(string(body))

		if err != nil {
			log.Printf("Failed to parse event, %v", err)
			continue
		}

		ext := filepath.Ext(abs_path)

		json_path := strings.Replace(abs_path, ext, ".json", 1)

		wr, err := os.OpenFile(json_path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			log.Fatalf("Failed to open %s for writing, %v", json_path, err)
		}

		enc := json.NewEncoder(wr)
		err = enc.Encode(e)

		if err != nil {
			log.Fatalf("Failed to encode event, %v", err)
		}

		err = wr.Close()

		if err != nil {
			log.Fatalf("Failed to close %s after writing, %v", json_path, err)
		}
	}

}
