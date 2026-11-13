package asset

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/fs"
	"os"
	"path/filepath"
)

// WritePNG writes img as a PNG file at path (its folder made first), in one
// write. The one place a PNG gets written to disk: every error comes back.
func WritePNG(path string, img image.Image) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

// DecodePNG is b as a picture, or why it is not a PNG.
func DecodePNG(b []byte) (image.Image, error) {
	im, err := png.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, fmt.Errorf("not a PNG: %w", err)
	}
	return im, nil
}

// ReadPNG is the PNG file at path, its bytes checked to decode; nil, nil
// when there is no file.
func ReadPNG(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := DecodePNG(b); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return b, nil
}

// ErrIconShape: a PNG icon is square or twice as wide as high.
var ErrIconShape = errors.New("not a PNG, square or twice as wide as high")

// IconShape checks that b is a PNG a combat icon can be cut from.
func IconShape(b []byte) error {
	im, err := DecodePNG(b)
	if err != nil {
		return err
	}
	if r := im.Bounds(); r.Dx() != r.Dy() && r.Dx() != 2*r.Dy() {
		return ErrIconShape
	}
	return nil
}

// IconPose cuts pose (0 left, 1 right; the same square when im is square)
// out of im, mirrored when flip.
func IconPose(im image.Image, pose int, flip bool) image.Image {
	b := im.Bounds()
	w := b.Dy()
	x0 := b.Min.X + pose*(b.Dx()-w)
	out := image.NewRGBA(image.Rect(0, 0, w, w))
	for y := 0; y < w; y++ {
		for x := 0; x < w; x++ {
			sx := x0 + x
			if flip {
				sx = x0 + w - 1 - x
			}
			out.Set(x, y, im.At(sx, b.Min.Y+y))
		}
	}
	return out
}
