package carpe

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"time"

	"github.com/strickyak/resize"
	"github.com/strickyak/rxtx/font5x7"
)

type SurveyRec struct {
	Filename    string
	Filesize    int64
	Tag         string
	TimeString  string
	ShapeString string
	Time        time.Time
	Width       int
	Height      int
}

// For all tags on all days, this is everything.
type Survey struct {
	Spool      string
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

func NewSurvey(spool string) *Survey {
	return &Survey{
		Spool:      spool,
		TagDayHash: make(map[string]*TagSurvey),
	}
}

func RenameFileForImageSize(spool string, filename string) (string, *SurveyRec) {
	script := fmt.Sprintf("set -x; file %q", filename)
	cmd := exec.Command("bash", "-c", script)
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Panicf("Cannot fork script: %q", script)
	}
	cmd.Start()

	bb, err := ioutil.ReadAll(r)
	if err != nil {
		log.Panicf("Cannot read script output: %q", script)
	}
	result := string(bb)

	rec := ParseSurveyLine(result)
	if rec == nil {
		log.Printf("ParseSurveyLine failed on %q", result)
		return "", nil
	}

	newdir := fmt.Sprintf("%s/%s~%dx%d.d", spool, rec.Tag, rec.Width, rec.Height)
	os.MkdirAll(newdir, 0755)
	newname := fmt.Sprintf("%s/%s.%dx%d.%s.jpg", newdir, rec.Tag, rec.Width, rec.Height, rec.TimeString)
	err = os.Rename(filename, newname) // If this fails, it probably already had the corect name.
	if err != nil {
		log.Printf("Cannot rename %q to %q: %v", filename, newname, err)
	}
	return newname, rec
}

// Survey uses simple shell scripts, so don't have Spaces or Newlines or Exploits or web-user-provided data in the file paths.
func (o *Survey) Walk() {
	// Insist that the spool ends in '/', so we can use `find`.

	if o.Spool[len(o.Spool)-1] != '/' {
		log.Panicf("Spool must end in '/' to Survey: %q", o.Spool)
	}

	// Sorting just makes things more predictable, for debugging.
	script := fmt.Sprintf("set -x; find '%s' -type f -name '*.jpg' -print | sort | xargs file", o.Spool)
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
		o.handleSurveyLine(line)
	}
}

// Example line:
// spool/wd4elg.40.d/wd4elg.40.2018-06-02-174220.jpg:       JPEG image data, JFIF standard 1.01, aspect ratio, density 1x1, segment length 16, baseline, precision 8, 1789x997, frames 3

const d4 = "2[0-9][0-9][0-9]"
const d2 = "-[0-9][0-9]"
const d6 = "-[0-9][0-9][0-9][0-9][0-9][0-9]"
const datePattern = "(" + d4 + d2 + d2 + d6 + ")"
const widthHeightPattern = "([0-9]+)x([0-9]+)"

const surveryLinePattern = "^([^:]*/([^/:]+)[.]d/([^/:]+)[.]" + widthHeightPattern + "[.]" + datePattern + "[.]jpg):.*JPEG.*, ([0-9]+)x([0-9]+),.*\n?$"

var surveyLineMatch = regexp.MustCompile(surveryLinePattern).FindStringSubmatch

func ParseSurveyLine(line string) *SurveyRec {
	m := surveyLineMatch(line)
	if m == nil {
		return nil
	}

	filename := m[1]
	dir := m[2]
	tag := m[3]
	oldWidth := m[4]
	oldHeight := m[5]
	timestamp := m[6]
	width := m[7]
	height := m[8]
	_, _, _ = dir, oldWidth, oldHeight

	fileinfo, err := os.Stat(filename)
	if err != nil {
		log.Printf("CANNOT STAT: %q: %v", filename, err)
		return nil
	}
	filesize := fileinfo.Size()
	if filesize < 512 {
		log.Printf("Unreasonably small: %q: %d", filename, filesize)
		return nil
	}

	w_, _ := strconv.ParseInt(width, 10, 64)
	h_, _ := strconv.ParseInt(height, 10, 64)
	rec := &SurveyRec{
		Filename:   filename,
		Filesize:   filesize,
		Tag:        tag,
		TimeString: timestamp,
		Width:      int(w_),
		Height:     int(h_),
	}
	const pattern = "2006-01-02-150405"
	t, err := time.Parse(pattern, timestamp)
	if err != nil {
		log.Printf("Cannot parse timestamp: %q: %v", timestamp, err)
		return nil
	}
	rec.Time = t.UTC()
	return rec
}

