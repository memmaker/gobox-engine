package classic

import (
	"image"
	"image/color"
	"strconv"
	"sync"

	"gobox/front"
	"gobox/game"
	"gobox/sprites"
)

// clearPix is the Pix index of a transparent pixel (front.IndexedRGBA leaves 17 out).
const clearPix = 17

// iconParts are the palette positions of the icon's six parts (the body, arm, leg,
// hair and face, shield, weapon): the low colour sits there, the high one eight above.
var iconParts = [6]int{1, 2, 3, 4, 6, 7}

// poolIcon is the combat icon of pool head and body names in pose (0 ready,
// 1 action), mirrored when flip: the body with the head OR'd over it by
// palette index (a transparent pixel takes the other's), the six parts
// recoloured by colours (low nibble the low colour, high nibble the high one;
// fewer than six: as drawn; VGA art the same, by index). EGA art comes back as indices (Palette nil),
// other art as indices into its own Palette. ok is false when a picture is missing.
func poolIcon(head, body string, colours []byte, pose int, flip bool) (pic front.Picture, ok bool) {
	load := func(name string) *image.Paletted {
		if pose == 1 {
			if p := sprites.Pose(name); p != "" {
				name = p
			}
		}
		im, _ := sprites.Image("pool", "combat", name).(*image.Paletted)
		return im
	}
	h, b := load(head), load(body)
	if h == nil || b == nil {
		return pic, false
	}
	ega := len(b.Palette) == 17 // the 16 EGA colours, then transparent
	clear := func(im *image.Paletted, v uint8) bool {
		_, _, _, a := im.Palette[v].RGBA()
		return a == 0
	}
	w := b.Bounds().Dx()
	d := append([]uint8(nil), b.Pix...)
	for i, v := range h.Pix {
		switch {
		case i >= len(d) || clear(h, v):
		case clear(b, d[i]):
			d[i] = v
		default:
			d[i] |= v
		}
	}
	to := [16]uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	if len(colours) >= 6 {
		for i, pos := range iconParts {
			to[pos], to[pos+8] = colours[i]&15, colours[i]>>4
		}
	}
	out := make([]uint8, len(d))
	for i, v := range d {
		x, y := i%w, i/w
		if flip {
			x = w - 1 - x
		}
		switch {
		case ega && v >= 16:
			v = clearPix
		case ega, v < 16 && !clear(b, v): // VGA art (PoD) recolours the same sixteen indices, its transparent colour aside (R1C3)
			v = to[v]
		} // other art keeps its other indices: its palette has the alpha
		out[y*w+x] = v
	}
	pic = front.Picture{Pix: out, PixW: w}
	if !ega {
		for _, c := range b.Palette {
			r, g, bl, a := c.RGBA()
			pic.Palette = append(pic.Palette, color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(bl >> 8), uint8(a >> 8)})
		}
	}
	return pic, true
}

// GameID is game.Data.ID, the origins.json key that maps a member's original
// icon numbers to pool names when it has no pool ids (old saves, imports).
var GameID string

var iconOrigins = sync.OnceValue(sprites.Origins)

// memberIcon is m's combat icon head and body pool names: its pool ids, else
// its original CHEAD and CBODY numbers (Icon "head", "body") through origins.json.
func memberIcon(m game.Member) (head, body string) {
	if m.IconHeadID != "" {
		return m.IconHeadID, m.IconBodyID
	}
	o := iconOrigins()[GameID]
	if o == nil || m.Icon == nil {
		return "", ""
	}
	return o["CHEAD.DAX"][strconv.Itoa(m.Icon["head"])], o["CBODY.DAX"][strconv.Itoa(m.Icon["body"])]
}
