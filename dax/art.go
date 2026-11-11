package dax

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobox/asset"
	"gobox/ega"
)

// ArtFile is one graphics archive for ExportArt: its DOS name and whether
// an animation's frames after the first are stored XORed against it (PIC*,
// FINAL*). Whether a block is a picture or an animation is its layout's
// (blockKind): FINAL holds both.
type ArtFile struct {
	Name   string
	XOR    bool
	NotArt []int // blocks that hold no picture (8X8D1's font, 201)
}

// ExportArt is the art import: every block of files, as the loader gives
// it, becomes a gobox/asset sheet, the DOS art in dst/dos and the Amiga
// blocks standing in for it (UseAmiga) in dst/amiga, under the DOS name.
// A file neither release has is skipped.
func ExportArt(l *Loader, dst string, files []ArtFile) error {
	pal := sheetPalette()
	for _, v := range []asset.Version{asset.DOS, asset.Amiga} {
		if v == asset.Amiga && !l.HasAmiga() {
			continue
		}
		dir := filepath.Join(dst, string(v))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		for _, f := range files {
			dos := l.File(f.Name)
			var blocks []asset.Block
			for id := range 256 {
				if slices.Contains(f.NotArt, id) {
					continue
				}
				var b, p []byte
				if dos != nil {
					b = dos.BlockShared(id)
				}
				if v == asset.Amiga {
					if b = l.AmigaShared(f.Name, id); b != nil {
						p = l.AmigaPalette(f.Name, id)
					}
				}
				if b == nil {
					continue
				}
				blk, err := artBlock(f, id, b)
				if err != nil {
					return fmt.Errorf("%s %s block %d: %w", v, f.Name, id, err)
				}
				blk.Palette = p
				blocks = append(blocks, blk)
			}
			if len(blocks) == 0 {
				continue
			}
			if err := asset.WriteSheet(dir, strings.ToUpper(f.Name), blocks, pal); err != nil {
				return err
			}
		}
	}
	return nil
}

// artBlock decodes block b unmasked; its colour base is its Amiga tag's slot.
func artBlock(f ArtFile, id int, b []byte) (asset.Block, error) {
	kind, err := blockKind(b)
	if err != nil {
		return asset.Block{}, err
	}
	blk := asset.Block{ID: id, Kind: kind}
	switch kind {
	case asset.Still:
		im := ParseImage(b, false, 0)

		blk.ColourBase = colourBase(b[9:])
		for i := range im.Count {
			fr := NewImage(1, im.Width, im.Height)
			fr.X, fr.Y = im.X, im.Y
			copy(fr.Data, im.Frame(i))
			blk.Frames = append(blk.Frames, asset.Frame{Image: fr})
		}
	case asset.Anim:
		blk.Frames = ParseSprite(b, false, 0, f.XOR)
		if len(b) > 22 { // the first frame's header holds the tag
			blk.ColourBase = colourBase(b[13:])
		}
	}
	return blk, nil
}

// blockKind is the layout b fills exactly: a picture's (ParseImage) or an
// animation's (ParseSprite); neither or both fails the import.
func blockKind(b []byte) (asset.Kind, error) {
	still := len(b) >= 17 && 17+int(b[8])*le16(b, 0)*le16(b, 2)*8/packed(b[9:]) == len(b)
	anim := len(b) >= 1
	p := 1
	for f := 0; anim && f < int(b[0]); f++ {
		if p+21 > len(b) {
			anim = false
			break
		}
		p += 21 + le16(b, p+4)*le16(b, p+6)*8/packed(b[p+12:])
	}
	anim = anim && p == len(b)
	switch {
	case still && !anim:
		return asset.Still, nil
	case anim && !still:
		return asset.Anim, nil
	}
	return "", fmt.Errorf("%d bytes: picture layout %v, animation layout %v", len(b), still, anim)
}

func le16(b []byte, i int) int { return int(b[i]) | int(b[i+1])<<8 }

// packed is how many pixels a byte of a picture holds.
func packed(hdr []byte) int {
	if wide(hdr) {
		return 1
	}
	return 2
}

// colourBase is where amigaTag moves a tagged header's colours (0: untagged).
func colourBase(hdr []byte) int {
	if len(hdr) < 3 || hdr[0] != 'A' || hdr[1] != 'M' {
		return 0
	}
	return ega.AmigaBase + 16*int(hdr[2])
}

// sheetPalette colours a sheet's PNG: the EGA colours, the rest grey (an
// Amiga picture's own colours are its block's Palette).
func sheetPalette() color.Palette {
	p := make(color.Palette, 256)
	for i := range p {
		p[i] = color.Gray{Y: byte(i)}
	}
	for i := range 16 {
		p[i] = ega.RGBA(i)
	}
	return p
}
