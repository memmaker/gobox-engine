package asset

import (
	"errors"
	"image"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// WritePNG returns the write's error (a read-only folder) and writes a PNG otherwise.
func TestWritePNG(t *testing.T) {
	dir := t.TempDir()
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	if err := WritePNG(filepath.Join(dir, "sub", "a.png"), img); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o500); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o700)
	if err := WritePNG(filepath.Join(dir, "b.png"), img); !errors.Is(err, fs.ErrPermission) {
		t.Errorf("WritePNG in a read-only folder: %v, want the permission error", err)
	}
}
