// event/parse.go
package event

import (
	"strings"

	"github.com/anaskhan96/soup"
)

// Event holds the information we want to extract.
type Event struct {
	Title    string   // e.g. "Museums and AI: Findings from a National Convening"
	Blurb    string   // the long description
	Date     string   // e.g. "Friday October 23, 2026"
	Time     string   // e.g. "2:10pm – 2:30pm"
	Location string   // e.g. "512 Willapa"
	Speakers []string // e.g. ["Kate Haley Goldman"]
}

// ParseEvent parses an HTML snippet and returns an Event struct.
func ParseEvent(html string) (*Event, error) {
	doc := soup.HTMLParse(html)

	// ---- Title -------------------------------------------------------
	title := ""
	if els := doc.FindAll("span", "class", "session-title"); len(els) > 0 {
		title = strings.TrimSpace(els[0].Text())
	}

	// ---- Blurb -------------------------------------------------------
	blurb := ""
	if els := doc.FindAll("div", "class", "tip-description"); len(els) > 0 {
		blurb = strings.TrimSpace(els[0].Text())
	}

	// ---- Date & Time ---------------------------------------------------
	date, time := "", ""
	if els := doc.FindAll("div", "class", "list-single__date"); len(els) > 0 {
		raw := strings.TrimSpace(els[0].Text())
		if parts := strings.SplitN(raw, "  ", 2); len(parts) == 2 {
			date = strings.TrimSpace(parts[0])

			tp := strings.TrimSpace(parts[1])
			// strip trailing 3‑letter TZ (e.g. “PDT”)
			if i := strings.LastIndex(tp, " "); i != -1 && len(tp[i:]) == 4 {
				tp = tp[:i]
			}
			tp = strings.ReplaceAll(tp, "  ", " ")
			time = strings.TrimSpace(tp)
		}
	}

	// ---- Location ----------------------------------------------------
	location := ""
	if els := doc.FindAll("div", "class", "list-single__location"); len(els) > 0 {
		if aels := els[0].FindAll("a"); len(aels) > 0 {
			location = strings.TrimSpace(aels[0].Text())
		}
	}

	// ---- Speakers ----------------------------------------------------
	speakers := []string{}
	for _, s := range doc.FindAll("div", "class", "sched-person-session") {
		if els := s.FindAll("h2", "a"); len(els) > 0 {
			name := strings.TrimSpace(els[0].Text())
			if name != "" {
				speakers = append(speakers, name)
			}
		}
	}

	return &Event{
		Title:    title,
		Blurb:    blurb,
		Date:     date,
		Time:     time,
		Location: location,
		Speakers: speakers,
	}, nil
}
