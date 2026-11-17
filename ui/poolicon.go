package ui

import (
	"cmp"
	"image"
	"image/color"
	"maps"
	"slices"
	"strconv"
	"sync"

	"gobox/asset"
	"gobox/game"
)

// clearPix is the Pix index of a transparent pixel (front.IndexedRGBA leaves 17 out).
const clearPix = 17

// iconPartRows are the palette positions of the icon's six parts (the body, arm, leg,
// hair and face, shield, weapon): the low colour sits there, the high one eight above.
var iconPartRows = [6]int{1, 2, 3, 4, 6, 7}

// poolIcon is the combat icon of pool head and body names in pose (0 ready,
// 1 action), mirrored when flip: the body with the head OR'd over it by
// palette index among the sixteen colours, laid over it past them (a transparent pixel takes the other's), the six parts
// recoloured by colours (low nibble the low colour, high nibble the high one;
// fewer than six: as drawn; VGA art the same, by index). EGA art comes back as indices (Palette nil),
// other art as indices into its own Palette. ok is false when a picture is missing.
func poolIcon(head, body string, colours []byte, pose int, flip bool) (Picture, bool) {
	k := poolIconKey{Series, amigaShown(), head, body, string(colours), pose, flip}
	poolIcons.Lock()
	defer poolIcons.Unlock()
	if r, ok := poolIcons.m[k]; ok {
		return r.pic, r.ok
	}
	if poolIcons.m == nil || len(poolIcons.m) > 512 { // the icons a player's picks make; the art changing (F12, a new series) is in the key
		poolIcons.m = map[poolIconKey]poolIconResult{}
	}
	pic, ok := composeIcon(head, body, colours, pose, flip)
	poolIcons.m[k] = poolIconResult{pic, ok}
	return pic, ok
}

// poolIconKey is what poolIcon's picture is made of.
type poolIconKey struct {
	series     string
	amiga      bool
	head, body string
	colours    string
	pose       int
	flip       bool
}

type poolIconResult struct {
	pic Picture
	ok  bool
}

// poolIcons keeps the icons composed: a picture is composed once, not on every look at the sheet.
var poolIcons struct {
	sync.Mutex
	m map[poolIconKey]poolIconResult
}

