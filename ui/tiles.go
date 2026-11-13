package ui

import (
	"image"
	"sort"
	"strconv"

	"gobox/game"
)

// BodyTiles are the pool bodies of the weapon picker for a member that has no
// pool ids: the origins.json blocks of CBODY.DAX in the size class of its Icon
// "body" (32 blocks from the class base), by block. nil when it has pool ids
// or no original numbers.
func BodyTiles(own game.Member) (names []string, blocks []int) {
	m := subject(own)
	o := gameOrigins()
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
func TilePicker(own game.Member, cells image.Point, page, cols, rows int) *FrontPicker {
	names, blocks := BodyTiles(own)
	m := subject(own)
	head, _ := memberIcon(m)
	k := &game.Picker{Pics: make([]game.Picture, len(names))}
	for i, n := range names {
		p := game.Picture{File: "CBODY", Block: blocks[i], Clear: -1, W: 24, H: 24} // a 3 by 3 cell tile
		if pic, ok := poolIcon(head, n, m.IconColors, 0, false); ok && pic.Palette == nil {
			p.W, p.H, p.Pixels, p.Clear = pic.PixW, len(pic.Pix)/pic.PixW, pic.Pix, clearPix
		}
		k.Pics[i] = p
		if blocks[i] == m.Icon["body"] {
			k.Cur = i
		}
	}
	return Picker(k, cells, page, cols, rows)
}

// PickTile is tile i's pool body name ("": none).
func PickTile(own game.Member, i int) string {
	names, _ := BodyTiles(own)
	if i < 0 || i >= len(names) {
		return ""
	}
	return names[i]
}
