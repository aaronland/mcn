// fetch-events parses a local .ics (calendar) file containing MCN schedule data. Each event URL in the
// calendar file is retrieved from the web and the body of the div@sched-content-inner element is written
// to a local HTML file. Note that this program, as written, assumes event data is being retrieved from
// Sched (sched.com) and that relevant session (event) data is stored in the aforementioned element. This
// program does not attempt to derive any structured data from the HTML data it retrieves. That is left for
// another tool to do. For example:
//
//	$> go run cmd/fetch-events/main.go -ics ../conference/2020/schedule.ics -destination ../conference/2020/events/
package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/anaskhan96/soup"
	"github.com/arran4/golang-ical"
)

func main() {

	var path_ics string
	var dest string

	flag.StringVar(&path_ics, "ics", "", "The path to a local MCN schedule '.ics' file.")
	flag.StringVar(&dest, "destination", "", "The folder where event data in the local MCN schedule '.ics' file will be written.")

	flag.Parse()

	abs_root, err := filepath.Abs(dest)

	if err != nil {
		log.Fatalf("Failed to derive absolute path for %s, %v", dest, err)
	}

	r, err := os.Open(path_ics)

	if err != nil {
		log.Fatalf("Failed to open %s for reading, %v", path_ics, err)
	}

	defer r.Close()

	cal, err := ics.ParseCalendar(r)

	if err != nil {
		log.Fatalf("Failed to parse %s, %v", path_ics, err)
	}

	for _, c := range cal.Components {

		switch c.(type) {
		case *ics.VEvent:
			// pass
		default:
			continue
		}

		v := c.(*ics.VEvent)
		id := v.Id()
		url := v.GetProperty("URL").Value

		logger := slog.Default()
		logger = logger.With("id", id)
		logger = logger.With("url", url)

		logger.Info("Process event")

		ev, err := fetchEvent(url)

		if err != nil {
			logger.Error("Failed to retrieve event", "error", err)
			continue
		}

		fname := fmt.Sprintf("%s.html", id)
		path := filepath.Join(abs_root, fname)

		logger = logger.With("path", path)

		_, err = os.Stat(path)

		if err == nil {
			logger.Info("Event file already exists, skipping")
			continue
		}
		
		wr, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0644)

		if err != nil {
			logger.Error("Failed to open file for writing", "error", err)
			continue
		}

		_, err = wr.Write([]byte(ev))

		if err != nil {
			logger.Error("Failed to write event", "error", err)
			continue
		}

		err = wr.Close()

		if err != nil {
			logger.Error("Failed to close after writing", "error", err)
			continue
		}
	}
}

func fetchEvent(url string) (string, error) {

	rsp, err := soup.Get(url)

	if err != nil {
		return "", fmt.Errorf("Failed to retrieve %s, %w", url, err)
	}

	doc := soup.HTMLParse(rsp)

	ev := doc.Find("div", "id", "sched-content-inner")
	return ev.HTML(), nil
}
