package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math/rand"
	"os"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func wrap(x, a, b int) int {
	if x < a {
		return x + b - a
	}
	if x > b {
		return x + a - b
	}
	return x
}

func offset(m float64) int {
	sample := rand.NormFloat64() * m
	return int(sample)
}

func brighten(r uint32, add uint32) uint32 {
	// return r*4/6 + 20000
	return r - r*add/65535 + add
}

func uint32_to_rgba(r, g, b, a uint32) color.RGBA {
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [options] [input] [output] \n", os.Args[0])
		flag.PrintDefaults()
	}
	magPtr := flag.Float64("mag", 7.0, "dissolve blur strength")
	blockHeightPtr := flag.Int("bheight", 10, "average distorted block height")
	blockOffsetPtr := flag.Float64("boffset", 30., "distorted block offset strength")
	strideMagPtr := flag.Float64("stride", 0.1, "distorted block stride strength")

	lagPtr := flag.Float64("lag", 0.005, "per-channel scanline lag strength")
	lrPtr := flag.Float64("lr", -7, "initial red scanline lag")
	lgPtr := flag.Float64("lg", 0, "initial green scanline lag")
	lbPtr := flag.Float64("lb", 3, "initial blue scanline lag")
	stdOffsetPtr := flag.Float64("stdoffset", 10, "std. dev. of red-blue channel offset (non-destructive)")
	addPtr := flag.Int("add", 39, "additional brightness control (0-255)")

	meanAbberPtr := flag.Int("meanabber", 10, "mean chromatic abberation offset")
	stdAbberPtr := flag.Float64("stdabber", 10, "std. dev. of chromatic abberation offset (lower values induce longer trails)")

	flag.Parse()

	reader, err := os.Open(flag.Args()[0])
	check(err)
	m, err := png.Decode(reader)
	check(err)
	reader.Close()

	b := m.Bounds()

	// first stage is dissolve+block corruption
	new_img := image.NewRGBA(b)
	line_off := 0
	stride := 0.
	yset := 0
	// const MAG = 2.5
	MAG := *magPtr
	// const MAG = 0
	// const MAG = 3
	BHEIGHT := *blockHeightPtr
	BOFFSET := *blockOffsetPtr
	STRIDE_MAG := *strideMagPtr
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if rand.Intn(BHEIGHT*(b.Max.X-b.Min.X)) == 0 {
				line_off = offset(BOFFSET)
				stride = rand.NormFloat64() * STRIDE_MAG
				yset = y
			}
			stride_off := int(stride * float64(y-yset))
			offx := offset(MAG) + line_off + stride_off
			offy := offset(MAG)
			src := m.At(
				wrap(x+offx, b.Min.X, b.Max.X),
				wrap(y+offy, b.Min.Y, b.Max.Y))
			new_img.Set(x, y, src)
		}
	}

	// second stage is adding per-channel scan inconsistency and brightening
	new_img1 := image.NewRGBA(b)

	lr, lg, lb := *lrPtr, *lgPtr, *lbPtr
	// lr, lg, lb := 0., 0., 0.
	LAG := *lagPtr
	ADD := uint32(*addPtr) << 8
	STD_OFFSET := *stdOffsetPtr
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			lr += rand.NormFloat64() * LAG
			lg += rand.NormFloat64() * LAG
			lb += rand.NormFloat64() * LAG
			offx := offset(STD_OFFSET)

			r, _, _, a := new_img.At(
				wrap(x+int(lr)-offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, g, _, _ := new_img.At(
				wrap(x+int(lg), b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, _, b, _ := new_img.At(
				wrap(x+int(lb)+offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()

			r, g, b = brighten(r, ADD), brighten(g, ADD), brighten(b, ADD)
			new_img1.Set(x, y, uint32_to_rgba(r, g, b, a))
		}
	}

	// third stage is to add chromatic abberation+chromatic trails
	// (trails happen because we're changing the same image we process)
	MEAN_ABBER := *meanAbberPtr
	STD_ABBER := *stdAbberPtr
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// offx := 10 + offset(40)
			offx := MEAN_ABBER + offset(STD_ABBER) // lower offset arg = longer trails
			r, _, _, a := new_img1.At(
				wrap(x+offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, g, _, _ := new_img1.At(
				wrap(x, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, _, b, _ := new_img1.At(
				wrap(x-offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			new_img1.Set(x, y, uint32_to_rgba(r, g, b, a))
		}
	}
	// for y := b.Min.Y; y < b.Max.Y; y++ {
	// 	for x := b.Min.X; x < b.Max.X; x++ {
	// 		offx := 10
	// 		r, _, _, a := new_img.At(
	// 			wrap(x+offx, b.Min.X, b.Max.X),
	// 			wrap(y, b.Min.Y, b.Max.Y)).RGBA()
	// 		_, g, _, _ := new_img.At(
	// 			wrap(x, b.Min.X, b.Max.X),
	// 			wrap(y, b.Min.Y, b.Max.Y)).RGBA()
	// 		_, _, b, _ := new_img.At(
	// 			wrap(x-offx, b.Min.X, b.Max.X),
	// 			wrap(y, b.Min.Y, b.Max.Y)).RGBA()
	// 		new_img.Set(x, y, uint32_to_rgba(r, g, b, a))
	// 	}
	// }

	writer, err := os.Create(flag.Args()[1])
	check(err)
	png.Encode(writer, new_img1)
	writer.Close()
}
