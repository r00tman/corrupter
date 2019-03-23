package main

import (
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

func main() {
	reader, err := os.Open(os.Args[1])
	check(err)
	m, err := png.Decode(reader)
	check(err)
	reader.Close()

	b := m.Bounds()

	new_img := image.NewRGBA(b)
	line_off := 0
	stride := 0.
	yset := 0
	// const MAG = 2.5
	const MAG = 7
	// const MAG = 0
	// const MAG = 3
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if rand.Intn(10*(b.Max.X-b.Min.X)) == 0 {
				line_off = offset(30)
				stride = rand.NormFloat64() * 0.1
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
	new_img1 := image.NewRGBA(b)

	lr, lg, lb := -7., 0., +3.
	// lr, lg, lb := 0., 0., 0.
	const LAG = 0.005
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			lr += rand.NormFloat64() * LAG
			lg += rand.NormFloat64() * LAG
			lb += rand.NormFloat64() * LAG
			offx := offset(10)

			r, _, _, a := new_img.At(
				wrap(x+int(lr)-offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, g, _, _ := new_img.At(
				wrap(x+int(lg), b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, _, b, _ := new_img.At(
				wrap(x+int(lb)+offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			// new_img1.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
			// new_img1.Set(x, y, color.RGBA{uint8((r*4/6 + 20000) >> 8), uint8((g*4/6 + 20000) >> 8), uint8((b*4/6 + 20000) >> 8), uint8(a >> 8)})
			new_img1.Set(x, y, color.RGBA{uint8((r*5/6 + 10000) >> 8), uint8((g*5/6 + 10000) >> 8), uint8((b*5/6 + 10000) >> 8), uint8(a >> 8)})
		}
	}
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			// offx := 10 + offset(40)
			offx := 10 + offset(10)
			r, _, _, a := new_img1.At(
				wrap(x+offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, g, _, _ := new_img1.At(
				wrap(x, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			_, _, b, _ := new_img1.At(
				wrap(x-offx, b.Min.X, b.Max.X),
				wrap(y, b.Min.Y, b.Max.Y)).RGBA()
			new_img1.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
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
	// 		new_img.Set(x, y, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)})
	// 	}
	// }

	writer, err := os.Create(os.Args[2])
	check(err)
	png.Encode(writer, new_img1)
	writer.Close()
}
