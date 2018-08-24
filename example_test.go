package locateimage_test

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/andreyvit/locateimage"
)

func ExampleFind() {
	red := color.RGBA{255, 0, 0, 255}

	// 10x10 uniform-red sample
	sample := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(sample, sample.Bounds(), image.NewUniform(red), image.ZP, draw.Src)

	// expected match rect
	r := image.Rect(15, 15, 25, 25)

	// 100x100 white canvas with a red rectangle at (10,10)
	canvas := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(canvas, canvas.Bounds(), image.NewUniform(color.White), image.ZP, draw.Src)
	draw.Draw(canvas, r, image.NewUniform(red), image.ZP, draw.Src)

	fmt.Printf("Find: ")
	m, err := locateimage.Find(context.Background(), canvas, sample, 0, locateimage.Only)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v %.*f\n", m.Rect, locateimage.SimilarityDigits, m.Similarity)

	fmt.Printf("All: ")
	mm, err := locateimage.All(context.Background(), canvas, sample, 0.04)
	if err != nil {
		panic(err)
	}
	fmt.Println(mm)

	fmt.Printf("Foreach: ")
	err = locateimage.Foreach(context.Background(), canvas, sample, 0.04, func(m locateimage.Match) error {
		fmt.Println(m)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Output:
	// Find: (15,15)-(25,25) 1.000000
	// All: [(15,15)+(10x10) 100.0000%]
	// Foreach: (15,15)+(10x10) 100.0000%
}
