package dax

import (
	"encoding/binary"
	"errors"
	"fmt"

	"gobox/asset"
)

// Image, SpriteFrame and Clear are gobox/asset's (the canonical art types).
type (
	Image       = asset.Image
	SpriteFrame = asset.Frame
)

// Clear marks a masked-out pixel.
const Clear = asset.Clear

// NewImage allocates a blank image (asset.NewImage).
func NewImage(count, width, height int) *Image { return asset.NewImage(count, width, height) }

// ParseImage decodes a picture block. If masked, maskColour becomes the
// image's transparent colour (the original's "masked load"). ErrNotImage:
// b is not a picture (shorter than its header says, e.g. a font block), so
// nothing is allocated beyond what b holds.
func ParseImage(b []byte, masked bool, maskColour byte) (*Image, error) {
	if len(b) < 17 || 17+int(b[8])*le16(b, 0)*le16(b, 2)*8/packed(b[9:]) > len(b) {
		return nil, fmt.Errorf("%d bytes: %w", len(b), ErrNotImage)
	}
	im := &Image{
		Height: int(binary.LittleEndian.Uint16(b[0:])),
		Width:  int(binary.LittleEndian.Uint16(b[2:])),
		X:      int(binary.LittleEndian.Uint16(b[4:])),
		Y:      int(binary.LittleEndian.Uint16(b[6:])),
		Count:  int(b[8]),
		Mask:   -1,
	}
	if masked {
		im.Mask = Clear
	}
	im.BPP = im.Height * im.Width * 8
	im.Data = make([]byte, im.Count*im.BPP)
	unpack(im.Data, b[17:], wide(b[9:]))
	mask := -1
	if masked {
		mask = int(maskColour)
	}
	amigaTag(b[9:], im.Data, mask)
	if masked {
		clearMask(im.Data, maskColour)
	}
	return im, nil
}

// ErrNotImage: a block ParseImage cannot read as a picture.
var ErrNotImage = errors.New("not a picture block")

// ParseSprite decodes an animation block: u8 frame count, then per frame a
// 21-byte header (u32 delay, u16 height, u16 width cells, u16 x, u16 y, 9
// unused) followed by its pixels. With xor set (PIC*, FINAL*) frames after
// the first are stored XORed against frame 0 and are reconstructed here.
func ParseSprite(b []byte, masked bool, maskColour byte, xor bool) []SpriteFrame {
	if len(b) < 1 {
		return nil
	}
	n := int(b[0])
	var frames []SpriteFrame
	p := 1
	var first []byte
	for f := 0; f < n; f++ {
		if p+21 > len(b) {
			break
		}
		delay := int(binary.LittleEndian.Uint32(b[p:]))
		h := int(binary.LittleEndian.Uint16(b[p+4:]))
		w := int(binary.LittleEndian.Uint16(b[p+6:]))
		x := int(binary.LittleEndian.Uint16(b[p+8:]))
		y := int(binary.LittleEndian.Uint16(b[p+10:]))
		p += 21
		size := h * w * 4
		bytes := wide(b[p-9:])
		if bytes {
			size *= 2
		}
		if p+size > len(b) {
			break // not an animation's frame: its pixels are past the block
		}
		raw := append([]byte(nil), b[p:p+size]...)
		p += size
		if f == 0 {
			first = raw
		} else if xor && string(b[p-size-9:p-size-7]) != "AM" { // Amiga frames are imported whole
			for i := range raw {
				if i < len(first) {
					raw[i] ^= first[i]
				}
			}
		}
		im := &Image{Height: h, Width: w, X: x, Y: y, Count: 1, BPP: h * w * 8, Mask: -1}
		im.Data = make([]byte, im.BPP)
		unpack(im.Data, raw, bytes)
		mask := -1
		if masked {
			mask = int(maskColour)
		}
		amigaTag(b[p-size-9:], im.Data, mask)
		if masked {
			im.Mask = Clear
			clearMask(im.Data, maskColour)
		}
		frames = append(frames, SpriteFrame{Delay: delay, Image: im})
	}
	return frames
}

// unpack fills px from packed nibbles (high first), or from one byte per
// pixel for a wide (32-colour Amiga) picture.
func unpack(px, src []byte, bytes bool) {
	if bytes {
		copy(px, src)
		return
	}
	for i := 0; i < len(src) && i*2+1 < len(px); i++ {
		px[i*2] = src[i] >> 4
		px[i*2+1] = src[i] & 15
	}
}

func clearMask(d []byte, colour byte) {
	for i, p := range d {
		if p == colour {
			d[i] = Clear
		}
	}
}
