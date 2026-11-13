package importer

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	"gobox/asset"
	"gobox/asset/orig"
)

// CoverFile is the cover picture Run writes into a Game.Cover game's folder.
const CoverFile = "COVER.PNG"

// writeCover writes out (CoverFile, the launcher's label) from the title
// picture: block 2 of TITLE.DAX in dosDir, in the EGA palette.
func writeCover(dosDir, out string) error {
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
