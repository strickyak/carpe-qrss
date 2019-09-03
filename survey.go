package carpe

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math/rand"
	"os"
	P "path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/strickyak/resize"
	"github.com/strickyak/rxtx/font5x7"
)

const WHOM = "W6REK"
const MAX_GIF = 200 // Problems with Out Of Memory.

const timestampPattern = "2006-01-02-150405"

func init() {
	rand.Seed(time.Now().Unix())
}

func DieIf(err error, args ...interface{}) {
	if err != nil {
		log.Printf("FATAL...: %v: %#v", err, args)
		debug.PrintStack()
		log.Fatalf("...FATAL: %v: %#v", err, args)
	}
}

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
	SeenOther  map[string]bool // Seen for mark-and-sweep
	UsedOther  map[string]bool // Marked for mark-and-sweep
}

type Products struct {
	MovieName    string
	MovieModTime time.Time
	MeanName     string
	MeanModTime  time.Time
}

// For one tag on all days.
type TagSurvey struct {
	DayHash     map[int]*TagDaySurvey
	NewProducts map[int]Products // int is Days Ago, 0 is today.
}

// For one tag on one day.
type TagDaySurvey struct {
	Surveys []*SurveyRec
}

func NewSurvey(spool string) *Survey {
	return &Survey{
		Spool:      spool,
		TagDayHash: make(map[string]*TagSurvey),
		SeenOther:  make(map[string]bool),
		UsedOther:  make(map[string]bool),
	}
}

// Survey uses simple shell scripts, so don't have Spaces or Newlines or Exploits or web-user-provided data in the file paths.
func (o *Survey) Walk() {
	spoolBeyondSymlink := o.Spool + "/."
	P.Walk(spoolBeyondSymlink, o.WalkFunc)
}

func (o *Survey) WalkFunc(filename string, info os.FileInfo, err error) error {
	if err != nil {
		log.Fatalf("Fatal: WalkFunc gets error for %q: %v", filename, err)
	}
	if info.IsDir() {
		return nil // Not necessary to visit directories.
	}

	o.handleSurveyFilename(filename, info)
	return nil
}

// Example line:
// spool/wd4elg.40.d/wd4elg.40.2018-06-02-174220.jpg:       JPEG image data, JFIF standard 1.01, aspect ratio, density 1x1, segment length 16, baseline, precision 8, 1789x997, frames 3

const d4 = "2[0-9][0-9][0-9]"
const d2 = "-[0-9][0-9]"
const d6 = "-[0-9][0-9][0-9][0-9][0-9][0-9]"
const datePattern = "(" + d4 + d2 + d2 + d6 + ")"
const widthHeightPattern = "([0-9]+)x([0-9]+)"

const primaryPattern = "^(.+)[.]" + widthHeightPattern + "[.]" + datePattern + "[.]jpg$"
//const stackPattern = "^(.+)[.]" + widthHeightPattern + "[.]" + datePattern + "[.]jpg[.]png$"

var primaryMatch = regexp.MustCompile(primaryPattern).FindStringSubmatch
//var stackMatch = regexp.MustCompile(stackPattern).FindStringSubmatch

// Originals filenames always end in `.jpg` (regardless of their image type).
func ParseFilenameForPrimary(filename string, info os.FileInfo) *SurveyRec {
	fd, err := os.Open(filename)
	if err != nil {
		log.Printf("ParseFilenameForPrimary cannot open %q: %v", filename, err)
		return nil
	}
	defer fd.Close()
	c, _, err := image.DecodeConfig(fd)
	if err != nil {
		log.Printf("ParseFilenameForPrimary cannot DecodeConfig %q: %v", filename, err)
		return nil
	}

	m := primaryMatch(P.Base(filename))
	if m == nil {
		log.Printf("ParseFilenameForPrimary cannot regexp match %q: %v", filename, err)
		return nil
	}

	tag := m[1]
	oldWidth := m[2]
	oldHeight := m[3]
	timestamp := m[4]
	_, _ = oldWidth, oldHeight

	filesize := info.Size()
	if filesize < 512 {
		log.Printf("Unreasonably small: %q: %d", filename, filesize)
		return nil
	}

	rec := &SurveyRec{
		Filename:   filename,
		Filesize:   filesize,
		Tag:        tag,
		TimeString: timestamp,
		Width:      c.Width,
		Height:     c.Height,
	}
	t, err := time.Parse(timestampPattern, timestamp)
	if err != nil {
		log.Printf("Cannot parse timestamp: %q: %v", timestamp, err)
		return nil
	}
	rec.Time = t.UTC()
	return rec
}

