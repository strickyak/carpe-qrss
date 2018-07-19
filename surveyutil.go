// +build main

package main

import (
	"bufio"
	"flag"
	"github.com/strickyak/carpe-qrss"
	"log"
	"os"
)

var SPOOL = flag.String("spool", "spool/", "spool dir prefix")

func main() {
	flag.Parse()
	s := carpe.NewSurvey(*SPOOL)
	s.UsedOther[*SPOOL+"/index.html"] = true
	s.UsedOther[*SPOOL+"index.html"] = true
	s.Walk()

	s.BuildMovies("tmp")
	s.CollectGarbage()
	// s.DumpProducts(os.Stderr)
	fd, err := os.Create(*SPOOL + "/index.html")
	carpe.DieIf(err, "os.Create", *SPOOL+"/index.html")
	w := bufio.NewWriter(fd)

	s.WriteWebPage(w)
	//for i := 0; i < 3; i++ {
	// s.WriteWebPageForDay(w, i)
	////}
	//s.WriteWebPageForDay(w, 0)
	//s.WriteWebPageForDay(w, 1)

	w.Flush()
	fd.Close()
	log.Printf("surveyutil DONE.")
}
