package locateimage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	red1 = color.RGBA{255, 0, 0, 255}
	red2 = color.RGBA{250, 0, 0, 255}
	red3 = color.RGBA{255, 5, 5, 255}

	s1 = sample(red1)
	s2 = sample(red2)
	s3 = sample(red3)
	ss = []image.Image{s1, s2, s3}

	s_1 = 1
	s_2 = 2
	s_3 = 3

	r1 = image.Rect(5, 5, 15, 15)
	r2 = image.Rect(10, 10, 20, 20)
	r3 = image.Rect(50, 5, 60, 15)
	r4 = image.Rect(90, 90, 100, 100)

	cnv = canvas()

	ctx = context.Background()
)

func canvas() *image.RGBA {
	cnv := image.NewRGBA(image.Rect(0, 0, 100, 100))
	draw.Draw(cnv, cnv.Bounds(), image.NewUniform(color.White), image.ZP, draw.Src)
	draw.Draw(cnv, r1, image.NewUniform(red3), image.ZP, draw.Src)
	draw.Draw(cnv, r2, image.NewUniform(red1), image.ZP, draw.Src)
	draw.Draw(cnv, r3, image.NewUniform(red2), image.ZP, draw.Src)
	draw.Draw(cnv, r4, image.NewUniform(red2), image.ZP, draw.Src)
	saveToFile(cnv, "locateimage_TestFind.png")
	return cnv
}
func TestAll(t *testing.T) {
	tests := []struct {
		sample    int
		tolerance float64
		err       error
		mm        []Match
	}{
		{s_1, 0.0, nil, []Match{Match{r2, 1.0}}},
		{s_2, 0.0, nil, []Match{Match{r3, 1.0}, Match{r4, 1.0}}},
		{s_3, 0.0, nil, []Match{}},

		{s_1, 0.001, nil, []Match{Match{r2, 1.0}}},
		{s_1, 0.04, nil, []Match{Match{r1, 0.999882}, Match{r2, 1.0}, Match{r3, 0.999921}, Match{r4, 0.999921}}},
		{s_3, 0.001, nil, []Match{}},
		{s_3, 0.007, nil, []Match{Match{r1, 0.999961}, Match{r2, 0.999842}}},
	}

	for _, test := range tests {
		Sort(test.mm)
		mm, err := All(ctx, cnv, ss[test.sample-1], test.tolerance)
		if a, e := str2(mm, err), str2(test.mm, test.err); a != e {
			t.Errorf("** All(s%d, %.0f%%) = %v, wanted %v", test.sample, 100*test.tolerance, a, e)
		} else {
			t.Logf("✓ All(s%d, %.0f%%) = %v", test.sample, 100*test.tolerance, a)
		}
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		sample    int
		tolerance float64
		selection Selection
		err       error
		m         Match
	}{
		{1, 0.0, Only, nil, Match{r2, 1.0}},
	}

	for _, test := range tests {
		m, err := Find(ctx, cnv, ss[test.sample-1], test.tolerance, test.selection)
		if a, e := str1(m, err), str1(test.m, test.err); a != e {
			t.Errorf("** Find(s%d, %.0f%%, %v) = %v, wanted %v", test.sample, 100*test.tolerance, test.selection, a, e)
		} else {
			t.Logf("✓ Find(s%d, %.0f%%, %v) = %v", test.sample, 100*test.tolerance, test.selection, a)
		}
	}
}

func sample(c color.Color) *image.RGBA {
	s := image.NewRGBA(image.Rect(0, 0, 10, 10))
	draw.Draw(s, s.Bounds(), image.NewUniform(c), image.ZP, draw.Src)
	return s
}

func str1(m Match, err error) string {
	if err != nil {
		if m.Similarity > 0 {
			return fmt.Sprintf("%v <%v>", m, err.Error())
		} else {
			return fmt.Sprintf("<%v>", err.Error())
		}
	} else {
		return m.String()
	}
}

func str2(mm []Match, err error) string {
	if err != nil {
		if len(mm) > 0 {
			return fmt.Sprintf("%v <%v>", strmm(mm), err.Error())
		} else {
			return fmt.Sprintf("<%v>", err.Error())
		}
	} else {
		if len(mm) > 0 {
			return strmm(mm)
		} else {
			return "none"
		}
	}
}

func strmm(mm []Match) string {
	var ss []string
	for _, m := range mm {
		ss = append(ss, m.String())
	}
	return strings.Join(ss, "; ")
}

func saveToFile(img image.Image, name string) {
	fn := filepath.Join(os.TempDir(), name)
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		panic(err)
	}
	err = ioutil.WriteFile(fn, buf.Bytes(), 0644)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stderr, "saved: %s\n", fn)
}
