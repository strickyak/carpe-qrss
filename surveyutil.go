// +build main

package main

import (
	"flag"
	"fmt"
	"github.com/strickyak/carpe-qrss"
)

var SPOOL = flag.String("spool", "spool/", "spool dir prefix")

func main() {
	flag.Parse()
	s := carpe.NewSurvey(*SPOOL)
	s.Walk()

	for k1, v1 := range s.TagDayHash {
		fmt.Printf("A %q\n", k1)
		for k2, v2 := range v1.DayHash {
			fmt.Printf("B %q %d\n", k1, k2)
			for k3, v3 := range v2.Surveys {
				fmt.Printf("C %q %d %d %#v\n", k1, k2, k3, v3)
			}

		}
	}
	s.BuildMovies("tmp")
}
