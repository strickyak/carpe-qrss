package carpe

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

import (
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

func BuildAnimatedGif(filenames []string, delay time.Duration, outfile string) {
	var ms []*image.Paletted
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
		ds[i] = int(100 * delay.Seconds()) // Hundredths of a second.
	}
	log.Println("Generating", outfile)
	f, err := os.Create(outfile)
	if err != nil {
		log.Fatalf("error creating %v: %v", outfile, err)
	}
	err = gif.EncodeAll(f, &gif.GIF{Image: ms, Delay: ds, LoopCount: -1})
	if err != nil {
		log.Fatalf("error writing %v: %v", outfile, err)
	}
	err = f.Close()
	if err != nil {
		log.Fatalf("error closing %v: %v", outfile, err)
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
