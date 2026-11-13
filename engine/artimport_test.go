package engine

import (
	"bytes"
	"path/filepath"
	"testing"

	"gobox/asset"
	"gobox/asset/orig"
	"gobox/rules"
)

// TestArtEqualsDAX: every block of the imported art reads back as the
// archives decode it (dax.ParseImage, ParseSprite), unmasked and masked on
// colours 0 and 13, DOS and Amiga; every block of the graphics archives the
// game draws (artFiles) is there.
func TestArtEqualsDAX(t *testing.T) {
	data := filepath.Join("..", "data", "por", "sources")
	l, err := orig.NewLoader(filepath.Join(data, DOSDir))
	if err != nil {
		t.Skip("no game files")
	}
	l.UseAmiga(filepath.Join(data, AmigaDir))
	art := asset.OpenArt(filepath.Join(data, AssetsDir, ArtDir))
	n := 0
	for _, f := range artFiles() {
		for _, v := range []asset.Version{asset.DOS, asset.Amiga} {
			for id := range 256 {
				if f.Name == "8X8D1" && id == fontBlock {
					continue
				}
				var b []byte
				if v == asset.Amiga {
					b = l.AmigaShared(f.Name, id)
				} else if df := l.File(f.Name); df != nil {
					b = df.BlockShared(id)
				}
				blk := art.Block(v, f.Name, id)
				if (b == nil) != (blk == nil) {
					t.Fatalf("%s %s %d: in the archive %v, in the art %v", v, f.Name, id, b != nil, blk != nil)
				}
				if b == nil {
					continue
				}
				n++
				for _, m := range []int{asset.NoMask, 0, 13} {
					masked := m != asset.NoMask
					if blk.Kind == asset.Still {
						want, err := orig.ParseImage(b, masked, byte(max(m, 0)))
						if err != nil {
							t.Fatalf("%s %s %d: %v", v, f.Name, id, err)
						}
						got := blk.Image(m)
						if !sameImage(want, got) {
							t.Fatalf("%s %s %d mask %d: differs", v, f.Name, id, m)
						}
						continue
					}
					want, got := orig.ParseSprite(b, masked, byte(max(m, 0)), orig.XORed(f.Name)), blk.Anim(m)
					if len(want) != len(got) {
						t.Fatalf("%s %s %d: %d frames, want %d", v, f.Name, id, len(got), len(want))
					}
					for i := range want {
						if want[i].Delay != got[i].Delay || !sameImage(want[i].Image, got[i].Image) {
							t.Fatalf("%s %s %d frame %d mask %d: differs", v, f.Name, id, i, m)
						}
					}
				}
			}
		}
	}
	t.Logf("%d blocks", n)
}

func sameImage(a, b *asset.Image) bool {
	return a.Width == b.Width && a.Height == b.Height && a.X == b.X && a.Y == b.Y && a.Count == b.Count && a.Mask == b.Mask && bytes.Equal(a.Data, b.Data)
}

// fontBlock is 8X8D1's font (text.go), no picture: gobox's fonts import has it.
const fontBlock = 201

// artFiles are the graphics archives the game draws from: the numbered
// sets' and the single ones.
func artFiles() []struct{ Name string } {
	var out []struct{ Name string }
	add := func(n string) { out = append(out, struct{ Name string }{n}) }
	for s := rules.Set(1); s <= 8; s++ {
		for _, f := range []string{"8X8D", "BODY", "CPIC", "HEAD", "PIC", "FINAL", "SPRIT"} {
			add(s.File(f))
		}
	}
	for _, f := range []string{"BACPAC", "CBODY", "CHEAD", "COMSPR", "DUNGCOM", "ICON", "RANDCOM", "SQRPACI", "TITLE", "WILDCOM"} {
		add(f)
	}
	return out
}