func (o *Survey) handleSurveyFilename(filename string, info os.FileInfo) {
	if info.IsDir() {
		return // Don't process directories.
	}

	rec := ParseFilenameForPrimary(filename, info)
	if rec == nil {
		o.SeenOther[filename] = true
		log.Printf("Not Primary: %q", filename)
		return
	}
	// Mark possible stack names.
	stackname := P.Join(P.Dir(filename), "stack." + P.Base(filename) + ".png")
	o.UsedOther[stackname] = true

	tagAndShape := fmt.Sprintf("%s~%dx%d", rec.Tag, rec.Width, rec.Height)
	day := int(rec.Time.Unix() / 86400)

	h := o.TagDayHash[tagAndShape]
	if h == nil {
		h = &TagSurvey{
			DayHash:     make(map[int]*TagDaySurvey),
			NewProducts: make(map[int]Products),
		}
		o.TagDayHash[tagAndShape] = h
	}

	h2 := h.DayHash[day]
	if h2 == nil {
		h2 = &TagDaySurvey{}
		h.DayHash[day] = h2
	}
	h2.Surveys = append(h2.Surveys, rec)
}

const GC_GRACE_HOURS = 2

func (o *Survey) CollectGarbage() {
	log.Printf("Start CollectGarbage [[[")
	for filename, _ := range o.SeenOther {
		used, ok := o.UsedOther[filename]
		log.Printf("Consider %q (%v, %v)", filename, used, ok)
		if !used {
			info, err := os.Stat(filename)
			DieIf(err, "stat", filename)
			hoursOld := time.Since(info.ModTime()).Hours()
			if hoursOld < GC_GRACE_HOURS {
				log.Printf("Grace for %q; age is %v hours", filename, hoursOld)
				continue // Grace time
			}
			log.Printf("Removing garbage: %q", filename)
			err = os.Remove(filename)
			DieIf(err, "remove", filename)
		}
	}
	log.Printf("End CollectGarbage ]]]")
}

var qM = flag.Bool("qM", false, "DEBUG: quickly skip over movie code")

func (o *Survey) BuildMovies(prefix string) {
	var mutex sync.Mutex
	/*
		done := make(chan string)
	*/

	todaysDay := int(time.Now().Unix()) / 86400
	for k1, v1 := range o.TagDayHash {
		for k2, v2 := range v1.DayHash {
			/*go*/ func(k1 string, v1 *TagSurvey, k2 int, v2 *TagDaySurvey) {
				task := fmt.Sprintf("%s:%d", k1, k2)
				log.Printf("Task Starting: %s", task)

				if len(v2.Surveys) < 1 {
					/*
						done <- task
					*/
					return
				}

				daysAgo := todaysDay - k2

				var estimatedSize int64
				digest := md5.New()
				var inputs []string
				for _, v := range v2.Surveys {
					inputs = append(inputs, v.Filename)
					digest.Write([]byte(v.Filename))
					estimatedSize += v.Filesize
				}
				digestStr := fmt.Sprintf("%X", digest.Sum(nil))

				tmpgif := P.Clean(fmt.Sprintf("%s/%s.d/%s.%d.%s.tmp", o.Spool, k1, prefix, k2, digestStr))
				gifname := P.Clean(fmt.Sprintf("%s/%s.d/%s.%d.%s.gif", o.Spool, k1, prefix, k2, digestStr))
				meanname := P.Clean(fmt.Sprintf("%s/%s.d/%s.%d.%s.png", o.Spool, k1, prefix, k2, digestStr))
				mutex.Lock()
				{
					o.UsedOther[gifname] = true  // Save from garbage collection.
					o.UsedOther[meanname] = true // Save from garbage collection.

					v1.NewProducts[daysAgo] = Products{
						MovieName:    gifname,
						MovieModTime: time.Now(),
						MeanName:     meanname,
						MeanModTime:  time.Now(),
					}
				}
				mutex.Unlock()

				_, err := os.Stat(gifname)
				if err == nil {
					log.Printf("Already exists: %q", gifname)
					/*
						done <- task
					*/
					return
				}

				log.Printf("Building gif from %d inputs estimatedSize %d (%.3f MiB): %q", len(inputs), estimatedSize, float64(estimatedSize)/1024/1024, gifname)
				if *qM {
					log.Printf("qM: not calling o.Build1Giffy")
				} else {
					o.Build1Giffy(inputs, tmpgif, gifname, meanname)
				}

				/*
					done <- task
				*/
			}(k1, v1, k2, v2)
		}
	}

	/*
		// Now wait for them all to finish.
		for _, v1 := range o.TagDayHash {
			for _, _ = range v1.DayHash {
				task := <-done
				log.Printf("Task Finished: %s", task)
			}
		}
	*/
}

