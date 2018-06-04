package carpe

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// For all tags on all days, this is everything.
type Survey struct {
	TagDayHash map[string]*TagSurvey
}

// For one tag on all days.
type TagSurvey struct {
	DayHash map[int]*TagDaySurvey
}

// For one tag on one day.
type TagDaySurvey struct {
	Surveys []*SurveyRec
}

func NewSurvey() *Survey {
	return &Survey{
		TagDayHash: make(map[string]*TagSurvey),
	}
}

// Survey uses simple shell scripts, so don't have Spaces or Newlines or Exploits or web-user-provided data in the file paths.
func (o *Survey) Walk(spool string) {
	// Insist that the spool ends in '/', so we can use `find`.

	if spool[len(spool)-1] != '/' {
		log.Panicf("Spool must end in '/' to Survey: %q", spool)
	}

	// Sorting just makes things more predictable, for debugging.
	// script := fmt.Sprintf("find '%s' -type f -name '*.jpg' -print | sort | xargs file", spool)
	script := fmt.Sprintf("set -x; find '%s' -type f -name '*.jpg' -print | xargs file", spool)
	cmd := exec.Command("bash", "-c", script)
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Panicf("Cannot fork script: %q", script)
	}

	r2 := bufio.NewReader(r)
	cmd.Start()
	for {
		line, err := r2.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Error reading from StdoutPipe: %v", err)
		}
		o.handleSurveyLine(strings.TrimRight(line, "\n"))
	}
}

// Example line:
// spool/wd4elg.40.d/wd4elg.40.2018-06-02-174220.jpg:       JPEG image data, JFIF standard 1.01, aspect ratio, density 1x1, segment length 16, baseline, precision 8, 1789x997, frames 3

const d4 = "2[0-9][0-9][0-9]"
const d2 = "-[0-9][0-9]"
const d6 = "-[0-9][0-9][0-9][0-9][0-9][0-9]"
const datePattern = d4 + d2 + d2 + d6

const surveryLinePattern = "^([^:]*/([^/:]+)[.]d/([^/:]+)[.](" + datePattern + ")[.]jpg):.*JPEG.*, ([0-9]+)x([0-9]+),.*$"

var surveyLineMatch = regexp.MustCompile(surveryLinePattern).FindStringSubmatch

type SurveyRec struct {
	Filename   string
	Tag        string
	TimeString string
	Time       time.Time
	Width      int
	Height     int
}

func (o *Survey) handleSurveyLine(line string) {
	m := surveyLineMatch(line)
	if m != nil {
		filename := m[1]
		tag1 := m[2]
		tag2 := m[3]
		timestamp := m[4]
		width := m[5]
		height := m[6]

		if tag1 != tag2 {
			log.Panicf("BAD %q != %q in %q", tag1, tag2, line)
		}

		w_, _ := strconv.ParseInt(width, 10, 64)
		h_, _ := strconv.ParseInt(height, 10, 64)
		rec := &SurveyRec{
			Filename:   filename,
			Tag:        tag1,
			TimeString: timestamp,
			Width:      int(w_),
			Height:     int(h_),
		}
		const pattern = "2006-01-02-150405"
		t, err := time.Parse(pattern, timestamp)
		if err != nil {
			log.Printf("Cannot parse timestamp: %q: %v", timestamp, err)
			return
		}
		t = t.UTC()
		rec.Time = t
		unix := t.Unix()
		day := int(unix / 86400)
		h := o.TagDayHash[tag1]
		if h == nil {
			h = &TagSurvey{
				DayHash: make(map[int]*TagDaySurvey),
			}
			o.TagDayHash[tag1] = h
		}

		h2 := h.DayHash[day]
		if h2 == nil {
			h2 = &TagDaySurvey{}
			h.DayHash[day] = h2
		}
		h2.Surveys = append(h2.Surveys, rec)
	} else {
		log.Printf("Failed match: %q", line)
	}
}
