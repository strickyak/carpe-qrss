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
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

type ImageConverter func(img image.Image, filename string) image.Image

func BuildAnimatedGif(filenames []string, delay time.Duration, converter ImageConverter, outgif string, outmean string) (failures int) {
	var ms []*image.Paletted
	var sums []int64
	var r image.Rectangle
	var xlen, ylen int
	for i, name := range filenames {
		runtime.Gosched()
		log.Printf("Reading %v [%d/%d]\n", name, i+1, len(filenames))
		m, err := readImage(name)
		if err != nil {
			log.Printf("error reading image: %q: %v", name, err)
			// Remove it, so it won't bother us and mess up the digest again.
			log.Printf("Removing bad image: %q", name)
			os.Remove(name)
			failures++
			continue
		}
		if converter != nil {
			m = converter(m, name)
		}
		r = m.Bounds()
		pm := image.NewPaletted(r, palette.Plan9)
		runtime.Gosched()
		draw.FloydSteinberg.Draw(pm, r, m, image.ZP)
		runtime.Gosched()
		ms = append(ms, pm)

		xlen = r.Max.X - r.Min.X
		ylen = r.Max.Y - r.Min.Y

		if outmean != "" {
			if sums == nil {
				sums = make([]int64, xlen*ylen*3)
			}
			for x := 0; x < xlen; x++ {
				for y := 0; y < ylen; y++ {
					cr, cg, cb, _ := pm.At(x+r.Min.X, y+r.Min.Y).RGBA()
					sums[x*ylen*3+y*3+0] += int64(cr)
					sums[x*ylen*3+y*3+1] += int64(cg)
					sums[x*ylen*3+y*3+2] += int64(cb)
				}
			}
		}
	}

	if len(ms) == 0 {
		return
	}

	if outmean != "" {
		// Write the mean.
		pm := image.NewPaletted(r, palette.Plan9)
		for x := 0; x < xlen; x++ {
			for y := 0; y < ylen; y++ {
				// cr, cg, cb, ca := pm.At(x + r.Min.x, y + r.Min.y).Color()

				cr := sums[x*ylen*3+y*3+0] / int64(len(ms))
				cg := sums[x*ylen*3+y*3+1] / int64(len(ms))
				cb := sums[x*ylen*3+y*3+2] / int64(len(ms))
				pm.Set(x+r.Min.X, y+r.Min.Y, color.NRGBA64{uint16(cr), uint16(cg), uint16(cb), 0xFFFF})
			}
		}

		os.MkdirAll(filepath.Dir(outmean), 0755)
		fd, _ := os.Create(outmean)
		runtime.Gosched()
		png.Encode(fd, pm)
		runtime.Gosched()
		fd.Close()
	}

	ds := make([]int, len(ms))
	for i := range ds {
		ds[i] = int(100 * delay.Seconds()) // Hundredths of a second.
	}
	log.Println("Generating", outgif)
	os.MkdirAll(filepath.Dir(outgif), 0755)
	fd, err := os.Create(outgif)
	if err != nil {
		debug.PrintStack()
		log.Panicf("error creating %v: %v", outgif, err)
	}
	defer fd.Close()
	runtime.Gosched()
	err = gif.EncodeAll(fd, &gif.GIF{Image: ms, Delay: ds, LoopCount: -1})
	runtime.Gosched()
	if err != nil {
		debug.PrintStack()
		log.Panicf("error writing %v: %v", outgif, err)
	}
	err = fd.Close()
	if err != nil {
		debug.PrintStack()
		log.Panicf("error closing %v: %v", outgif, err)
	}
	log.Printf("Valid inputs: %d; Error inputs: %d; output %q", len(ms), failures, outgif)
	return
}

func readImage(name string) (image.Image, error) {
	fd, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	m, _, err := image.Decode(fd)
	return m, err
}
