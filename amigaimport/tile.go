package amigaimport

import "encoding/binary"

// TILE pictures: BE u16 height, x, y; u8 width in bytes (8-pixel cells);
// u8 kind. Kind bit 0: masked, bit 1: packed; high nibble 0xe on 32-colour
// pictures, but the plane count follows from the sizes anyway. Kind 8
// (0xe8) is a palette: 16 or 32 RGB byte triplets.
//
//	0 plain:  P bitplanes one after the other
//	1 masked: a mask plane (1 = clear) then P bitplanes
//	2 packed: ByteRun1; with a high nibble of 0 the result is 4-bit chunky
//	  pixels (high nibble first), else bitplanes row by row (row y: plane
//	  0..P-1)
//	3 sparse: u16 A, u16 B; A bytes of row codes, then P planes of B bytes
//	  each. Codes: 00 next row, 01-3F skip n clear bytes, 4n n mask bytes
//	  follow (1 = opaque), 8n n opaque bytes sharing one stored byte,
//	  Cn n opaque bytes. Each opaque or masked byte takes the next byte of
//	  every plane.
//
// Checked against the DOS archives: Curse's CHEAD, CBODY, CPIC, HEAD,
// BODY, BIGPIC and TITLE decode to the DOS pixels under one colour map.
type tile struct {
	h, w, x, y int
	px         []byte // h*w*8 colour indices, clear for transparent
}

const clear = 0xff

func isPalette(d []byte) bool { return len(d) >= 8 && d[7]&0x0f == 8 }

// decodeTile returns nil for a palette or a malformed picture.
func decodeTile(d []byte) (t *tile) {
	defer func() {
		if recover() != nil {
			t = nil
		}
	}()
	if len(d) < 8 || isPalette(d) {
		return nil
	}
	t = &tile{h: be16(d, 0), x: int(int16(be16(d, 2))), y: int(int16(be16(d, 4))), w: int(d[6])}
	h, w := t.h, t.w
	t.px = make([]byte, h*w*8)
	if h*w == 0 {
		return t
	}
	body := d[8:]
	kind := d[7] & 3
	if kind == 3 {
		sparse(t, body)
		return t
	}
	if kind == 2 {
		body = byteRun(body)
		if d[7]&0xf0 == 0 { // 4-bit chunky, high nibble first
			for i := range t.px {
				t.px[i] = body[i/2] >> (4 * (1 - i%2)) & 15
			}
			return t
		}
	}
	mask := kind == 1
	planes := len(body) / (h * w)
	colours := planes
	if mask {
		colours--
	}
	for y := 0; y < h; y++ {
		row := body[y*planes*w:]
		for x := 0; x < w*8; x++ {
			bit := func(p int) bool {
				if kind != 2 {
					return body[p*h*w+y*w+x/8]&(0x80>>(x%8)) != 0
				}
				return row[p*w+x/8]&(0x80>>(x%8)) != 0
			}
			var c byte
			for p := 0; p < colours; p++ {
				pp := p
				if mask {
					pp++
				}
				if bit(pp) {
					c |= 1 << p
				}
			}
			if mask && bit(0) {
				c = clear
			}
			t.px[y*w*8+x] = c
		}
	}
	return t
}

func sparse(t *tile, b []byte) {
	a, n := be16(b, 0), be16(b, 2)
	codes, col := b[4:4+a], b[4+a:]
	planes := len(col) / n
	for i := range t.px {
		t.px[i] = clear
	}
	y, x, k := 0, 0, 0 // row, byte column, next colour byte
	put := func(m byte) {
		for bit := 0; bit < 8; bit++ {
			if m&(0x80>>bit) == 0 {
				continue
			}
			var c byte
			for p := 0; p < planes; p++ {
				if col[p*n+k]&(0x80>>bit) != 0 {
					c |= 1 << p
				}
			}
			t.px[y*t.w*8+x*8+bit] = c
		}
		x++
	}
	for i := 0; i < len(codes) && y < t.h; i++ {
		c := int(codes[i])
		cnt := c & 0x3f
		switch c >> 6 {
		case 0:
			if c == 0 {
				y, x = y+1, 0
			} else {
				x += cnt
			}
		case 1:
			for ; cnt > 0; cnt-- {
				i++
				put(codes[i])
				k++
			}
		case 2:
			for ; cnt > 0; cnt-- {
				put(0xff)
			}
			k++
		case 3:
			for ; cnt > 0; cnt-- {
				put(0xff)
				k++
			}
		}
	}
}

