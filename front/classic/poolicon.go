package classic

import (
	"image"
	"image/color"
	"slices"
	"strconv"
	"sync"

	"gobox/ega"
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
			if p := sprites.Pose(Series, name); p != "" {
				name = p
			}
		}
		im, _ := poolImage("combat", name).(*image.Paletted)
		return im
	}
	h, b := load(head), load(body)
	if h == nil || b == nil {
		return pic, false
	}
	ega := len(b.Palette) == 17 // the 16 EGA colours, then transparent
	clear := func(im *image.Paletted, v uint8) bool {
		_, _, _, a := im.Palette[v].RGBA()
		return a == 0 || len(im.Palette) != 17 && len(im.Pix) > 0 && v == im.Pix[0] // VGA art (PoD) keys its opaque background, the corner's colour
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
	var hole uint8 // VGA art: the keyed background drawn as the palette's transparent entry
	for i, c := range b.Palette {
		if _, _, _, a := c.RGBA(); a == 0 {
			hole = uint8(i)
			break
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
		case !ega && clear(b, v):
			v = hole
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

// Amiga reports whether the Amiga art shows now (F12): icons then come from <GameID>-amiga only.
var Amiga = func() bool { return false }

// Series is game.Data.Series, the asset series whose shared pool the members' ids name ("pool" when unset).
var Series = "pool"

var iconOrigins = sync.OnceValue(sprites.Origins)

// gameOrigins is the shown art's origins: <GameID>-amiga under F12's Amiga art, never the DOS set.
func gameOrigins() map[string]map[string]string {
	if Amiga() {
		return iconOrigins()[GameID+"-amiga"]
	}
	return iconOrigins()[GameID]
}

// memberIcon is m's combat icon head and body pool names: its pool ids, else
// its original CHEAD and CBODY numbers (Icon "head", "body") through origins.json.
func memberIcon(m game.Member) (head, body string) {
	if m.IconHeadID != "" {
		return m.IconHeadID, m.IconBodyID
	}
	o := gameOrigins()
	if o == nil || m.Icon == nil {
		return "", ""
	}
	return o["CHEAD.DAX"][strconv.Itoa(m.Icon["head"])], o["CBODY.DAX"][strconv.Itoa(m.Icon["body"])]
}

// PoolIcon is v with the creation draft's combat icon (pool ids, else its Icon numbers) in the
// sheet's Icon box, in place of the game's picture there (v itself when the draft has none).
func PoolIcon(v front.View, s game.State) front.View {
	cr, m := v.Creation, subject(s)
	if cr == nil {
		return v
	}
	if p := s.SheetPage; p != nil && p.Icons[0] != nil { // a dropped PNG: the sheet draws it
		return v
	}
	box := cr.Icon.Pixels()
	head, body := memberIcon(m)
	composed := m.IconHeadID == "" && head != ""
	if head == "" || body == "" {
		if slices.ContainsFunc(v.Pictures, func(pc front.Picture) bool { return !pc.Fill && pc.At.Overlaps(box) }) {
			return v // the game's own icon
		}
		head, body = "combat_head_000", "combat_body_000" // nothing picked yet: the box still shows an icon, the pool's first
	}
	var keep []front.Picture
	for _, pc := range v.Pictures {
		if !pc.Fill && pc.At.Overlaps(box) {
			continue
		}
		keep = append(keep, pc)
	}
	for pose := range 2 {
		pic, ok := poolIcon(head, body, m.IconColors, pose, false)
		if !ok {
			return v
		}
		if composed { // composed from the original numbers: drawn from CBODY's block (+128: the action pose)
			pic.File, pic.Block = "CBODY", m.Icon["body"]+128*pose
		}
		x := box.Min.X + 40*max(1, cr.Scale)*pose // the poses 5 cells apart
		pic.Name = head
		pic.At = image.Rectangle{image.Pt(x, box.Min.Y), image.Pt(x+pic.PixW, box.Min.Y+len(pic.Pix)/pic.PixW)}
		keep = append(keep, pic)
	}
	v.Pictures = keep
	return v
}

// PickIcon is the subject's combat icon (ready pose) with head or body swapped
// for the named one ("": its own), for a picker cell.
func PickIcon(s game.State, head, body string) image.Image {
	h, b := memberIcon(subject(s))
	if h == "" || b == "" {
		h, b = "combat_head_000", "combat_body_000"
	}
	if head != "" {
		h = head
	}
	if body != "" {
		b = body
	}
	pic, ok := poolIcon(h, b, subject(s).IconColors, 0, false)
	if !ok {
		return nil
	}
	if pic.Palette != nil {
		return front.PalettedRGBA(pic.Pix, pic.PixW, pic.Palette)
	}
	var ink [16]color.RGBA
	for i := range ink {
		ink[i] = ega.RGBA(i)
	}
	return front.IndexedRGBA(pic.Pix, pic.PixW, -1, &ink)
}
