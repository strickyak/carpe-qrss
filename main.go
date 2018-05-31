// +build main

package main

import (
	"flag"
	"time"

	carpe "github.com/strickyak/carpe-qrss"
)

var DELAY = flag.Duration("delay", 300*time.Second, "delay between fetch rounds")
var SPOOL = flag.String("spool", "/tmp/carpe.", "prefix for created filenames")

func main() {
	flag.Parse()
	for {
		carpe.Fetch(*SPOOL)
		println("Sleeping...", DELAY.String())
		time.Sleep(*DELAY)
		println("Awake.")
	}
}
