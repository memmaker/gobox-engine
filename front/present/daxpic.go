// daxpic.go: pictures the game names by DAX file and block (Img nil) are
// decoded here from Data.DAX, so a game never decodes images.

package present

import (
	"image"
	"image/color"
	"slices"
	"strings"
	"time"

	"gobox/asset"
	"gobox/ega"
	"gobox/front"
)

type daxKey struct {
	file               string
	block, frame, mask int
}

var daxPics = map[daxKey]image.Image{} // decoded once; nil: not decodable

// daxArt is the game's imported art (asset.Dir of Data.Files, which the
// game's import writes: dax.ImportArt) and whether its Amiga art shows.
var daxArt struct {
	art   *asset.Art // nil: no art
	amiga bool       // the Amiga version shows (F12)
}

// amigaShown reports whether the Amiga art shows.
func amigaShown() bool { return daxArt.amiga && hasAmiga() }

// hasAmiga reports whether the game has Amiga art to switch to.
func hasAmiga() bool { return daxArt.art != nil && daxArt.art.Has(asset.Amiga) }

// artBlock is block id of file as shown: the Amiga one while it shows and
// there is one, else the DOS one; nil: none.
func artBlock(file string, id int) *asset.Block {
	if daxArt.art == nil {
		return nil
	}
	file = strings.TrimSuffix(strings.ToUpper(file), ".DAX")
	if amigaShown() {
		if b := daxArt.art.Block(asset.Amiga, file, id); b != nil {
			return b
		}
	}
	return daxArt.art.Block(asset.DOS, file, id)
}

// Arts is the shown art (DOS: true) and how to switch it, when Amiga art
// stands in for the game's own (set nil: one art). The shell cross-fades
// the switch (F12); every picture and border decodes anew.
func (f *Front) Arts() (dos bool, set func(bool)) {
	if !hasAmiga() {
		return true, nil
	}
	return true, func(dos bool) {
		daxArt.amiga = !dos
		clear(daxPics)
		clear(spriteAnims) // the other art's frames and delays
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
	m := artRGBA(artBlock(file, block), file, frame, mask)
	if m == nil {
		daxPics[k] = nil
		return nil
	}
	daxPics[k] = m
	return m
}

// artRGBA is frame of b in its colours, colour index mask left transparent
// (nil: no such frame).
func artRGBA(b *asset.Block, file string, frame, mask int) *image.RGBA {
	if b == nil || frame < 0 || frame >= len(b.Frames) {
		return nil
	}
	im := b.Frames[frame].Image
	w := im.Width * 8
	if w == 0 {
		return nil
	}
	m := image.NewRGBA(image.Rect(0, 0, w, im.Height))
	pal := b.Palette
	switch b.Kind {
	case asset.VGAStill:
		for i, c := range im.Data {
			if int(c) != mask {
				m.SetRGBA(i%w, i/w, vgaRGBA(pal, c))
			}
		}
	case asset.VGAAnim:
		sprite := strings.HasPrefix(strings.ToUpper(file), "SPRIT") // O124:0981: a sprite's 0 and 13 swap, 13 is clear
		for i, c := range im.Data {
			switch {
			case sprite && c == 0, int(c) == mask:
			case sprite && c == 13:
				m.SetRGBA(i%w, i/w, vgaRGBA(pal, 0))
			default:
				m.SetRGBA(i%w, i/w, vgaRGBA(pal, c))
			}
		}
	default:
		for i, c := range im.Data {
			if (c < asset.Clear || c >= ega.AmigaBase) && int(c) != mask { // EGA, or an Amiga block's (ColourBase on)
				m.SetRGBA(i%w, i/w, pixelRGBA(c, pal))
			}
		}
	}
	return m
}

// vgaRGBA is colour c of a VGA block's palette (RGB triplets).
func vgaRGBA(pal []byte, c uint8) color.RGBA {
	i := int(c) * 3
	return color.RGBA{pal[i], pal[i+1], pal[i+2], 255}
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

// spriteAnim is an animation block's frame delays and when it came up.
type spriteAnim struct {
	delays      []time.Duration
	loop        int // the frame play goes back to after the last
	start, seen time.Time
}

var spriteAnims = map[daxKey]*spriteAnim{} // by file and block (frame 0)

// spriteFrame is the frame of pc's DAX animation block shown at now, as the
// originals' view window plays it (por-go anim.go, ovr26 DisplayInput): each
// frame for its delay in tenths of a second, then the next, after the last
// back to its loop frame (a VGA block's own, BR O10F:109E; else the first).
// ok is false for a still picture, a sprite (SPRIT*: the game picks its
// frame, O10F:102E), one the game cycles itself (Frames) or names by
// pixels or pool name.
func spriteFrame(pc front.Picture, now time.Time) (int, bool) {
	if pc.File == "" || pc.Img != nil || pc.Pix != nil || pc.Name != "" || len(pc.Frames) > 0 || daxArt.art == nil || strings.HasPrefix(strings.ToUpper(pc.File), "SPRIT") {
		return 0, false
	}
	k := daxKey{file: pc.File, block: pc.Block}
	a, ok := spriteAnims[k]
	if !ok {
		a = &spriteAnim{start: now}
		var fs []asset.Frame
		loop := 0
		if b := artBlock(pc.File, pc.Block); b != nil && !b.Kind.Pictured() {
			fs, loop = b.Frames, b.Loop
		}
		a.loop = loop
		for _, f := range fs {
			a.delays = append(a.delays, max(time.Duration(f.Delay)*100*time.Millisecond, 10*time.Millisecond)) // a 0 delay: the next tick (10 ms)
		}
		spriteAnims[k] = a
	} else if now.Sub(a.seen) > 250*time.Millisecond { // shown again after a while: from its first frame
		a.start = now
	}
	a.seen = now
	if len(a.delays) < 2 {
		return 0, false
	}
	return animFrame(a.delays, a.loop, now.Sub(a.start)), true
}

// animFrame is the frame shown t into playing delays once from the first,
// then over and over from loop.
func animFrame(delays []time.Duration, loop int, t time.Duration) int {
	var intro, cycle time.Duration
	for i, d := range delays {
		if i < loop {
			intro += d
		} else {
			cycle += d
		}
	}
	if t >= intro {
		t = intro + (t-intro)%cycle
	}
	for i, d := range delays {
		if t < d {
			return i
		}
		t -= d
	}
	return loop
}
