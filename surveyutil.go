// +build main

package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/strickyak/carpe-qrss"
	"log"
	"os"
	"path/filepath"
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

	fd, err := os.Create(*SPOOL + "/index.html")
	carpe.DieIf(err, "os.Create", *SPOOL+"/index.html")
	w := bufio.NewWriter(fd)
	// s.WriteWebPage(w)
	redirect := `<!DOCTYPE html>
	<html><head>
	<meta http-equiv="refresh"
            content="0; url=index0.html">
	</head></html>`
	fmt.Fprintln(w, redirect)
	w.Flush()
	fd.Close()

	for i := 0; i < 8; i++ {
		filename := fmt.Sprintf("%s/index%d.html", *SPOOL, i)
		s.UsedOther[filename] = true
		s.UsedOther[filepath.Clean(filename)] = true

		fd, err := os.Create(filename)
		carpe.DieIf(err, "os.Create", filename)
		w := bufio.NewWriter(fd)
		s.WriteWebPageForDay(w, i)
		w.Flush()
		fd.Close()
	}

	log.Printf("surveyutil DONE.")
}
