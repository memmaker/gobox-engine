package orig

import (
	"encoding/binary"
	"image/color"

	"gobox/asset"
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
		im.Palette[i] = asset.EGARGBA(i & 15)
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

// ParseVGAAnim decodes a VGA animation block (Buck Rogers, Pools of
// Darkness PIC*/SPRIT*: BR O124:02D0): u16 height, u16 width cells, u16 x,
// u16 y, u8 last frame (at most 8), u8 first colour, u8 colours, the
// colours' 6-bit RGB, their EGA stand-ins (two a byte), 4 bytes of 2-bit
// stand-ins, u8 key frame, u8 loop frame, per frame u16 size and u8 delay
// (tenths of a second), then each frame packed (R129:0103) and, but for the
// key frame, XORed on it (R129:00DB). Playback (O10F:105F) shows frame 0
// on, each for its delay, and after the last goes back to loop. nil: not
// such a block.
func ParseVGAAnim(b []byte) (frames []SpriteFrame, loop int) {
	if len(b) < 11 {
		return nil, 0
	}
	h, w := int(binary.LittleEndian.Uint16(b)), int(binary.LittleEndian.Uint16(b[2:]))
	x, y := int(binary.LittleEndian.Uint16(b[4:])), int(binary.LittleEndian.Uint16(b[6:]))
	n, first, colours := min(int(b[8]), 8)+1, int(b[9]), int(b[10])
	p := 11 + colours*3 + colours/2 + 4
	if h == 0 || w == 0 || h > 200 || w > 40 || p+2+3*n > len(b) {
		return nil, 0
	}
	pal := make([]color.RGBA, 256)
	for i := range pal {
		pal[i] = asset.EGARGBA(i & 15)
	}
	for k := 0; k < colours && first+k < 256; k++ {
		c := func(v byte) uint8 { v &= 63; return v<<2 | v>>4 }
		pal[first+k] = color.RGBA{c(b[11+k*3]), c(b[12+k*3]), c(b[13+k*3]), 255}
	}
	key, loop := int(b[p]), int(b[p+1])
	p += 2
	sizes := make([]int, n)
	frames = make([]SpriteFrame, n)
	for i := range n {
		sizes[i] = int(binary.LittleEndian.Uint16(b[p:]))
		frames[i].Delay = int(b[p+2])
		p += 3
	}
	if key >= n || loop >= n {
		return nil, 0
	}
	size := h * w * 8
	for i := range n {
		if p+sizes[i] > len(b) {
			return nil, 0
		}
		im := NewImage(1, w, h)
		im.X, im.Y, im.Palette = x, y, pal
		if unpackRuns(im.Data, b[p:p+sizes[i]]) != size {
			return nil, 0
		}
		frames[i].Image = im
		p += sizes[i]
	}
	if p != len(b) {
		return nil, 0
	}
	for i, f := range frames {
		if i != key {
			for j := range size {
				f.Image.Data[j] ^= frames[key].Image.Data[j]
			}
		}
	}
	return frames, loop
}

// unpackRuns = R129:0103: a control byte c up to 0x7F copies the next c+1
// bytes, a bigger one repeats the next byte 257-c times; src is used up.
func unpackRuns(dst, src []byte) int {
	o := 0
	for i := 0; i < len(src); {
		c := int(src[i])
		if c <= 0x7F {
			end := min(i+1+c+1, len(src))
			o += copy(dst[min(o, len(dst)):], src[i+1:end])
			i = end
			continue
		}
		if i+1 >= len(src) {
			break
		}
		for range 257 - c {
			if o < len(dst) {
				dst[o] = src[i+1]
			}
			o++
		}
		i += 2
	}
	return o
}
