// +build main

// Thanks to https://github.com/nf/giffy
// Modified 2018 by Henry Strickland (github.com/strickyak) to use flag and command line arguments, according to the following license.

/*
Copyright 2013 Google Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// This command giffy reads all the JPEG and PNG files from the command line arguments
// and writes them to an animated GIF as the file named by the -o flag.
package main

import (
	"flag"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"os"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

var OUT = flag.String("o", "/dev/stdout", "The output gif filename")
var DELAY = flag.Duration("delay", 100*time.Millisecond, "delay per frame")

func main() {
	flag.Parse()
	var ms []*image.Paletted
	filenames := flag.Args()
	for i, n := range filenames {
		log.Printf("Reading %v [%d/%d]\n", n, i+1, len(filenames))
		m, err := readImage(n)
		if err != nil {
			log.Fatalf("error reading image: %v: %v", n, err)
		}
		r := m.Bounds()
		pm := image.NewPaletted(r, palette.Plan9)
		draw.FloydSteinberg.Draw(pm, r, m, image.ZP)
		ms = append(ms, pm)
	}
	ds := make([]int, len(ms))
	for i := range ds {
		ds[i] = int(100 * DELAY.Seconds()) // Hundredths of a second.
	}
	log.Println("Generating", *OUT)
	f, err := os.Create(*OUT)
	if err != nil {
		log.Fatalf("error creating %v: %v", *OUT, err)
	}
	err = gif.EncodeAll(f, &gif.GIF{Image: ms, Delay: ds, LoopCount: -1})
	if err != nil {
		log.Fatalf("error writing %v: %v", *OUT, err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("error closing %v: %v", *OUT, err)
	}
}

func readImage(name string) (image.Image, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	m, _, err := image.Decode(f)
	return m, err
}
