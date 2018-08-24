package locateimage

import (
	"context"
	"image"
)

func foreachRGBA(ctx context.Context, canvas, sample *image.RGBA, tolerance float64, f func(m Match) error) error {
	const pixSize = 4
	strideCanv, strideSamp := canvas.Stride, sample.Stride
	pixCanv, pixSamp := canvas.Pix, sample.Pix
	// ocanv, osamp := canvas.Rect.Min, sample.Rect.Min
	wcanv, hcanv, wsamp, hsamp := canvas.Rect.Dx(), canvas.Rect.Dy(), sample.Rect.Dx(), sample.Rect.Dy()

	if wcanv < wsamp || hcanv < hsamp || wsamp == 0 || hsamp == 0 {
		return nil
	}
	hcanv -= hsamp - 1
	wcanv -= wsamp - 1

	pixelTolerance := 2 * tolerance
	pixelDiffThreshold := int64(3*256*pixelTolerance + 0.5)
	maxTotalDiff := int64(wcanv) * int64(hcanv) * 3 * 255
	totalDiffThreshold := int64(tolerance*float64(maxTotalDiff) + 0.5)

	rowsampmax := strideSamp * hsamp
	for y := 0; y < hcanv; y++ {
		firstrowcanv := y * strideCanv

		if err := ctx.Err(); err != nil {
			return err
		}

		for x := 0; x < wcanv; x++ {

			rowcanv := firstrowcanv + x*pixSize
			matched := true
			diff := int64(0)
		matchLoop:
			for rowsamp := 0; rowsamp < rowsampmax; rowsamp += strideSamp {
				icanv := rowcanv
				isamp := rowsamp
				isampmax := rowsamp + pixSize*wsamp
				for isamp < isampmax {
					d0 := abs(int64(pixCanv[icanv+0]) - int64(pixSamp[isamp+0]))
					d1 := abs(int64(pixCanv[icanv+1]) - int64(pixSamp[isamp+1]))
					d2 := abs(int64(pixCanv[icanv+2]) - int64(pixSamp[isamp+2]))
					delta := d0 + d1 + d2
					if delta > pixelDiffThreshold {
						matched = false
						break matchLoop
					}
					diff += delta
					if diff > totalDiffThreshold {
						matched = false
						break matchLoop
					}

					icanv += pixSize
					isamp += pixSize
				}

				rowcanv += strideCanv
			}

			if matched {
				pt := canvas.Rect.Min.Add(image.Point{x, y})
				err := f(Match{
					Rect:       image.Rectangle{pt, pt.Add(image.Point{wsamp, hsamp})},
					Similarity: 1 - float64(diff)/float64(maxTotalDiff),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