// composeIcon is poolIcon, composed.
func composeIcon(head, body string, colours []byte, pose int, flip bool) (pic Picture, ok bool) {
	load := func(name string) *image.Paletted {
		if pose == 1 {
			if p := poolParts()[poolEntryID(name)].pose; p != "" {
				name = p
			}
		}
		im, _ := poolPicture("combat", name).(*image.Paletted)
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
		case v < 16 && d[i] < 16: // the sixteen colours' bit planes
			d[i] |= v
		default: // another palette's colours are no bit planes (OR'd they leave the palette): the head's pixel shows
			d[i] = v
		}
	}
	to := [16]uint8{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	if len(colours) >= 6 {
		for i, pos := range iconPartRows {
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
	pic = Picture{Pix: out, PixW: w}
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
func Amiga() bool { return amigaShown() }

// Series is game.Data.Series, the asset series whose shared pool the members' ids name ("pool" when unset).
var Series = "pool"

var iconOrigins = sync.OnceValue(asset.Origins)

// gameOrigins is the shown art's origins: <GameID>-amiga under F12's Amiga art, never the DOS set.
func gameOrigins() map[string]map[string]string {
	if Amiga() {
		return iconOrigins()[GameID+"-amiga"]
	}
	return iconOrigins()[GameID]
}

// memberIcon is m's combat icon head and body pool names: each its pool id,
// else its original CHEAD or CBODY number (Icon "head", "body") through
// origins.json (a member with only a head picked keeps its own body).
func memberIcon(m game.Member) (head, body string) {
	head, body = m.IconHeadID, m.IconBodyID
	if o := gameOrigins(); o != nil && m.Icon != nil {
		head = cmp.Or(head, o["CHEAD.DAX"][strconv.Itoa(m.Icon["head"])])
		body = cmp.Or(body, o["CBODY.DAX"][strconv.Itoa(m.Icon["body"])])
	}
	parts := poolParts()
	if hp, bp := parts[poolEntryID(head)], parts[poolEntryID(body)]; hp.large != bp.large { // a pool id is of one size: it follows the other part's (the game's Size moves only its own numbers)
		switch {
		case m.IconHeadID != "" && hp.other != "":
			head = hp.other
		case m.IconBodyID != "" && bp.other != "":
			body = bp.other
		}
	}
	return head, body
}

// poolPart is a pool combat head or body in its ready pose.
type poolPart struct {
	large bool   // its size class: the archives hold two figures, the large one 64 blocks on (IconSize 2); a head fits a body of its own
	pose  string // the action pose's pool name ("": none of its own)
	other string // the same part in the other size ("": none)
}

// poolParts are the ready-pose heads and bodies by pool name, from every
// game's origins: a CHEAD or CBODY block below 128 is a ready pose, 128
// blocks on its action pose. A name not in it is an action pose's picture.
var poolParts = sync.OnceValue(func() map[string]poolPart {
	parts := map[string]poolPart{}
	o := iconOrigins()
	for _, g := range slices.Sorted(maps.Keys(o)) { // the same part every run where games share a name
		for _, f := range []string{"CHEAD.DAX", "CBODY.DAX"} {
			for _, ref := range slices.Sorted(maps.Keys(o[g][f])) {
				b, err := strconv.Atoi(ref)
				if _, done := parts[o[g][f][ref]]; err != nil || b >= 128 || done {
					continue
				}
				parts[o[g][f][ref]] = poolPart{large: b&64 != 0, pose: o[g][f][strconv.Itoa(b+128)], other: o[g][f][strconv.Itoa(b^64)]}
			}
		}
	}
	return parts
})

// PickIcon is the subject's combat icon (ready pose) with head or body swapped
// for the named one ("": its own), for a picker cell.
func PickIcon(own game.Member, head, body string) image.Image {
	h, b := memberIcon(subject(own))
	if h == "" || b == "" {
		h, b = "combat_head_000", "combat_body_000"
	}
	if head != "" {
		h = head
	}
	if body != "" {
		b = body
	}
	return iconImage(h, b, subject(own).IconColors, 0)
}

// memberPoses are own's combat icon in its two poses (the pool's first
// while nothing is picked).
func memberPoses(own game.Member) (out [2]image.Image) {
	if own.IconPoses[0].File != "" { // the game's whole picture
		return wholePoses(own.IconPoses)
	}
	m := subject(own)
	h, b := memberIcon(m)
	if h == "" || b == "" {
		h, b = "combat_head_000", "combat_body_000"
	}
	for pose := range out {
		out[pose] = iconImage(h, b, m.IconColors, pose)
	}
	return out
}

// wholePoses are a game's whole-picture icon poses (Member.IconPoses, IconEditor.Poses) as images, from the DAX cache.
func wholePoses(poses [2]game.Picture) (out [2]image.Image) {
	for pose, pic := range poses {
		if pic.File != "" {
			out[pose] = picImage(drawnPicture(pic), egaInk)
		}
	}
	return out
}

// iconImage is the pool icon of head and body in pose, as an image (nil: none).
func iconImage(h, b string, colors []byte, pose int) image.Image {
	pic, ok := poolIcon(h, b, colors, pose, false)
	if !ok {
		return nil
	}
	if pic.Palette != nil {
		return PalettedRGBA(pic.Pix, pic.PixW, pic.Palette)
	}
	var ink [16]color.RGBA
	for i := range ink {
		ink[i] = asset.EGARGBA(i)
	}
	return IndexedRGBA(pic.Pix, pic.PixW, -1, &ink)
}
