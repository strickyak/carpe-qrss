// +build main

/*
Command to crop an image.
Input image on stdin, Output JPEG image on stdout.
Four integer command line arguments:  left right top bottom.
They are how many pixels to trim from left, right, top, and bottom.
*/
package main

import (
	"flag"
	"image"
	"log"
	"strconv"

	carpe "github.com/strickyak/carpe-qrss"
)

// Luckily this works for jpeg.  Maybe for other Image Decodes?
type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func parseInt(a string) int {
	z, err := strconv.ParseInt(a, 10, 64)
	if err != nil {
		log.Fatalf("Not an int: %q: %v", a, err)
	}
	return int(z)
}

func main() {
	flag.Parse()

	// Four arguments are Edge Thicknesses to be cropped:
	left := parseInt(flag.Arg(0))
	right := parseInt(flag.Arg(1))
	top := parseInt(flag.Arg(2))
	bottom := parseInt(flag.Arg(3))
	margins := []int{left, right, top, bottom}

	pic := carpe.ReadImage("/dev/stdin")
	cropped := carpe.Crop(pic, margins)
	carpe.WriteJpegImage(cropped, "/dev/stdout")
}
