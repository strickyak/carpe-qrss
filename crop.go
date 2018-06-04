package carpe

import (
	"bufio"
	"image"
	"os"

	_ "image/gif"
	"image/jpeg"
	_ "image/png"

	"flag"
	"log"
)

const VerboseCrop = true
const BUFSIZ = 128 << 10 // For reading and writing images to/from files.

var QUALITY = flag.Int("quality", 90, "Jpeg output quality, typically 90 to 95.")

var JpegOpts = &jpeg.Options{Quality: *QUALITY}

// Luckily this works for jpeg.  Maybe for other Image Decodes?
type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func Crop(img image.Image, margins []int) image.Image {
	left, right, top, bottom := margins[0], margins[1], margins[2], margins[3]

	b1 := img.Bounds()
	if VerboseCrop {
		log.Printf("INPUT  bounds: (%d, %d) .. (%d, %d)", b1.Min.X, b1.Min.Y, b1.Max.X, b1.Max.Y)
	}

	cropR := image.Rectangle{
		Min: image.Point{b1.Min.X + left, b1.Min.Y + top},
		Max: image.Point{b1.Max.X - right, b1.Max.Y - bottom},
	}
	log.Printf("CROP   bounds: (%d, %d) .. (%d, %d)", cropR.Min.X, cropR.Min.Y, cropR.Max.X, cropR.Max.Y)

	cropped := img.(SubImager).SubImage(cropR)

	if VerboseCrop {
		b2 := cropped.Bounds()
		log.Printf("OUTPUT bounds: (%d, %d) .. (%d, %d)", b2.Min.X, b2.Min.Y, b2.Max.X, b2.Max.Y)
	}

	return cropped
}

func ReadImage(filename string) image.Image {
	fd, err := os.Open(filename)
	if err != nil {
		log.Panicf("ReadImage: Cannot open %q: %v", filename, err)
	}
	defer fd.Close()

	r := bufio.NewReaderSize(fd, BUFSIZ)
	img, _, err := image.Decode(r)
	if err != nil {
		log.Panicf("Error in image.Decode(stdin): %v", err)
	}
	return img
}

func WriteJpegImage(img image.Image, filename string) {
	fd, err := os.Create(filename)
	if err != nil {
		log.Panicf("WriteImage: Cannot create %q: %v", filename, err)
	}
	defer fd.Close()
	w := bufio.NewWriterSize(fd, BUFSIZ)
	err = jpeg.Encode(w, img, JpegOpts)
	w.Flush()
}
