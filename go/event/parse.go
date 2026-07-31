package event

import (
	"regexp"
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

// ParseEventPage extracts schedule layout details into an Event struct.
func ParseEventPage(htmlContent string) (*Event, error) {

	doc := soup.HTMLParse(htmlContent)
	event := new(Event)

	titleSpan := doc.Find("span", "class", "session-title")

	if titleSpan.Error != nil {
		return nil, titleSpan.Error
	}

	event.Title = strings.TrimSpace(titleSpan.Text())

	dateDiv := doc.Find("div", "class", "list-single__date")

	if dateDiv.Error != nil {
		return nil, dateDiv.Error
	}

	rawDateTime := dateDiv.Text()

	cleanDateTime := strings.TrimSuffix(rawDateTime, "PDT")
	cleanDateTime = strings.TrimSuffix(cleanDateTime, "EDT")
	cleanDateTime = strings.TrimSpace(cleanDateTime)

	// Collapse multi-spaces into single predictable spaces
	fields := strings.Fields(cleanDateTime)
	if len(fields) >= 5 {
		// e.g., ["Friday", "October", "23,", "2026", "1:45pm", "-", "2:30pm"]
		event.Date = strings.Join(fields[0:4], " ")
		event.Time = strings.Join(fields[4:], " ")
	} else {
		event.Date = cleanDateTime
	}

	locationDiv := doc.Find("div", "class", "list-single__location")

	if locationDiv.Error != nil {
		return nil, locationDiv.Error
	}

	if venueLink := locationDiv.Find("a"); venueLink.Error == nil {
		event.Location = strings.TrimSpace(venueLink.Text())
	}

	detailsDiv := doc.Find("div", "class", "sched-event-details")

	if detailsDiv.Error != nil {
		return nil, detailsDiv.Error
	}

	// Target and strip out the check-in sub-container if present in the tree node
	if checkinDiv := detailsDiv.Find("div", "id", "checkin-success"); checkinDiv.Error == nil {
		// Wipe out text within this element subtree or bypass its node structure entirely
		// Because soup lacks a true node deletion API, we fall back to an isolated match
		// below by focusing specifically on the content strings or splitting it cleanly.
	}

	fullText := detailsDiv.FullText()

	// Always remove the explicit notification headers unconditionally before processing
	if strings.Contains(fullText, "You're checked in!") {
		parts := strings.Split(fullText, "You're checked in!")
		if len(parts) >= 2 {
			fullText = parts[1]
		}
	}

	// Also remove trailing system boilerplate layout strings if present
	if idx := strings.Index(fullText, "Sign up or log in"); idx != -1 {
		fullText = fullText[:idx]
	}
	if idx := strings.Index(fullText, "Tweet"); idx != -1 {
		fullText = fullText[:idx]
	}

	// Strip out the implicit date/time values text echoes trailing at the end of text dumps
	if event.Date != "" {
		if idx := strings.Index(fullText, event.Date); idx != -1 {
			fullText = fullText[:idx]
		}
	}

	// Separate Speakers section out from the main text body flow safely
	if parts := strings.Split(fullText, "Speakers"); len(parts) >= 2 {
		event.Blurb = sanitizeWhitespace(parts[0])
		speakerSection := parts[1]

		// Parse isolated lines to look for speaker profiles
		lines := strings.Split(speakerSection, "\n")
		for _, line := range lines {
			line = sanitizeWhitespace(line)
			if line == "" {
				continue
			}

			nameEnd := len(line)
			indicators := []string{"Technical Project", "Assistant Director", "Director of", "Digital Accessibility", "Manager"}
			for _, ind := range indicators {
				if idx := strings.Index(line, ind); idx != -1 && idx < nameEnd {
					nameEnd = idx
				}
			}

			speakerName := strings.TrimSpace(line[:nameEnd])
			if len(speakerName) > 2 && len(speakerName) < 40 {
				event.Speakers = append(event.Speakers, speakerName)
			}
		}
	} else {
		// If no "Speakers" marker exists, the cleaned text block is completely the Blurb
		event.Blurb = sanitizeWhitespace(fullText)
	}

	return event, nil
}

// Helper function to replace non-breaking spaces and clean messy whitespace blocks
func sanitizeWhitespace(s string) string {
	// Replace non-breaking spaces (\u00a0) with regular ASCII spaces
	s = strings.ReplaceAll(s, "\u00a0", " ")

	spaceRegex := regexp.MustCompile(`\s+`)
	s = spaceRegex.ReplaceAllString(s, " ")

	return strings.TrimSpace(s)
}
