/*
Package locateimage performs an (exact or fuzzy) search of a sample image within
a larger image, returning the coordinates and similarity scores of the matches.

The package currently only deals with image.RGBA-encoded images. You can use
Convert to convert any image to this format. (Note that reading a PNG file
returns an NRGBA format, so the conversion will be required. It makes sense
to add support for formats like NRGBA in the future.)

The search is currently slow, taking tens or hundreds of milliseconds on
large images (screenshots of a 27" screen). This can likely be improved, and
contributions are welcome.
*/
package locateimage

import (
	"context"
	"errors"
	"fmt"
	"image"
	"math"
	"sort"
)

var (
	// ErrBreak is a convenient error you can use to break out of Foreach loop.
	// There is no special behavior associated with this error; it will be
	// returned from Foreach normally.
	ErrBreak = errors.New("done iterating")

	// ErrUnsupportedImageType is returned when trying to process an image outside
	// of the supported image types, which currently is only image.RGBA.
	ErrUnsupportedImageType = errors.New("unsupported image type")

	// ErrNotFound is returned by Find when no match has been found.
	ErrNotFound = errors.New("not found")

	// ErrMultipleFound is returned by Find in Only selection mode if more than
	// one match has been found.
	ErrMultipleFound = errors.New("multiple matches found")
)

const (
	// SimilarityDigits is the number of fractional digits to be taken into account
	// in similarity and tolerance values. Use this in %.*f when printing.
	SimilarityDigits = 6

	// SimilarityPrecision is 10^(-SimilarityDigits), the minimal difference in
	// the value of similarity or tolerance that should be considered significant.
	SimilarityPrecision float64 = 0.000001
)

// Match describes a single match of a sample image within a canvas.
type Match struct {
	// Rect is the part of the canvas that matched the sample. Its size is
	// equal to the size of the sample.
	Rect image.Rectangle

	// Similarity is a score from 0 (completely dissimilar) to 1 (exact match).
	// A returned match will have a similarity of at least (1 - tolerance).
	Similarity float64
}

// String returns a string description of this match for debugging purposes.
func (m Match) String() string {
	return fmt.Sprintf("(%d,%d)+(%dx%d) %0.*f%%", m.Rect.Min.X, m.Rect.Min.Y, m.Rect.Dx(), m.Rect.Dy(), SimilarityDigits-2, 100*m.Similarity)
}

// Before returns whether this match should sort before the given match.
//
// Higher-similarity matches come before lower-similarity matches, and
// equally-similar matches are ordered by Y, then by X, ascending.
func (a Match) Before(b Match) bool {
	if math.Abs(a.Similarity-b.Similarity) >= SimilarityPrecision {
		return a.Similarity > b.Similarity
	}
	if a.Rect.Min.Y != b.Rect.Min.Y {
		return a.Rect.Min.Y < b.Rect.Min.Y
	}
	if a.Rect.Min.X != b.Rect.Min.X {
		return a.Rect.Min.X < b.Rect.Min.X
	}
	return false
}

// Matches implements sort.Interface for []Match.
type matches []Match

func (mm matches) Len() int           { return len(mm) }
func (mm matches) Swap(i, j int)      { mm[i], mm[j] = mm[j], mm[i] }
func (mm matches) Less(i, j int) bool { return mm[i].Before(mm[j]) }

// Sort sorts the list of matches in the order established by Match.Before.
//
// Higher-similarity matches come before lower-similarity matches, and
// equally-similar matches are ordered by Y, then by X, ascending.
func Sort(mm []Match) {
	sort.Sort(matches(mm))
}

// Selection determines the way Find selects a match to return.
type Selection int

const (
	// Fastest asks for whatever match is encountered first, in undefined order.
	Fastest = Selection(iota)

	// Best asks for the match with the best similarity score.
	Best

	// Only asks for the best match like Best mode, but asks to verify that
	// only a single match exists. If multiple matches are found,
	// ErrMultipleFound will be returned together with the best match.
	Only
)

// String returns a string description of this Selection value for debugging purposes.
func (s Selection) String() string {
	switch s {
	case Fastest:
		return "Fastest"
	case Best:
		return "Best"
	case Only:
		return "Only"
	default:
		panic("unknown Selection")
	}
}

/*
Find locates and returns a single match of the sample image within a canvas
image. Depending on the selection argument, this will either be the best match
(which requires searching the entire canvas) or the first encountered match
(which will stop searching after the match is found).

Tolerance is a value between 0 and 1 specifying how much difference is
tolerated between the sample and its match. Pass 0 to find exact matches only.
A reasonable value for fuzzy matching is around 0.05.

Searching stops if the context is canceled or expired (i.e. the Err() method of
the context returns a non-nil value). In this case, the function returns
the best match found so far, together with a non-nil error.

This function can return one of the following errors:

- ErrUnsupportedImageType when one of the images is not image.RGBA

- ErrNotFound is no matches have been found after searching the entire canvas

- any error returned by ctx.Err()
*/
func Find(ctx context.Context, canvas, sample image.Image, tolerance float64, selection Selection) (Match, error) {
	var best Match
	var count = 0
	err := Foreach(ctx, canvas, sample, tolerance, func(m Match) error {
		if count == 0 || m.Similarity > best.Similarity {
			best = m
		}
		count++
		if selection == Fastest {
			return ErrBreak
		}
		return nil
	})
	if err == ErrBreak {
		err = nil
	}
	if err == nil {
		if count == 0 {
			err = ErrNotFound
		} else if selection == Only && count > 1 {
			err = ErrMultipleFound
		}
	}
	return best, err
}

/*
All locates and returns all matches of the sample image within a canvas image.
The returned matches are sorted using the Sort function of this package.

Tolerance is a value between 0 and 1 specifying how much difference is
tolerated between the sample and its match. Pass 0 to find exact matches only.
A reasonable value for fuzzy matching is around 0.05.

Searching stops if the context is canceled or expired (i.e. the Err() method of
the context returns a non-nil value). In this case, the function returns all
matches found so far, together with a non-nil error.

This function can return one of the following errors:

- ErrUnsupportedImageType when one of the images is not image.RGBA

- any error returned by ctx.Err()
*/
func All(ctx context.Context, canvas, sample image.Image, tolerance float64) ([]Match, error) {
	var mm []Match
	err := Foreach(ctx, canvas, sample, tolerance, func(m Match) error {
		mm = append(mm, m)
		return nil
	})
	Sort(mm)
	return mm, err
}

/*
Foreach locates matches of the sample image within a canvas image, and invokes
the provided callback function for each match.

Tolerance is a value between 0 and 1 specifying how much difference is
tolerated between the sample and its match. Pass 0 to find exact matches only.
A reasonable value for fuzzy matching is around 0.05.

Searching stops if the callback returns an error. You can use the convenient
predefined ErrBreak from this package, or return any other error. Searching
also stops if the context is canceled or expired (i.e. the Err() method of
the context returns a non-nil value).

Foreach can return one of the following errors:

- ErrUnsupportedImageType when one of the images is not image.RGBA

- any error returned by your callback

- any error returned by ctx.Err()
*/
func Foreach(ctx context.Context, canvas, sample image.Image, tolerance float64, f func(m Match) error) error {
	switch s := sample.(type) {
	case *image.RGBA:
		switch c := canvas.(type) {
		case *image.RGBA:
			return foreachRGBA(ctx, c, s, tolerance, f)
		default:
			return ErrUnsupportedImageType
		}
	default:
		return ErrUnsupportedImageType
	}
}