// ThinStrings drops 1 random input string.
func ThinStrings(a []string) []string {
	n := len(a)
	r := rand.Intn(n)
	z := make([]string, 0, n-1)
	for i := 0; i < n; i++ {
		if i != r {
			z = append(z, a[i])
		}
	}
	return z
}

func (o *Survey) Build1Giffy(inputs []string, tmpgif, gifname, meanname string) (ok bool) {
	ok = true
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("Recovering after panic in BuildAnimatedGif %q: %v", gifname, r)
			debug.PrintStack()
			ok = false
		}
	}()
	for len(inputs) > MAX_GIF {
		// Repeat dropping 1 string until right size.
		inputs = ThinStrings(inputs)
	}
	BuildAnimatedGif(inputs, 200*time.Millisecond, o.ConvertToModest, tmpgif, meanname)
	err := os.Rename(tmpgif, gifname)
	DieIf(err, "rename", tmpgif, gifname)
	log.Printf("Renamed Gif to ===>  %q  <===", gifname)
	return
}

// A modest size for video frames.
const WID = 800
const HEI = 500

var GREEN = image.NewUniform(color.NRGBA{20, 200, 20, 255})
var YELLOW = image.NewUniform(color.NRGBA{200, 200, 20, 255})

const timestampJpgPattern = "^.*" + datePattern + "[.]jpg$"

var timestampJpgMatch = regexp.MustCompile(timestampJpgPattern).FindStringSubmatch

func (o *Survey) ConvertToModest(img image.Image, filename string) image.Image {
	secsWithinDay := -1
	m := timestampJpgMatch(filename)
	if m != nil {
		log.Printf("timestampJpgMatch: %#v", m)
		t, err := time.Parse(timestampPattern, m[1])
		log.Printf("timestampJpgMatch: %#v [%v]", t, err)
		if err == nil {
			secsWithinDay = int(t.UTC().Unix() % 86400)
			log.Printf("timestampJpgMatch: %#v -> %d", t, secsWithinDay)
		}
	}

	t := resize.Thumbnail(WID, HEI, img, resize.Bilinear)
	b := t.Bounds()
	width := b.Max.X - b.Min.X
	height := b.Max.Y - b.Min.Y

	zb := image.Rectangle{
		Max: image.Point{WID, 25 + HEI},
	}
	z := image.NewRGBA(zb)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			z.Set(x, y, t.At(x, y))
		}
	}
	for i, ch := range P.Base(filename) {
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
	if secsWithinDay >= 0 {
		for i := 0; i < 6; i++ {
			for j := 0; j < 6; j++ {
				x := int(2 + (WID-8)*float64(secsWithinDay)/86400)
				z.Set(x+i, HEI+19+j, YELLOW)
			}
		}
	}
	return z
}

func RenameFileForImageSize(spool string, filename string) (string, *SurveyRec) {
	info, err := os.Stat(filename)
	if err != nil {
		log.Printf("RenameFileForImageSize cannot stat %q: %v", filename, err)
	}
	rec := ParseFilenameForPrimary(filename, info)
	if rec == nil {
		log.Printf("ParseFilenameForPrimary failed on %q", filename)
		return "", nil
	}

	newdir := P.Clean(fmt.Sprintf("%s/%s~%dx%d.d",
		spool, rec.Tag, rec.Width, rec.Height))
	os.MkdirAll(newdir, 0755)

	newname := P.Clean(fmt.Sprintf("%s/%s.%dx%d.%s.jpg",
		newdir, rec.Tag, rec.Width, rec.Height, rec.TimeString))
	err = os.Rename(filename, newname) // If this fails, it probably already had the corect name.
	if err != nil {
		log.Printf("Cannot rename %q to %q: %v", filename, newname, err)
	}
	return newname, rec
}

