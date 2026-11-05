package combat

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"
)

func TestProps(t *testing.T) {
	fsys := fstest.MapFS{"props.json": {Data: []byte(`{"crown":{"mode":"extrude","trunk":["a","b"]},"b":{"mode":"extrude"},"pond":{"mode":"flat"},"x":{"mode":"mesh"}}`)}}
	names := []string{"", "crown", "a", "b", "pond", "x"}
	ps := readProps(fsys, names)
	if ps.modeOf(1) != "extrude" || ps.modeOf(4) != "flat" || ps.modeOf(5) != "" || ps.modeOf(2) != "" {
		t.Errorf("modes %q", ps.mode[:6])
	}
	if got := ps.trunkOf(1); len(got) != 2 || got[0] != 2 || got[1] != 3 {
		t.Errorf("trunks %v, want [2 3]", got)
	}
	if readProps(fstest.MapFS{}, names) != nil || readProps(nil, names) != nil {
		t.Error("no file must be no props")
	}
	// Height 0 crown (Buck Rogers) over listed trunk b: one tree on the crown's square, the trunk's square floor
	none := func(int) bool { return false }
	var g [25][50]Square
	g[3][4], g[4][4] = Square{Tile: 1, Wall: true}, Square{Tile: 3, Wall: true}
	k, foot := layout(&g, 0, none, none, ps)
	if k[3][4] != stand || k[4][4] != plain || foot[3][4] != 4 {
		t.Errorf("crown %v trunk %v foot %v, want stand, plain, 4", k[3][4], k[4][4], foot[3][4])
	}
	if k, _ := layout(&g, 0, none, none, nil); k[3][4] != flat {
		t.Errorf("unlisted Height 0 crown is %v, want flat (inferred)", k[3][4])
	}
	tiles := make([]image.Image, 6)
	for i := range tiles {
		im := image.NewRGBA(image.Rect(0, 0, tilePx, tilePx))
		for j := 3; j < len(im.Pix); j += 4 {
			im.Pix[j] = 255
		}
		tiles[i] = im
	}
	lo, hi := float32(9), float32(-9)
	for _, v := range stacked(tiles[1], tiles[3], tileUV(1), tileUV(3), 4, 3) {
		lo, hi = min(lo, v.Y), max(hi, v.Y)
	}
	if lo != 0 || hi != 2 {
		t.Errorf("stack spans %v..%v, want 0..2 (whole tree on the crown's square)", lo, hi)
	}
}

// A block tile's cube pictures come from blocks/: the side into its own
// slot, the top into a slot of its own after floorSlot; others get none.
func TestReadBlocks(t *testing.T) {
	var buf bytes.Buffer
	png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, tilePx, tilePx)))
	fsys := fstest.MapFS{
		"props.json":       {Data: []byte(`{"w":{"mode":"block"},"v":{"mode":"block"},"f":{"mode":"flat"}}`)},
		"blocks/w.png":     {Data: buf.Bytes()},
		"blocks/w_top.png": {Data: buf.Bytes()},
		"blocks/f.png":     {Data: buf.Bytes()},
	}
	names := []string{"f", "w", "v"}
	ps := readProps(fsys, names)
	got := readBlocks(fsys, names, ps)
	if len(got) != 2 || got[1] == nil || got[floorSlot+1] == nil || ps.top[1] != floorSlot+1 || ps.top[2] != 0 {
		t.Fatalf("slots %v, tops %v", got, ps.top[:3])
	}
}
