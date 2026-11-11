// Package asset holds an adventure folder's canonical data: every picture
// and animation as a PNG pixel atlas with a metadata text file (a sheet,
// sheet.go), the Image and Frame types the games draw, and the numbered
// binary blocks and JSON records (data.go).
package asset

import "image/color"

// Image is a picture: a decoded block of the original's art (dax, at import) or
// a sheet's (Sheet.Image): Count frames of Height x Width*8 pixels,
// one byte per pixel (EGA colour 0..15). A masked load (0464:0111 with the
// mask flag) does what the original does: pixels of the mask colour go into
// a separate mask and read as colour 0. Here they are stored as Clear (16)
// and Mask is set to Clear: masked draws skip them, plain draws show black.
//
// Block header: u16 height, u16 width (8-pixel cells), u16 x, u16 y, u8 count,
// 8 unused bytes, then count frames of height*width*4 bytes, two 4-bit
// pixels per byte, high nibble first.
type Image struct {
	Height, Width int // Width in 8-pixel cells
	X, Y          int
	Count         int
	BPP           int // bytes (pixels) per frame
	Data          []byte
	Mask          int          // Clear for a masked load, -1 = none
	File          string       // where it came from (the loader sets it): the DAX file...
	Block, Base   int          // ...its block, and the block's frame that is frame 0 here (an animation's frames)
	Alt           *Image       // the same picture in the other version's art (DOS for an Amiga one); a smaller one fitted to this size, a bigger one its own; nil: none
	Palette       []color.RGBA // VGA pictures (ParseVGA): the 256 colours Data indexes; nil for EGA
}

// Clear marks a masked-out pixel; the page draws it as colour 0.
const Clear = 16

// NewImage allocates a blank image.
func NewImage(count, width, height int) *Image {
	im := &Image{Height: height, Width: width, Count: count, BPP: height * width * 8, Mask: -1}
	im.Data = make([]byte, count*im.BPP)
	return im
}

// Frame returns the pixel slice for frame i.
func (im *Image) Frame(i int) []byte { return im.Data[i*im.BPP : (i+1)*im.BPP] }

// Clone deep-copies the image.
func (im *Image) Clone() *Image {
	c := *im
	c.Data = append([]byte(nil), im.Data...)
	if im.Alt != nil {
		c.Alt = im.Alt.Clone()
	}
	return &c
}

// FlipLeftToRight mirrors every frame horizontally.
func (im *Image) FlipLeftToRight() {
	w := im.Width * 8
	for f := 0; f < im.Count; f++ {
		d := im.Frame(f)
		for y := 0; y < im.Height; y++ {
			row := d[y*w : (y+1)*w]
			for i, j := 0, w-1; i < j; i, j = i+1, j-1 {
				row[i], row[j] = row[j], row[i]
			}
		}
	}
	if im.Alt != nil {
		im.Alt.FlipLeftToRight()
	}
}

// Frame is one frame of an animation: its picture and how long it shows.
type Frame struct {
	Delay int // ticks (tenths of a second)
	Image *Image
}

// Fit is im redrawn at width (in 8-pixel units) by height: scaled, nearest
// pixel, to fit with its aspect kept, centred, the rest Clear.
func (im *Image) Fit(width, height int) *Image {
	out := NewImage(im.Count, width, height)
	out.Mask, out.File, out.Block, out.Base = im.Mask, im.File, im.Block, im.Base
	sw, dw := im.Width*8, width*8
	if sw == 0 || im.Height == 0 {
		return out
	}
	s := min(float64(dw)/float64(sw), float64(height)/float64(im.Height))
	w, h := int(float64(sw)*s), int(float64(im.Height)*s)
	x0, y0 := (dw-w)/2, (height-h)/2
	for f := range im.Count {
		src, dst := im.Frame(f), out.Frame(f)
		for i := range dst {
			x, y := i%dw-x0, i/dw-y0
			if x < 0 || y < 0 || x >= w || y >= h {
				dst[i] = Clear
				continue
			}
			dst[i] = src[min(y*im.Height/h, im.Height-1)*sw+min(x*sw/w, sw-1)]
		}
	}
	return out
}
