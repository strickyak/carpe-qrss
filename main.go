// +build main

package main

import (
	"flag"
	"time"

	carpe "github.com/strickyak/carpe-qrss"
)

var DELAY = flag.Duration("delay", 8*time.Minute, "delay between fetch rounds")
var SPOOL = flag.String("spool", "spool/", "prefix for created filenames")
var BIND = flag.String("bind", ":1919", "where to bind web server")

func main() {
	flag.Parse()

	carpe.StartWeb(*BIND, *SPOOL)

	for {
		carpe.Fetch(*SPOOL)
		println("Sleeping...", DELAY.String())
		time.Sleep(*DELAY)
		println("Awake.")
	}
}
