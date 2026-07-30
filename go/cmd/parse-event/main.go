package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/aaronland/mcn/event"
)

func main() {

	flag.Parse()

	for _, path := range flag.Args() {

		body, err := os.ReadFile(path)

		if err != nil {
			log.Fatalf("Failed to read %s, %v", path, err)
		}

		e, err := event.ParseEvent(string(body))

		if err != nil {
			log.Fatalf("Failed to parse event, %v", err)
		}

		enc := json.NewEncoder(os.Stdout)
		err = enc.Encode(e)

		if err != nil {
			log.Fatalf("Failed to encode event, %v", err)
		}
	}

}
