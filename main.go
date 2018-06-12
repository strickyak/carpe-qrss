// +build main

package main

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"time"

	carpe "github.com/strickyak/carpe-qrss"
)

var DELAY = flag.Duration("delay", 8*time.Minute, "delay between fetch rounds")
var SPOOL = flag.String("spool", "spool/", "prefix for created filenames")
var BIND = flag.String("bind", ":7899", "where to bind web server")

func main() {
	flag.Parse()

	robotsFilename := filepath.Join(*SPOOL, "robots.txt")
	err := ioutil.WriteFile(robotsFilename, []byte(DEAR_ROBOTS), 0644)
	carpe.DieIf(err, "WriteFile", robotsFilename)

	carpe.StartWeb(*BIND, *SPOOL)

	for {
		carpe.Fetch(*SPOOL)
		println("Sleeping...", DELAY.String())
		time.Sleep(*DELAY)
		println("Awake.")
	}
}

const DEAR_ROBOTS = `# robots.txt
# Disallow (as best as we can) all except / and /index.html
# because the images are hardly indexable
# (and they vanish within days)
# and it's just a waste of bandwidth to allow them.

User-agent: *
Allow: /$
Allow: /index.html$
Disallow: /A
Disallow: /B
Disallow: /C
Disallow: /D
Disallow: /E
Disallow: /F
Disallow: /G
Disallow: /H
Disallow: /J
Disallow: /K
Disallow: /L
Disallow: /M
Disallow: /N
Disallow: /O
Disallow: /P
Disallow: /Q
Disallow: /R
Disallow: /S
Disallow: /T
Disallow: /U
Disallow: /V
Disallow: /W
Disallow: /X
Disallow: /Y
Disallow: /Z
Disallow: /a
Disallow: /b
Disallow: /c
Disallow: /d
Disallow: /e
Disallow: /f
Disallow: /g
Disallow: /h
Disallow: /j
Disallow: /k
Disallow: /l
Disallow: /m
Disallow: /n
Disallow: /o
Disallow: /p
Disallow: /q
Disallow: /r
Disallow: /s
Disallow: /t
Disallow: /u
Disallow: /v
Disallow: /w
Disallow: /x
Disallow: /y
Disallow: /z
`