func (o *Survey) handleSurveyLine(line string) {
	rec := ParseSurveyLine(line)
	if rec == nil {
		log.Printf("SKIPPING unparsable line: %q", line)
		return
	}
	newname, _ := RenameFileForImageSize(o.Spool, rec.Filename)
	newname, rec3 := RenameFileForImageSize(o.Spool, newname)
	rec = rec3

	tag := fmt.Sprintf("%s~%dx%d", rec.Tag, rec.Width, rec.Height)
	unix := rec.Time.Unix()
	day := int(unix / 86400)
	h := o.TagDayHash[tag]
	if h == nil {
		h = &TagSurvey{
			DayHash: make(map[int]*TagDaySurvey),
		}
		o.TagDayHash[tag] = h
	}

	h2 := h.DayHash[day]
	if h2 == nil {
		h2 = &TagDaySurvey{}
		h.DayHash[day] = h2
	}
	h2.Surveys = append(h2.Surveys, rec)
}

func (o *Survey) BuildMovies(prefix string) {
	for k1, v1 := range o.TagDayHash {
		fmt.Printf("A %q\n", k1)
		for k2, v2 := range v1.DayHash {
			if len(v2.Surveys) < 2 {
				continue
			}

			var estimatedSize int64
			digest := md5.New()
			var inputs []string
			for _, v := range v2.Surveys {
				inputs = append(inputs, v.Filename)
				digest.Write([]byte(v.Filename))
				estimatedSize += v.Filesize
			}
			digestStr := fmt.Sprintf("%X", digest.Sum(nil))

			tmpname := fmt.Sprintf("%s/%s.d/%s.%d.%s.tmp", o.Spool, k1, prefix, k2, digestStr)
			outname := fmt.Sprintf("%s/%s.d/%s.%d.%s.gif", o.Spool, k1, prefix, k2, digestStr)

			_, err := os.Stat(outname)
			if err == nil {
				log.Printf("Already exists: %q", outname)
				continue
			}

			log.Printf("Building gif from %d inputs estimatedSize %d (%.3f MiB): %q", len(inputs), estimatedSize, float64(estimatedSize)/1024/1024, outname)
			o.Build1Giffy(inputs, tmpname, outname)
		}
	}
}

func (o *Survey) Build1Giffy(inputs []string, tmpname, outname string) (ok bool) {
	ok = true
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Recovering after panic in BuildAnimatedGif %q: %v", outname, r)
			ok = false
		}
	}()
	BuildAnimatedGif(inputs, 200*time.Millisecond, o.ConvertToModest, tmpname, tmpname+".mean.png")
	err := os.Rename(tmpname, outname)
	if err != nil {
		log.Panicf("Cannot rename %q to %q: %v", tmpname, outname, err)
	}
	log.Printf("Renamed to %q", outname)
	return
}

// A modest size for video frames.
const WID = 800
const HEI = 500

var GREEN = image.NewUniform(color.NRGBA{20, 200, 20, 255})

func (o *Survey) ConvertToModest(img image.Image, filename string) image.Image {
	t := resize.Thumbnail(WID, HEI, img, resize.Bilinear)
	b := t.Bounds()
	width := b.Max.X - b.Min.X
	height := b.Max.Y - b.Min.Y

	zb := image.Rectangle{
		Max: image.Point{WID, 20 + HEI},
	}
	z := image.NewRGBA(zb)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			z.Set(x, y, t.At(x, y))
		}
	}
	for i, ch := range path.Base(filename) {
		for r := 0; r < 8; r++ {
			for c := 0; c < 5; c++ {
				if i*7+c+10 > WID/2-10 {
					continue
				}
				if font5x7.Pixel(byte(ch), r, c) {
					z.Set(2*(i*7+c+10), (HEI + 2 + 2*r), GREEN)
					z.Set(1+2*(i*7+c+10), (HEI + 2 + 2*r), GREEN)
					z.Set(2*(i*7+c+10), 1+(HEI+2+2*r), GREEN)
					z.Set(1+2*(i*7+c+10), 1+(HEI+2+2*r), GREEN)
				}
			}
		}
	}
	return z
}
