package carpe

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Map URL to Date String
var LastModified = make(map[string]string)

func Fetch(spool string) {
	for _, t := range Targets {
		println("GET", t.Nick, t.Band, t.URL)
		filename, status, err := Get(t, spool)
		println("...", filename, status, err)
	}
}

func Get(t Target, spool string) (filename string, status int, err error) {
	req, err := http.NewRequest("GET", t.URL, nil /*empty -- body io.Reader*/)
	req.Header.Add("User-Agent", "github.com/strickyak/carpe-qrss")

	last, _ := LastModified[t.URL]
	if last != "" {
		req.Header.Add("If-Modified-Since", last)
	}

	c := &http.Client{
		Timeout: 20 * time.Second,
	}

	resp, err := c.Do(req)
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

		unix := time.Now().Unix()
		lm := resp.Header.Get("Last-Modified")
		LastModified[t.URL] = lm
		t1, e1 := time.Parse(time.RFC1123, lm)
		t2, e2 := time.Parse(time.RFC1123Z, lm)
		if e1 == nil {
			unix = t1.Unix()
			println("Using t1", t1.String(), unix)
		} else if e2 == nil {
			unix = t2.Unix()
			println("Using t2", t2.String(), unix)
		}

		filename := fmt.Sprintf("%s%s.%d.%010d.pic", spool, t.Nick, t.Band, unix)
		ioutil.WriteFile(filename, body, 0777)

		return filename, resp.StatusCode, nil
	} else {
		return "", resp.StatusCode, nil
	}
}
