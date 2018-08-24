locateimage (Go)
================

[![GoDoc](https://godoc.org/github.com/andreyvit/locateimage?status.svg)](https://godoc.org/github.com/andreyvit/locateimage)

Performs an (exact or fuzzy) search of a sample image within a larger image, returning the coordinates and similarity scores of the matches.


Installation
------------

This package contains a Go module, so in Go 1.11+ module mode you can import it directly. Otherwise do:

```bash
go get -u github.com/andreyvit/locateimage
```


Example
-------

See [godoc.org/github.com/andreyvit/locateimage](https://godoc.org/github.com/andreyvit/locateimage) for a full reference.

`All` locates all matches:

```go
mm, err := locateimage.All(context.Background(), canvas, sample, 0.04)
if err != nil {
    log.Fatal(err)
}
log.Print(mm)
```

`Find` locates a single match:

```go
m, err := locateimage.Find(context.Background(), canvas, sample, 0, locateimage.Fastest)
if err != nil {
    log.Print(err)
} else {
    log.Printf("sample found at %v, similarity = %.*f%%", m.Rect, locateimage.SimilarityDigits-2, 100*m.Similarity)
}
```

`Find` returns `ErrNotFound` if no match is found. Available selection modes are `Fastest`, `Best` and `Only`.

`Foreach` invokes a callback for each match:

```go
err = locateimage.Foreach(context.Background(), canvas, sample, 0.04, func(m locateimage.Match) error {
    log.Print(m)
    return nil
})
if err != nil {
    log.Fatal(err)
}
```


Caveats
-------

The package currently only deals with `image.RGBA`-encoded images. You can use `Convert` to convert any image to this format. (Note that reading a PNG file returns an NRGBA format, so the conversion will be required. It makes sense to add support for formats like NRGBA in the future.)

The search is currently slow, taking tens or hundreds of milliseconds on large images (screenshots of a 27" screen). This can likely be improved, and contributions are welcome.


Versions
--------

None currently.
