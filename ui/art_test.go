package ui

import (
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
)

// useArt points the presenter at the art the game in dosDir imports
// (dax.ImportArtIfStale, as the game's start does) until the test ends;
// false: no such game files.
func useArt(t *testing.T, dosDir string) bool {
	t.Helper()
	if _, err := os.Stat(dosDir); err != nil {
		return false
	}
	if err := orig.ImportArtIfStale(dosDir, filepath.Join(filepath.Dir(dosDir), "amiga"), asset.Dir(dosDir)); err != nil {
		t.Fatal(err)
	}
	old := daxArt
	daxArt.art, daxArt.amiga = asset.OpenArt(asset.Dir(dosDir)), false
	clear(daxPics)
	t.Cleanup(func() { daxArt = old; clear(daxPics); clear(spriteAnims) })
	return true
}

// TestArtEqualsDAX: every picture the presenter shows from a game's
// imported art (daxMasked, spriteFrame's delays) is the one it decoded from
// the DAX archives before the import (daxDecode), DOS and Amiga, for every
// installed game.
func TestArtEqualsDAX(t *testing.T) {
	home, _ := os.UserHomeDir()
	n := 0
	for _, g := range []string{"por", "cotab", "sotsb", "pod", "buckrogers", "buckrogers2"} {
		dir := filepath.Join(home, "Projects", g+"-go", "data", "dos")
		if !useArt(t, dir) {
			t.Logf("skip %s: no data", g)
			continue
		}
		l, err := orig.NewLoader(dir)
		if err != nil {
			t.Fatal(err)
		}
		l.UseAmiga(filepath.Join(filepath.Dir(dir), "amiga"))
		ents, _ := os.ReadDir(dir)
		for _, amiga := range []bool{false, true} {
			if amiga && !l.HasAmiga() {
				continue
			}
			l.SetDOS(!amiga)
			daxArt.amiga = amiga
			for _, e := range ents {
				file := strings.ToUpper(strings.TrimSuffix(e.Name(), filepath.Ext(e.Name())))
				if !strings.EqualFold(filepath.Ext(e.Name()), ".dax") {
					continue
				}
				for id := range 256 {
					b := l.BlockShared(file, id)
					if b == nil {
						continue
					}
					if _, err := os.Stat(filepath.Join(asset.Dir(dir), "dos", file+".png")); err != nil {
						break // no art in it: a script, map or data archive
					}
					want := daxDecode(l, file, id)
					for frame := range len(want) + 1 {
						var w *image.RGBA
						if frame < len(want) {
							w = want[frame]
						}
						got := artRGBA(artBlock(file, id), file, frame, -1)
						if !sameRGBA(w, got) {
							t.Fatalf("%s amiga %v %s %d frame %d: art %v, DAX %v", g, amiga, file, id, frame, got != nil, w != nil)
						}
					}
					n += len(want)
				}
			}
		}
	}
	t.Logf("%d frames", n)
}

func sameRGBA(a, b *image.RGBA) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Rect == b.Rect && string(a.Pix) == string(b.Pix)
}

// daxDecode is every frame daxMasked decoded from the archives before the
// art import (unmasked), until the first it had none for.
func daxDecode(l *orig.Loader, file string, block int) []*image.RGBA {
	var out []*image.RGBA
	b := l.BlockShared(file, block)
	rgba := func(im *orig.Image, px []byte, colour func(c byte) (color.RGBA, bool)) *image.RGBA {
		w := im.Width * 8
		m := image.NewRGBA(image.Rect(0, 0, w, im.Height))
		for i, c := range px {
			if rgb, ok := colour(c); ok {
				m.SetRGBA(i%w, i/w, rgb)
			}
		}
		return m
	}
	if im := orig.ParseVGA(b, false, 0); im != nil {
		for f := range im.Count {
			out = append(out, rgba(im, im.Frame(f), func(c byte) (color.RGBA, bool) { return im.Palette[c], true }))
		}
		return out
	}
	if fs, _ := orig.ParseVGAAnim(b); fs != nil {
		sprite := strings.HasPrefix(strings.ToUpper(file), "SPRIT")
		for _, f := range fs {
			im := f.Image
			out = append(out, rgba(im, im.Data, func(c byte) (color.RGBA, bool) {
				switch {
				case sprite && c == 0:
					return color.RGBA{}, false
				case sprite && c == 13:
					return im.Palette[0], true
				}
				return im.Palette[c], true
			}))
		}
		return out
	}
	pal := l.AmigaPalette(file, block)
	ega := func(c byte) (color.RGBA, bool) { return pixelRGBA(c, pal), c < orig.Clear }
	if len(b) < 17 || (int(b[0])|int(b[1])<<8)*(int(b[2])|int(b[3])<<8)*int(b[8])*4 > len(b) { // ParseImage allocates what the header says
	} else if im, err := orig.ParseImage(b, false, 0); err == nil && im.Width > 0 && im.Count > 0 {
		for f := range im.Count {
			out = append(out, rgba(im, im.Frame(f), ega))
		}
		return out
	}
	if !saneSprite(b) {
		return out
	}
	for _, f := range orig.ParseSprite(b, false, 0, orig.XORed(file)) {
		if f.Image.Width == 0 {
			break
		}
		out = append(out, rgba(f.Image, f.Image.Frame(0), ega))
	}
	return out
}

// saneSprite reports whether animation block b's frame headers fit in it
// (ParseSprite allocates what they say: a script read as one took GBs).
func saneSprite(b []byte) bool {
	p := 1
	for f := 0; len(b) > 0 && f < int(b[0]); f++ {
		if p+21 > len(b) {
			return false
		}
		n := int(b[p+4]) | int(b[p+5])<<8
		n *= int(b[p+6]) | int(b[p+7])<<8
		px := n * 4 // two pixels a byte, an Amiga tagged frame's one (dax wide)
		if b[p+12] == 'A' && b[p+13] == 'M' && b[p+15] == 8 {
			px = n * 8
		}
		if px > len(b) {
			return false
		}
		p += 21 + px
	}
	return len(b) > 0
}