// byteRun unpacks ILBM ByteRun1 (n < 128: n+1 literals; n > 128: the next
// byte 257-n times; 128: nop).
func byteRun(s []byte) []byte {
	var o []byte
	for i := 0; i < len(s); {
		c := int(s[i])
		i++
		switch {
		case c < 128:
			o = append(o, s[i:min(len(s), i+c+1)]...)
			i += c + 1
		case c > 128:
			if i < len(s) {
				for k := 0; k < 257-c; k++ {
					o = append(o, s[i])
				}
			}
			i++
		}
	}
	return o
}

// tlbBlocks converts a .TLB library into DOS-layout blocks (first seen
// wins: the same archive may sit on several disks) and its palettes.
func tlbBlocks(data []byte, g Gfx, blocks, pals map[int][]byte) {
	_, es, err := glib(data)
	if err != nil {
		return
	}
	add := func(m map[int][]byte, id int, b []byte) {
		// ids past 255 (Curse HEAD 449, 549: disk 4's and 5's own block 49)
		// have no place in one DOS archive and are left out.
		if _, ok := m[id]; !ok && b != nil && id >= 0 && id < 256 {
			m[id] = b
		}
	}
	indexed := data[11]&1 != 0
	var all []*tile
	for k, e := range es {
		id := e.id
		if !indexed {
			id = g.Base + k
		}
		if g.Frames && !indexed {
			id = g.Base
		}
		var frames []*tile
		if len(e.data) >= 4 && string(e.data[:4]) == "GLIB" {
			_, in, _ := glib(e.data)
			for _, f := range in {
				if isPalette(f.data) {
					add(pals, id, f.data[8:])
				} else if t := decodeTile(f.data); t != nil {
					frames = append(frames, t)
				}
			}
		} else if isPalette(e.data) {
			add(pals, id, e.data[8:])
		} else if t := decodeTile(e.data); t != nil {
			frames = []*tile{t}
		}
		if g.Frames && !indexed {
			all = append(all, frames...)
		} else if len(frames) > 0 {
			add(blocks, id, encode(frames, g))
		}
	}
	if len(all) > 0 {
		add(blocks, g.Base, encode(all, g))
	}
}

// encode writes frames in the DOS picture (or, Sprite, animation) layout,
// tagged "AM" with the palette slot, one byte per pixel when any colour
// is past 15.
func encode(frames []*tile, g Gfx) []byte {
	frames = frames[:min(len(frames), 255)] // ponytail: a DOS block counts frames in a byte; PoD's 8X8D sets hold 256
	if g.XOR {                              // later frames are drawn over the first
		for _, f := range frames[1:] {
			if f.h != frames[0].h || f.w != frames[0].w {
				continue
			}
			for i, c := range f.px {
				if c == clear {
					f.px[i] = frames[0].px[i]
				}
			}
		}
	}
	wide := byte(0)
	for _, f := range frames {
		for _, c := range f.px {
			if c != clear && c > 15 {
				wide = 8
			}
		}
	}
	px := func(f *tile) []byte {
		out := make([]byte, 0, len(f.px))
		for i, c := range f.px {
			if c == clear {
				c = g.Mask
			}
			if wide != 0 {
				out = append(out, c)
			} else if i%2 == 0 {
				out = append(out, c<<4)
			} else {
				out[len(out)-1] |= c
			}
		}
		return out
	}
	le := binary.LittleEndian
	pos := func(v int) uint16 { return uint16(max(v, 0)) } // ponytail: negative Amiga offsets (SPRIT) dropped, their meaning unknown
	if g.Sprite {
		out := []byte{byte(len(frames))}
		for _, f := range frames {
			delay := uint32(0)
			if len(frames) > 1 {
				delay = 2 // ponytail: the libraries hold no timing; DOS PIC frames mostly wait 2 ticks
			}
			out = le.AppendUint32(out, delay)
			out = le.AppendUint16(out, uint16(f.h))
			out = le.AppendUint16(out, uint16(f.w))
			out = le.AppendUint16(out, pos(f.x))
			out = le.AppendUint16(out, pos(f.y))
			out = append(out, 'A', 'M', g.Slot, wide, 0, 0, 0, 0, 0)
			out = append(out, px(f)...)
		}
		return fits(out)
	}
	f0 := frames[0]
	var keep []*tile
	for _, f := range frames {
		if f.h == f0.h && f.w == f0.w {
			keep = append(keep, f)
		}
	}
	out := le.AppendUint16(nil, uint16(f0.h))
	out = le.AppendUint16(out, uint16(f0.w))
	out = le.AppendUint16(out, pos(f0.x))
	out = le.AppendUint16(out, pos(f0.y))
	out = append(out, byte(len(keep)), 'A', 'M', g.Slot, wide, 0, 0, 0, 0)
	for _, f := range keep {
		out = append(out, px(f)...)
	}
	return fits(out)
}

// fits drops a block too big for a DOS archive (16-bit sizes).
func fits(b []byte) []byte {
	if len(b) > 0xffff {
		return nil
	}
	return b
}
