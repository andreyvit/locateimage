package locateimage

import (
	"image"
	"image/draw"
)

/*
Convert converts any image to an image format supported by this package.
If the image is already in a supported format, it is returned as is.

Currently Convert always converts to image.RGBA.
*/
func Convert(src image.Image) image.Image {
	return convertToRGBA(src)
}

/*
convertToRGBA converts any image to image.RGBA, currently the only image type
supported by this package. If the image is already RGBA, it is returned as is.
*/
func convertToRGBA(src image.Image) *image.RGBA {
	if rgba, ok := src.(*image.RGBA); ok {
		return rgba
	}

	b := src.Bounds()
	rgba := image.NewRGBA(b)
	draw.Draw(rgba, b, src, b.Min, draw.Src)
	return rgba
}
