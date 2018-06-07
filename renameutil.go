// +build main

package main

import (
	"flag"
	"fmt"
	"github.com/strickyak/carpe-qrss"
	"os"
)

var SPOOL = flag.String("spool", "spool/", "spool dir prefix")

func main() {
	flag.Parse()
	for _, a := range flag.Args() {
		newname, _ := carpe.RenameFileForImageSize(*SPOOL, a)
		fmt.Fprintf(os.Stderr, "%q -> %q\n", a, newname)
	}
}