func (o *Survey) DumpProducts(w io.Writer) {
	for k1, k2 := range o.TagDayHash {
		fmt.Fprintf(w, "TAG: %q\n", k1)

		for i := 0; i < 14; i++ {
			p, ok := k2.NewProducts[i]
			if !ok {
				continue
			}
			movie := p.MovieName
			mean := p.MeanName
			fmt.Fprintf(w, "    [%2d] Movie: %q\n", i, movie)
			fmt.Fprintf(w, "    [%2d] Mean:  %q\n", i, mean)
		}
	}
}

func (o *Survey) SortedWebTags() []string {
	var tags []string
	for k1, k2 := range o.TagDayHash {
		active := false
		for i := 0; i < 32; i++ {
			p, ok := k2.NewProducts[i]
			if ok && (p.MovieName != "" || p.MeanName != "") {
				active = true
				break
			}
		}
		if active {
			tags = append(tags, k1)
		}
	}
	sort.Strings(tags)
	return tags
}

//func (o *Survey) WriteWebPage(w io.Writer) {
//	title := fmt.Sprintf(`Carpe QRSS at %s`, time.Now().UTC().Format(time.UnixDate))
//	fmt.Fprintf(w, `<html>
//<head>
//  <META NAME="ROBOTS" CONTENT="INDEX, NOFOLLOW">
//  <title>%s</title>
//</head>
//<body>
//
//<h3>%s</h3>
//<p>
//  <b>This is Experimental, Alpha quality.</b>
//  I'm debugging with a small set of QRSS Grabber sources.
//  I'll add more later.
//  For each source, data is grouped by Zulu day.
//  For each source on each day, there are two images, composed of frames
//  which are images seized from qrss grabbers.
//  The image on the left is the average of all the frames;
//  the one on the right is an animated GIF made up of all the frames.
//  A caption in green tells what the source was, prehaps the band,
//  and roughly what time the image was made or seized.
//<p>
//  <b>All times and dates are Zulu.</b>
//<p>
//  Source is at <a href="https://github.com/strickyak/carpe-qrss">https://github.com/strickyak/carpe-qrss</a>.
//  Site hosted in Digital Ocean.
//<p>
//  -- 7e3 de %s
//<p>
//<br>
//<br>
//`, title, title, WHOM)
//	tags := o.SortedWebTags()
//	for _, tag := range tags {
//		fmt.Fprintf(w, "[<a href='#%s'>%s</a>] &nbsp;\n", tag, tag)
//	}
//	fmt.Fprintf(w, `<p><br><br>\n`)
//
//	fmt.Fprintf(w, `
//<table cellpadding=5 border=1>
//  <tr>
//    <th>Days<br>Ago</th>
//    <th align=center>RGB-wise Average of Frames</th>
//    <th align=center>Animated GIF of Frames</th>
//  </tr>
//`)
//
//	for _, tag := range tags {
//		shortTag := strings.Split(tag, "~")[0]
//		fmt.Fprintf(w, `
//  <tr>
//    <th>Days<br>Ago</th>
//    <th colspan=2 align=center><a name="%s"><tt> <big><big><big>%s &nbsp; &nbsp; </big></big></big> <a href="%s.d/">%q</a> </tt></a></th>
//  </tr>
//`, tag, shortTag, tag, tag)
//
//		v1 := o.TagDayHash[tag]
//		n := 0
//		for i := 0; i < 7; i++ {
//			p, ok := v1.NewProducts[i]
//			if !ok {
//				continue
//			}
//			movie := p.MovieName
//			mean := p.MeanName
//			fmt.Fprintf(w, `
//  <tr>
//    <th align=center ><big>%d</big></th>
//    <td><img src="%s/%s"></td>
//    <td><img src="%s/%s"></td>
//  <tr>
//`, i, P.Base(P.Dir(mean)), P.Base(mean), P.Base(P.Dir(movie)), P.Base(movie))
//			n++
//			if n > 2 {
//				break
//			}
//		}
//	}
//	fmt.Fprintf(w, `
//</body></html>
//`)
//}

