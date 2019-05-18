package main

import (
	"flag"
	"fmt"
	"image"
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

// force x to stay in [0, b) range. x is assumed to be in [-b,2*b) range
func wrap(x, b int) int {
	if x < 0 {
		return x + b
	}
	if x >= b {
		return x - b
	}
	return x
}

// get normally distributed (rounded to int) value with the specified std. dev.
func offset(stddev float64) int {
	sample := seededRand.NormFloat64() * stddev
	return int(sample)
}

// brighten the color safely, i.e., by simultaneously reducing contrast
func brighten(r uint8, add uint8) uint8 {
	r32, add32 := uint32(r), uint32(add)
	return uint8(r32 - r32*add32/255 + add32)
}

func main() {
	// command line parsing
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [input] [output]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "   or: %s [options] - (for stdin+stdout processing)\n", os.Args[0])
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
	addPtr := flag.Int("add", 37, "additional brightness control (0-255)")

	meanAbberPtr := flag.Int("meanabber", 10, "mean chromatic abberation offset")
	stdAbberPtr := flag.Float64("stdabber", 10, "std. dev. of chromatic abberation offset (lower values induce longer trails)")

	seedPtr := flag.Int64("seed", -1, "random seed. set to -1 if you want to generate it from time. the old version has used seed=1")

	flag.Parse()

	if *seedPtr == -1 {
		seededRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))
	} else if *seedPtr != 1 {
		seededRand = rand.New(rand.NewSource(*seedPtr))
	}

	// flag.Args() contain all non-option arguments, i.e., our input and output files
	reader := (*os.File)(nil)
	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(2)
	} else if flag.Args()[0] == "-" {
		// stdin/stdout processing
		reader = os.Stdin
	} else if len(flag.Args()) == 2 {
		err := error(nil)
		reader, err = os.Open(flag.Args()[0])
		check(err)
	} else {
		flag.Usage()
		os.Exit(2)
	}
	m, err := png.Decode(reader)
	check(err)
	reader.Close()

	// trying to obtain raw pointers to color data, since .At(), .Set() are very slow
	m_raw_stride, m_raw_pix := 0, []uint8(nil)

	switch m.(type) {
	default:
		log.Fatal("unknown image type")
	case *image.NRGBA:
		m_raw := m.(*image.NRGBA)
		m_raw_stride = m_raw.Stride
		m_raw_pix = m_raw.Pix
	case *image.RGBA:
		m_raw := m.(*image.RGBA)
		m_raw_stride = m_raw.Stride
		m_raw_pix = m_raw.Pix
	}

	b := m.Bounds()

	// first stage is dissolve+block corruption
	new_img := image.NewNRGBA(b)
	line_off := 0
	stride := 0.
	yset := 0
	MAG := *magPtr
	BHEIGHT := *blockHeightPtr
	BOFFSET := *blockOffsetPtr
	STRIDE_MAG := *strideMagPtr
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			// Every BHEIGHT lines in average a new distorted block begins
			if seededRand.Intn(BHEIGHT*b.Max.X) == 0 {
				line_off = offset(BOFFSET)
				stride = seededRand.NormFloat64() * STRIDE_MAG
				yset = y
			}
			// at the line where the block has begun, we don't want to offset the image
			// so stride_off is 0 on the block's line
			stride_off := int(stride * float64(y-yset))

			// offset is composed of the blur, block offset, and skew offset (stride)
			offx := offset(MAG) + line_off + stride_off
			offy := offset(MAG)

			// copy the corresponding pixel (4 bytes) to the new image
			src_idx := m_raw_stride*wrap(y+offy, b.Max.Y) + 4*wrap(x+offx, b.Max.X)
			dst_idx := new_img.Stride*y + 4*x

			copy(new_img.Pix[dst_idx:dst_idx+4], m_raw_pix[src_idx:src_idx+4])
		}
	}

	// second stage is adding per-channel scan inconsistency and brightening
	new_img1 := image.NewNRGBA(b)

	lr, lg, lb := *lrPtr, *lgPtr, *lbPtr
	LAG := *lagPtr
	ADD := uint8(*addPtr)
	STD_OFFSET := *stdOffsetPtr
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			lr += seededRand.NormFloat64() * LAG
			lg += seededRand.NormFloat64() * LAG
			lb += seededRand.NormFloat64() * LAG
			offx := offset(STD_OFFSET)

			// obtain source pixel base offsets. red/blue border is also smoothed by offx
			ra_idx := new_img.Stride*y + 4*wrap(x+int(lr)-offx, b.Max.X)
			g_idx := new_img.Stride*y + 4*wrap(x+int(lg), b.Max.X)
			b_idx := new_img.Stride*y + 4*wrap(x+int(lb)+offx, b.Max.X)

			// pixels are stored in (r, g, b, a) order in memory
			r := new_img.Pix[ra_idx]
			a := new_img.Pix[ra_idx+3]
			g := new_img.Pix[g_idx+1]
			b := new_img.Pix[b_idx+2]

			r, g, b = brighten(r, ADD), brighten(g, ADD), brighten(b, ADD)

			// copy the corresponding pixel (4 bytes) to the new image
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
			offx := MEAN_ABBER + offset(STD_ABBER) // lower offset arg = longer trails

			// obtain source pixel base offsets. only red and blue are distorted
			ra_idx := new_img1.Stride*y + 4*wrap(x+offx, b.Max.X)
			g_idx := new_img1.Stride*y + 4*x
			b_idx := new_img1.Stride*y + 4*wrap(x-offx, b.Max.X)

			// pixels are stored in (r, g, b, a) order in memory
			r := new_img1.Pix[ra_idx]
			a := new_img1.Pix[ra_idx+3]
			g := new_img1.Pix[g_idx+1]
			b := new_img1.Pix[b_idx+2]

			// copy the corresponding pixel (4 bytes) to the SAME image. this gets us nice colorful trails
			dst_idx := new_img1.Stride*y + 4*x
			copy(new_img1.Pix[dst_idx:dst_idx+4], []uint8{r, g, b, a})
		}
	}

	// write the image
	writer := (*os.File)(nil)
	if flag.Args()[0] == "-" {
		// stdin/stdout processing
		writer = os.Stdout
	} else {
		writer, err = os.Create(flag.Args()[1])
		check(err)
	}
	e := png.Encoder{CompressionLevel: png.NoCompression}
	e.Encode(writer, new_img1)
	writer.Close()
}
