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
	"time"
)

var seededRand = rand.New(rand.NewSource(1))

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func wrap(x, b int) int {
	if x < 0 {
		return x + b
	}
	if x >= b {
		return x - b
	}
	return x
}

func offset(m float64) int {
	sample := seededRand.NormFloat64() * m
	return int(sample)
}

func brighten(r uint8, add uint8) uint8 {
	// return r*4/6 + 20000
	r32 := uint32(r)
	add32 := uint32(add)
	return uint8(r32 - r32*add32/255 + add32)
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

	seedPtr := flag.Int64("seed", -1, "random seed. set to -1 if you want to generate it from time. the old version has used seed=1")

	flag.Parse()

	if *seedPtr == -1 {
		seededRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	} else if *seedPtr != 1 {
		seededRand = rand.New(rand.NewSource(*seedPtr))
	}

	reader, err := os.Open(flag.Args()[0])
	check(err)
	m, err := png.Decode(reader)
	m_raw_stride, m_raw_pix := 0, []uint8(nil)

	switch m.(type) {
	case *image.NRGBA:
		m_raw := m.(*image.NRGBA)
		m_raw_stride = m_raw.Stride
		m_raw_pix = m_raw.Pix
	case *image.RGBA:
		m_raw := m.(*image.RGBA)
		m_raw_stride = m_raw.Stride
		m_raw_pix = m_raw.Pix
	}
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
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			if seededRand.Intn(BHEIGHT*b.Max.X) == 0 {
				line_off = offset(BOFFSET)
				stride = seededRand.NormFloat64() * STRIDE_MAG
				yset = y
			}
			stride_off := int(stride * float64(y-yset))
			offx := offset(MAG) + line_off + stride_off
			offy := offset(MAG)
			src_idx := m_raw_stride*wrap(y+offy, b.Max.Y) + 4*wrap(x+offx, b.Max.X)
			dst_idx := new_img.Stride*y + 4*x

			copy(new_img.Pix[dst_idx:dst_idx+4], m_raw_pix[src_idx:src_idx+4])
		}
	}

	// second stage is adding per-channel scan inconsistency and brightening
	new_img1 := image.NewRGBA(b)

	lr, lg, lb := *lrPtr, *lgPtr, *lbPtr
	// lr, lg, lb := 0., 0., 0.
	LAG := *lagPtr
	ADD := uint8(*addPtr)
	STD_OFFSET := *stdOffsetPtr
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			lr += seededRand.NormFloat64() * LAG
			lg += seededRand.NormFloat64() * LAG
			lb += seededRand.NormFloat64() * LAG
			offx := offset(STD_OFFSET)

			ra_idx := new_img.Stride*y + 4*wrap(x+int(lr)-offx, b.Max.X)
			g_idx := new_img.Stride*y + 4*wrap(x+int(lg), b.Max.X)
			b_idx := new_img.Stride*y + 4*wrap(x+int(lg)+offx, b.Max.X)

			r := new_img.Pix[ra_idx]
			a := new_img.Pix[ra_idx+3]
			g := new_img.Pix[g_idx+1]
			b := new_img.Pix[b_idx+2]

			r, g, b = brighten(r, ADD), brighten(g, ADD), brighten(b, ADD)
			dst_idx := new_img1.Stride*y + 4*x

			copy(new_img1.Pix[dst_idx:dst_idx+4], []uint8{r, g, b, a})
		}
	}

	// third stage is to add chromatic abberation+chromatic trails
	// (trails happen because we're changing the same image we process)
	MEAN_ABBER := *meanAbberPtr
	STD_ABBER := *stdAbberPtr
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			// offx := 10 + offset(40)
			offx := MEAN_ABBER + offset(STD_ABBER) // lower offset arg = longer trails

			ra_idx := new_img1.Stride*y + 4*wrap(x+offx, b.Max.X)
			g_idx := new_img1.Stride*y + 4*wrap(x, b.Max.X)
			b_idx := new_img1.Stride*y + 4*wrap(x-offx, b.Max.X)

			r := new_img1.Pix[ra_idx]
			a := new_img1.Pix[ra_idx+3]
			g := new_img1.Pix[g_idx+1]
			b := new_img1.Pix[b_idx+2]

			dst_idx := new_img1.Stride*y + 4*x
			copy(new_img1.Pix[dst_idx:dst_idx+4], []uint8{r, g, b, a})
		}
	}

	writer, err := os.Create(flag.Args()[1])
	check(err)
	png.Encode(writer, new_img1)
	writer.Close()
}
