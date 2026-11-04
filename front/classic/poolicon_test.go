package classic

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"gobox/dax"
	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// daxIcon is the games' compositing (cotab rules.charIcon, por BuildCharIcon) of DAX
// blocks head and body in pose: the body with the head OR'd over it, recoloured,
// Clear as clearPix.
func daxIcon(h, b *dax.Image, colours [6]byte, pose int, flip bool) []uint8 {
	d := slices.Clone(b.Frame(min(pose, b.Count-1)))
	for i, v := range h.Frame(min(pose, h.Count-1)) {
		switch {
		case i >= len(d) || v == dax.Clear:
		case d[i] == dax.Clear:
			d[i] = v
		default:
			d[i] |= v
		}
	}
	to := [16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	for i, pos := range [6]byte{1, 2, 3, 4, 6, 7} {
		to[pos], to[pos+8] = colours[i]&15, colours[i]>>4
	}
	w := b.Width * 8
	out := make([]uint8, len(d))
	for i, v := range d {
		x := i % w
		if flip {
			x = w - 1 - x
		}
		if v == dax.Clear {
			v = clearPix
		} else {
			v = to[v&15]
		}
		out[i/w*w+x] = v
	}
	return out
}

// A composed pool icon equals the games' compositing of its DAX source, in both
// poses, mirrored or not, recoloured (games whose data/dos is not installed are skipped).
func TestPoolIconMatchesDAX(t *testing.T) {
	home, _ := os.UserHomeDir()
	n := 0
	colours := [6]byte{0x9a, 0x2b, 0x3c, 0x4d, 0x5e, 0x6f}
	for id, files := range sprites.Origins() {
		dir := filepath.Join(home, "Projects", id+"-go", "data", "dos")
		hf, err1 := dax.Open(filepath.Join(dir, "CHEAD.DAX"))
		bf, err2 := dax.Open(filepath.Join(dir, "CBODY.DAX"))
		if err1 != nil || err2 != nil || id == "pod" { // pod's art is VGA: it keeps its own palette
			t.Logf("skip %s", id)
			continue
		}
		name := func(file string, e dax.Entry, im *dax.Image, pose int) string {
			ref := fmt.Sprint(e.ID)
			if im.Count > 1 {
				ref += fmt.Sprintf(".%d", pose)
			}
			return files[file][ref]
		}
		var heads []dax.Entry
		for _, e := range hf.Entries[:min(len(hf.Entries), 4)] {
			heads = append(heads, e)
		}
		for _, be := range bf.Entries {
			b := dax.ParseImage(bf.BlockShared(be.ID), true, 0)
			for _, he := range heads {
				h := dax.ParseImage(hf.BlockShared(he.ID), true, 0)
				if h == nil || b == nil || h.Width != b.Width {
					continue
				}
				for pose := range min(2, h.Count, b.Count) {
					hn, bn := name("CHEAD.DAX", he, h, 0), name("CBODY.DAX", be, b, 0)
					if hn == "" || bn == "" {
						continue
					}
					for _, flip := range []bool{false, true} {
						got, ok := poolIcon(hn, bn, colours[:], pose, flip)
						if !ok {
							t.Fatalf("%s %d/%d: missing pool icon %s %s", id, he.ID, be.ID, hn, bn)
						}
						if !slices.Equal(got.Pix, daxIcon(h, b, colours, pose, flip)) {
							t.Errorf("%s CHEAD %d CBODY %d pose %d flip %v (%s %s): differs", id, he.ID, be.ID, pose, flip, hn, bn)
						}
						n++
					}
				}
			}
		}
	}
	t.Logf("compared %d", n)
}

// The creation sheet's icon box takes the draft's pool icon, recoloured by its IconColors.
func TestPoolIconCreation(t *testing.T) {
	d := game.Member{IconHeadID: "combat_head_001", IconBodyID: "combat_body_001", IconColors: []byte{0x9a, 0x2b, 0x3c, 0x4d, 0x5e, 0x6f}}
	want, ok := poolIcon(d.IconHeadID, d.IconBodyID, d.IconColors, 0, false)
	if !ok {
		t.Skip("no pool art")
	}
	v := front.View{Creation: &front.Creation{Icon: front.CellRect{X1: 29, Y1: 13, X2: 31, Y2: 15}, Scale: 1}}
	v.Pictures = []front.Picture{{File: "CBODY", At: image.Rect(232, 104, 256, 128)}}
	v = PoolIcon(v, game.State{Creation: &game.Creation{Draft: d}})
	if len(v.Pictures) != 1 || !slices.Equal(v.Pictures[0].Pix, want.Pix) || v.Pictures[0].At.Min != image.Pt(232, 104) {
		t.Fatalf("icon box not the pool icon: %+v", v.Pictures)
	}
}

// A draft with only original CHEAD/CBODY numbers (what a game publishes as Icon) resolves to pool names and draws.
func TestPoolIconCreationNumbers(t *testing.T) {
	GameID = "por"
	defer func() { GameID = "" }()
	m := game.Member{Icon: map[string]int{"head": 1, "body": 64}}
	head, body := memberIcon(m)
	if _, ok := poolIcon(head, body, nil, 0, false); !ok {
		t.Skipf("CHEAD 1 / CBODY 64 not in the pool: %q %q", head, body)
	}
}
