// daxpic.go: pictures the game names by DAX file and block (Img nil) are
// decoded here from Data.DAX, so a game never decodes images.

package present

import (
	"image"
	"slices"

	"gobox/dax"
	"gobox/ega"
	"gobox/front"
)

type daxKey struct {
	file               string
	block, frame, mask int
}

var daxPics = map[daxKey]image.Image{} // decoded once; nil: not decodable

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
		if im := dax.ParseImage(daxLoader.BlockShared(file, block), false, 0); im != nil && frame < im.Count {
			m = image.NewRGBA(image.Rect(0, 0, im.Width*8, im.Height))
			for i, c := range im.Frame(frame) {
				if c < dax.Clear && int(c) != mask {
					m.Set(i%(im.Width*8), i/(im.Width*8), ega.RGBA(int(c)))
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

// daxPictures is ps with the pictures named by file and block decoded.
func daxPictures(ps []front.Picture) []front.Picture {
	out := ps
	for i, pc := range ps {
		if pc.Img != nil || pc.Fill || pc.File == "" {
			continue
		}
		mask := -1
		if pc.Mask != nil {
			mask = int(*pc.Mask)
		}
		if m := daxMasked(pc.File, pc.Block, pc.Frame, mask); m != nil {
			if &out[0] == &ps[0] {
				out = slices.Clone(ps)
			}
			out[i].Img = m
		}
	}
	return out
}