func (o *Survey) SortedWebTagsForDay(daysAgo int) []string {
	var tags []string
	for k1, k2 := range o.TagDayHash {
		p, ok := k2.NewProducts[daysAgo]
		if ok && (p.MovieName != "" || p.MeanName != "") {
			tags = append(tags, k1)
		}
	}
	sort.Strings(tags)
	return tags
}

func (o *Survey) WriteWebPageForDay(w io.Writer, daysAgo int) {
	todaysDay := time.Now().Unix() / 86400
	thatDay := todaysDay - int64(daysAgo)
	thatDate := time.Unix(thatDay*86400, 0).UTC()
	thatDateIsoDay := thatDate.Format("2006-01-02")

	title := fmt.Sprintf(`[%d days ago] Carpe QRSS for <big><b>%s</b></big> (at %s)`, daysAgo, thatDateIsoDay, time.Now().UTC().Format(time.UnixDate))

	fmt.Fprintf(w, `<html>
<head>
  <META NAME="ROBOTS" CONTENT="INDEX, NOFOLLOW">
  <title>%s</title>
</head>
<body>

<h3>%s</h3>
<p>
  <tt><b><big>
  LINKS TO DAYS AGO: &nbsp;
`, title, title)

	for i := 0; i < 8; i++ {
		fmt.Fprintf(w, ` <a href="index%d.html">[%d]</a> &nbsp; `, i, i)
	}

	fmt.Fprintf(w, `
  </big></b></tt>
<br>
<p>
  For each source, data is grouped by Zulu day.
  For each source on each zulu day, there are two images, composed of frames
  which are images seized from qrss grabbers.
  The image on the left is the average of all the frames;
  the one on the right is an animated GIF made up of all the frames.
  A caption in green tells what the source was, prehaps the band,
  and roughly what time the image was made or seized.
<p>
  <b>All times and dates are Zulu.</b>
<p>
  Source is at <a href="https://github.com/strickyak/carpe-qrss">https://github.com/strickyak/carpe-qrss</a>.
  Site hosted in Digital Ocean.
<p>
  -- 73 de %s
<p>
<br>
<br>
`, WHOM)
	tags := o.SortedWebTagsForDay(daysAgo)
	for _, tag := range tags {
		fmt.Fprintf(w, "[<a href='#%s.%d'>%s</a>] &nbsp;\n", tag, daysAgo, tag)
	}
	fmt.Fprintf(w, "<p><br><br>\n")

	fmt.Fprintf(w, `
<table cellpadding=5 border=1>
  <tr>
    <th>Days<br>Ago</th>
    <th align=center>RGB-wise Average of Frames</th>
    <th align=center>Animated GIF of Frames</th>
  </tr>
`)

	for _, tag := range tags {
		shortTag := strings.Split(tag, "~")[0]
		fmt.Fprintf(w, `
  <tr>
    <th>Days<br>Ago</th>
    <th colspan=2 align=center><a name="%s.%d"><tt> <big><big><big>%s &nbsp; &nbsp; </big></big></big> <a href="%s.d/">%q</a> </tt></a></th>
  </tr>
`, tag, daysAgo, shortTag, tag, tag)

		v1 := o.TagDayHash[tag]
		//n := 0
		//for i := 0; i < 30; i++ {
		p, ok := v1.NewProducts[daysAgo]
		if !ok {
			continue
		}
		movie := p.MovieName
		mean := p.MeanName
		fmt.Fprintf(w, `
  <tr>
    <th align=center ><big>%d</big></th>
    <td><img src="%s/%s"></td>
    <td><img src="%s/%s"></td>
  <tr>
`, daysAgo, P.Base(P.Dir(mean)), P.Base(mean), P.Base(P.Dir(movie)), P.Base(movie))
		//n++
		//if n > 2 {
		//break
		//}
		//}
	}
	fmt.Fprintf(w, `
  <hr> <hr> <hr>
</body></html>
`)
}
