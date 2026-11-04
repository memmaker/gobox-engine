package classic

import (
	"image"
	"sort"
	"strconv"

	"gobox/front"
	"gobox/game"
)

// BodyTiles are the pool bodies of the weapon picker for a member that has no
// pool ids: the origins.json blocks of CBODY.DAX in the size class of its Icon
// "body" (32 blocks from the class base), by block. nil when it has pool ids
// or no original numbers.
func BodyTiles(s game.State) (names []string, blocks []int) {
	m := subject(s)
	o := iconOrigins()[GameID]
	if m.IconHeadID != "" || m.Icon == nil || o == nil {
		return nil, nil
	}
	for k := range o["CBODY.DAX"] {
		if b, err := strconv.Atoi(k); err == nil && b&^31 == m.Icon["body"]&^31 {
			blocks = append(blocks, b)
		}
	}
	sort.Ints(blocks)
	for _, b := range blocks {
		names = append(names, o["CBODY.DAX"][strconv.Itoa(b)])
	}
	return names, blocks
}

// TilePicker is page page of BodyTiles' picker: each tile the member's icon
// with that body, carrying the original ref (File CBODY, its block, frame 0).
func TilePicker(s game.State, page, cols, rows int) *front.Picker {
	names, blocks := BodyTiles(s)
	m := subject(s)
	head, _ := memberIcon(m)
	k := &game.Picker{Cell: image.Pt(3, 3), Pics: make([]game.Picture, len(names))}
	for i, n := range names {
		p := game.Picture{File: "CBODY", Block: blocks[i], Clear: -1}
		if pic, ok := poolIcon(head, n, m.IconColors, 0, false); ok && pic.Palette == nil {
			p.W, p.H, p.Pixels, p.Clear = pic.PixW, len(pic.Pix)/pic.PixW, pic.Pix, clearPix
		}
		k.Pics[i] = p
		if blocks[i] == m.Icon["body"] {
			k.Cur = i
		}
	}
	s.Picker = k
	return Picker(s, page, cols, rows)
}

// PickTile is tile i's pool body name ("": none).
func PickTile(s game.State, i int) string {
	names, _ := BodyTiles(s)
	if i < 0 || i >= len(names) {
		return ""
	}
	return names[i]
}
