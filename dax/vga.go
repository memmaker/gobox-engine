package dax

import (
	"image/color"

	"gobox/ega"
)

// ParseVGA decodes a VGA picture block (Pools of Darkness): a header (height,
// width in 8-pixel cells, x, y, u16 frame count, first colour, colours-1),
// the VGA palette (6-bit RGB) of the colours the picture adds, each added
// colour's EGA stand-in (two a byte), padding, then one byte a pixel. Data
// keeps the 8-bit indices; Palette is the 256 colours (indices below 16 and
// those the picture adds not set are the EGA's), with the masked colour, if
// masked, transparent. nil: not a picture.
func ParseVGA(b []byte, masked bool, mask byte) *Image {
	if len(b) < 10 {
		return nil
	}
	height, width := int(b[0]), int(b[1])
	count := int(b[6]) | int(b[7])<<8
	first, colours := int(b[8]), int(b[9])+1
	pix := 10 + colours*3 + colours/2
	if first == 32 && colours == 112 {
		pix += 32
	} else {
		pix += 4
	}
	size := height * width * 8
	if size == 0 || count == 0 || pix+count*size > len(b) {
		return nil
	}
	im := NewImage(count, width, height)
	copy(im.Data, b[pix:])
	im.Palette = make([]color.RGBA, 256)
	for i := range im.Palette {
		im.Palette[i] = ega.RGBA(i & 15)
	}
	for k := 0; k < colours && first+k < 256; k++ {
		c := func(v byte) uint8 { v &= 63; return v<<2 | v>>4 }
		im.Palette[first+k] = color.RGBA{c(b[10+k*3]), c(b[11+k*3]), c(b[12+k*3]), 255}
	}
	if masked {
		im.Palette[mask].A = 0
		im.Palette[mask].R, im.Palette[mask].G, im.Palette[mask].B = 0, 0, 0
	}
	return im
}
