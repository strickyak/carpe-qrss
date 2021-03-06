package carpe

// Thanks to https://github.com/nf/giffy
// Heavily modified 2018 by Henry Strickland (github.com/strickyak)
// http://www.apache.org/licenses/LICENSE-2.0

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
	"math"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

type ImageConverter func(img image.Image, filename string) image.Image

func sqrt64(a int64) int64 {
	return int64(math.Sqrt(float64(a)))
}

const NS = 5  // How many to stack

func BuildAnimatedGif(filenames []string, delay time.Duration, converter ImageConverter, outgif string, outmean string) (failures int) {
	n := (len(filenames) + NS - 1) / NS
	var stacks [][]int64 = make([][]int64, 200)  // hack 200
	var stacknames []string = make([]string, 200)  // hack 200
	var ms []*image.Paletted
	var sums []int64
	var r image.Rectangle
	var xlen, ylen int
	for i, name := range filenames {
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
		draw.FloydSteinberg.Draw(pm, r, m, image.ZP)
		ms = append(ms, pm)

		xlen = r.Max.X - r.Min.X
		ylen = r.Max.Y - r.Min.Y

		if outmean != "" {
			if sums == nil {
				sums = make([]int64, xlen*ylen*3)
			}
			j := i / NS
			if stacks[j] == nil {
				stacks[j] = make([]int64, xlen*ylen*3)
				stacknames[j]= filepath.Join(filepath.Dir(name), "stack." + filepath.Base(name) + ".png")
			}
			for x := 0; x < xlen; x++ {
				for y := 0; y < ylen; y++ {
					cr, cg, cb, _ := pm.At(x+r.Min.X, y+r.Min.Y).RGBA()
					sums[x*ylen*3+y*3+0] += int64(cr) * int64(cr)
					sums[x*ylen*3+y*3+1] += int64(cg) * int64(cg)
					sums[x*ylen*3+y*3+2] += int64(cb) * int64(cb)
					stacks[j][x*ylen*3+y*3+0] += int64(cr) * int64(cr)
					stacks[j][x*ylen*3+y*3+1] += int64(cg) * int64(cg)
					stacks[j][x*ylen*3+y*3+2] += int64(cb) * int64(cb)
				}
			}
		}
	}

	if len(ms) == 0 {
		return
	}

	if outmean != "" {
		{
			// Write the mean.
			pm := image.NewPaletted(r, palette.Plan9)
			for x := 0; x < xlen; x++ {
				for y := 0; y < ylen; y++ {
					cr := sqrt64(sums[x*ylen*3+y*3+0] / int64(len(ms)))
					cg := sqrt64(sums[x*ylen*3+y*3+1] / int64(len(ms)))
					cb := sqrt64(sums[x*ylen*3+y*3+2] / int64(len(ms)))
					pm.Set(x+r.Min.X, y+r.Min.Y, color.NRGBA64{uint16(cr), uint16(cg), uint16(cb), 0xFFFF})
				}
			}

			log.Println("Creating MEAN:", outmean)
			os.MkdirAll(filepath.Dir(outmean), 0755)
			fd, _ := os.Create(outmean)
			png.Encode(fd, pm)
			fd.Close()
		}

		// Write the stacks.
		for j := 0; j < n; j++ {
			if stacks[j] == nil {
				continue
			}
			pm := image.NewPaletted(r, palette.Plan9)
			for x := 0; x < xlen; x++ {
				for y := 0; y < ylen; y++ {
					cr := sqrt64(stacks[j][x*ylen*3+y*3+0] / int64(NS))
					cg := sqrt64(stacks[j][x*ylen*3+y*3+1] / int64(NS))
					cb := sqrt64(stacks[j][x*ylen*3+y*3+2] / int64(NS))
					pm.Set(x+r.Min.X, y+r.Min.Y, color.NRGBA64{uint16(cr), uint16(cg), uint16(cb), 0xFFFF})
				}
			}

			log.Println("Creating STACK:", stacknames[j])
			os.MkdirAll(filepath.Dir(outmean), 0755)
			fd, _ := os.Create(stacknames[j])
			png.Encode(fd, pm)
			fd.Close()
		}
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
	err = gif.EncodeAll(fd, &gif.GIF{Image: ms, Delay: ds, LoopCount: -1})
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
