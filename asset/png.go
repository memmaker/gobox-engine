package asset

import (
	"bytes"
	"image"
	"image/png"
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
