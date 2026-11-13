package app

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"gobox/asset"
	"gobox/asset/orig"
)

// CopyFS writes the files under root in fsys into dst (root's own path cut
// off), overwriting what is there: a game's -setup step puts its own
// embedded assets beside the imported game files.
func CopyFS(fsys fs.FS, root, dst string) error {
	return fs.WalkDir(fsys, root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		out := filepath.Join(dst, filepath.FromSlash(strings.TrimPrefix(p, path.Clean(root)+"/")))
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, b, 0o644)
	})
}

// WriteCover writes out (COVER.PNG for gameimport's label) from the title
// picture: block 2 of TITLE.DAX in dosDir, in the EGA palette.
func WriteCover(dosDir, out string) error {
	l, err := orig.NewLoader(dosDir)
	if err != nil {
		return err
	}
	b := l.BlockShared("TITLE", 2)
	if b == nil {
		return errors.New("TITLE.DAX has no title picture")
	}
	im, err := orig.ParseImage(b, false, 0)
	if err != nil {
		return fmt.Errorf("TITLE.DAX block 2: %w", err)
	}
	cp := make(color.Palette, 16)
	for i := range cp {
		cp[i] = asset.EGARGBA(i)
	}
	img := image.NewPaletted(image.Rect(0, 0, asset.ScreenW, asset.ScreenH), cp)
	d, w := im.Frame(0), im.Width*8
	for y := 0; y < im.Height && y < asset.ScreenH; y++ {
		for x := 0; x < w && x < asset.ScreenW; x++ {
			img.SetColorIndex(x, y, d[y*w+x]&15)
		}
	}
	return asset.WritePNG(out, img)
}

// ImportArt is a game's art import, run at its start: the DAX archives in
// dosDir (with the Amiga ones in amiga/ beside it) as gobox/asset sheets in
// asset.Dir(dosDir), when missing or stale (dax.ImportArtIfStale). The
// presenter reads only those.
func ImportArt(dosDir string) error {
	return orig.ImportArtIfStale(dosDir, filepath.Join(filepath.Dir(dosDir), "amiga"), asset.Dir(dosDir))
}
