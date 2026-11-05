// daxpic.go: pictures the game names by DAX file and block (Img nil) are
// decoded here from Data.DAX, so a game never decodes images.

package present

import (
	"image"
	"image/color"
	"slices"
	"strings"

	"gobox/dax"
	"gobox/ega"
	"gobox/front"
)

type daxKey struct {
	file               string
	block, frame, mask int
}

var daxPics = map[daxKey]image.Image{} // decoded once; nil: not decodable

// Arts is the shown art (DOS: true) and how to switch it, when Amiga
// archives stand in for the game's DAX files (set nil: one art). The shell
// cross-fades the switch (F12); every picture and border decodes anew.
func (f *Front) Arts() (dos bool, set func(bool)) {
	if !daxLoader.HasAmiga() {
		return true, nil
	}
	return true, func(dos bool) {
		daxLoader.SetDOS(dos)
		clear(daxPics)
		f.ArtNew = true
	}
}

// daxImage is frame of block of file in the EGA colours (nil: none).
func daxImage(file string, block, frame int) image.Image { return daxMasked(file, block, frame, -1) }

// daxMasked is daxImage with colour index mask transparent too (-1: none).
func daxMasked(file string, block, frame, mask int) image.Image {
	k := daxKey{file, block, frame, mask}
	if m, ok := daxPics[k]; ok {
		return m
	}
	var m *image.RGBA
	if daxLoader != nil {
		b := daxLoader.BlockShared(file, block)
		if im := dax.ParseVGA(b, false, 0); im != nil && frame < im.Count { // a VGA block (byte height, width: an EGA one's height high byte is 0)
			m = image.NewRGBA(image.Rect(0, 0, im.Width*8, im.Height))
			for i, c := range im.Frame(frame) {
				if int(c) != mask {
					m.Set(i%(im.Width*8), i/(im.Width*8), im.Palette[c])
				}
			}
		} else if im := daxStill(file, b, frame); im != nil {
			m = image.NewRGBA(image.Rect(0, 0, im.Width*8, im.Height))
			pal := daxLoader.AmigaPalette(file, block)
			for i, c := range im.Frame(min(frame, im.Count-1)) {
				if c < dax.Clear && int(c) != mask {
					m.Set(i%(im.Width*8), i/(im.Width*8), pixelRGBA(c, pal))
				}
			}
		}
	}
	if m == nil {
		daxPics[k] = nil
		return nil
	}
	daxPics[k] = m
	return m
}

// refMask is the colour index of pc's DAX picture left transparent: its own
// Mask, else Data.DaxMask's for its file, else -1 (none).
func refMask(pc front.Picture) int {
	if pc.Mask != nil {
		return int(*pc.Mask)
	}
	if m, ok := daxMask[pc.File]; ok {
		return m
	}
	return -1
}

// daxPictures is ps with the pictures named by file and block decoded.
func daxPictures(ps []front.Picture) []front.Picture {
	out := ps
	for i, pc := range ps {
		if pc.Img != nil || pc.Fill || pc.File == "" && pc.Name == "" {
			continue
		}
		m := poolImage(pc)
		if m == nil && pc.File != "" {
			m = daxMasked(pc.File, pc.Block, pc.Frame, refMask(pc))
		}
		if m != nil {
			if &out[0] == &ps[0] {
				out = slices.Clone(ps)
			}
			out[i].Img = m
		}
	}
	return out
}

// pixelRGBA is pixel value c in the EGA colours, or past ega.AmigaBase in
// its picture's own Amiga palette pal (8-bit RGB triplets), else the shared
// ega.Amiga slots.
// ponytail: a 32-colour picture's index is taken modulo its palette, so its archive's Slot must be even.
func pixelRGBA(c uint8, pal []byte) color.RGBA {
	if c < ega.AmigaBase {
		return ega.RGBA(int(c))
	}
	if n := len(pal) / 3; n > 0 {
		i := int(c-ega.AmigaBase) % n * 3
		return color.RGBA{pal[i], pal[i+1], pal[i+2], 255}
	}
	return ega.AmigaRGBA(c)
}

// daxStill is frame of an EGA picture block: a plain image, else frame of an
// animation block (PIC*, FINAL*: frames XORed on the first; SPRIT*), so a
// view picture such as Silver Blades' PIC1 80 shows; nil if neither.
func daxStill(file string, b []byte, frame int) *dax.Image {
	if im := dax.ParseImage(b, false, 0); im != nil && im.Width > 0 && frame < im.Count {
		return im
	}
	up := strings.ToUpper(file)
	fs := dax.ParseSprite(b, false, 0, strings.HasPrefix(up, "PIC") || strings.HasPrefix(up, "FINAL"))
	if frame >= len(fs) || fs[frame].Image.Width == 0 {
		return nil
	}
	return fs[frame].Image
}
