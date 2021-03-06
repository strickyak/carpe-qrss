package carpe

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const TS_FORMAT = "2006-01-02-150405"

// Map URL to Date String
var LastModified = make(map[string]string)
var ETag = make(map[string]string)

func FetchEveryNSeconds(n uint, grabbers_url string, spool string) {
	targets := GetTargetsViaURL(grabbers_url)
	for _, t := range targets {
		localT := t
		log.Printf("new: Nick %q url %q", t.Nick, t.URL)
		EveryNSeconds(t.Nick, n, func() { Get(localT, spool) })
	}
	/*
		for _, t := range Targets {
			localT := t
			log.Printf("old: Nick %q url %q", t.Nick, t.URL)
			EveryNSeconds(t.Nick, n, func() { Get(localT, spool) })
		}
	*/
}

func Fetch(grabbers_url string, spool string) {
	runtime.Gosched()
	targets := GetTargetsViaURL(grabbers_url)
	for _, t := range targets {
		log.Println("GET", t.Nick, t.URL)
		filename, status, err := Get(t, spool)
		log.Println("...", status, err, filename)
	}
	runtime.Gosched()
}

func Get(t Target, spool string) (filename string, status int, err error) {
	runtime.Gosched()
	req, err := http.NewRequest("GET", t.URL, nil /*empty -- body io.Reader*/)
	req.Header.Add("User-Agent", "github.com/strickyak/carpe-qrss")

	last, _ := LastModified[t.URL]
	if last != "" {
		req.Header.Add("If-Modified-Since", last)
	} else {
		etag, _ := ETag[t.URL]
		if etag != "" {
			req.Header.Add("If-None-Match", etag)
		}
	}

	c := &http.Client{
		Timeout: 20 * time.Second,
	}

	runtime.Gosched()
	log.Printf("getting %q", t.URL)
	resp, err := c.Do(req)
	runtime.Gosched()
	if err != nil {
		return "", 418, err
	}

	if resp.StatusCode == 200 {
		var buf bytes.Buffer
		_, err = io.Copy(&buf, resp.Body)
		if err != nil {
			return "", 418, err
		}
		body := buf.Bytes()

		ts := time.Now()
		lm := resp.Header.Get("Last-Modified")
		LastModified[t.URL] = lm
		etag := resp.Header.Get("ETag")
		ETag[t.URL] = etag
		t1, e1 := time.Parse(time.RFC1123, lm)
		t2, e2 := time.Parse(time.RFC1123Z, lm)
		if e1 == nil {
			ts = t1
			log.Println("Using t1", t1.String())
		} else if e2 == nil {
			ts = t2
			log.Println("Using t2", t2.String())
		}

		tmpdir := fmt.Sprintf("%s/tmp.d", spool)
		err = os.MkdirAll(tmpdir, 0755)
		if err != nil {
			log.Fatalf("MkdirAll %q failed: %v", tmpdir, err)
		}

		timeString := ts.UTC().Format(TS_FORMAT)
		filename := fmt.Sprintf("%s/%s.0x0.%s.jpg", tmpdir, t.Nick, timeString)
		ioutil.WriteFile(filename, body, 0777)

		newname, _ := RenameFileForImageSize(spool, filename)
		if newname != "" {
			filename = newname
		}

		log.Printf("got %q", filename)
		return filename, resp.StatusCode, nil
	} else {
		return "", resp.StatusCode, nil
	}
}

func EveryNSeconds(key string, n uint, fn func()) {
	hasher := fnv.New32()
	hasher.Write([]byte(key))
	hash := hasher.Sum32()
	offset := uint(hash) % n
	log.Printf("Offset is %d for %q (hash %x)", offset, key, hash)

	go func() {
		for {
			now := uint(time.Now().Unix()) % n
			wait := (offset - now) % n
			if wait == 0 {
				wait = n
			}

			timer := time.NewTimer(time.Duration(wait) * time.Second)
			<-timer.C
			timer.Stop()

			log.Printf("EveryNSeconds: Running %q offset %d at %v", key, offset, time.Now())
			fn()
		}
	}()
}
