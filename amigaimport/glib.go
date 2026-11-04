package amigaimport

import (
	"encoding/binary"
	"errors"
)

// GLIB is the container of the later SSI Amiga ports (Curse, Silver Blades,
// Pools of Darkness): .TLB (graphics) and .GLB (data) files. BE header:
//
//	"GLIB", u32 total size, u16 entry count, u8 packing, u8 has-index,
//	4-byte type ("GLIB" nested libraries, "HLIB" frame sets, "TILE"
//	pictures, "DATA"), then count+1 u32 entry offsets from the start.
//
// With has-index set, entry 0 is u16 n and n (u16 id, u16 entry) pairs;
// otherwise entry k has id k. A packed library (packing 1 copy, 3 LZW,
// 5 LZ77) stores a u16 and the packed stream after the 16-byte header;
// it unpacks to bytes 16..size. Unpacking runs in 0x2000-byte output
// chunks, each restarting the stream's bit alignment (from the games'
// executables: Curse code hunk 0x3d1b6, Silver Blades 0x417ae).
type entry struct {
	id   int
	data []byte
}

var errGLIB = errors.New("not a GLIB")

func be16(b []byte, i int) int { return int(binary.BigEndian.Uint16(b[i:])) }
func be32(b []byte, i int) int { return int(binary.BigEndian.Uint32(b[i:])) }

// glib returns a library's entries and its type, unpacked.
func glib(b []byte) (typ string, es []entry, err error) {
	defer func() {
		if recover() != nil {
			typ, es, err = "", nil, errGLIB
		}
	}()
	if len(b) < 16 || string(b[:4]) != "GLIB" {
		return "", nil, errGLIB
	}
	size, n := be32(b, 4), be16(b, 8)
	if b[10] != 0 {
		b = unpack(b, size, b[10])
		if b == nil {
			return "", nil, errGLIB
		}
	}
	off := func(i int) int { return be32(b, 16+4*i) }
	part := func(e int) []byte { return b[off(e):off(e+1)] }
	if b[11]&1 == 0 {
		for k := 0; k < n; k++ {
			es = append(es, entry{k, part(k)})
		}
		return string(b[12:16]), es, nil
	}
	idx := part(0)
	for k := 0; k < be16(idx, 0); k++ {
		es = append(es, entry{be16(idx, 2+4*k), part(be16(idx, 4+4*k))})
	}
	return string(b[12:16]), es, nil
}

// unpack returns the whole unpacked library (header copied), nil if the
// packing is unknown.
func unpack(b []byte, size int, method byte) []byte {
	out := make([]byte, size+0x2000) // a chunk may overrun its end, as in place
	copy(out, b[:16])
	src := b[18:]
	var dict []byte
	if method == 3 {
		dict = lzwDict()
	}
	for d := 16; d < size; d += 0x2000 {
		n := min(0x2000, size-d)
		switch method {
		case 1:
			copy(out[d:d+n], src)
			src = src[n:]
		case 3:
			src = lzw(out, d, d+n, src, dict)
		case 5:
			src = lz77(out, d, src)
		default:
			return nil
		}
	}
	return out[:size]
}

// lzwDict: flags (0 free, -1 leaf, 1 prefix) at 0, suffixes at 0x402,
// BE u16 prefixes at 0x802, next free code at 0x1002 (Curse 0x3d4f0).
func lzwDict() []byte {
	d := make([]byte, 0x1004)
	for i := range 0x100 {
		d[i] = 0xff
		d[0x402+i] = byte(i)
		d[0x802+2*i], d[0x803+2*i] = 0xff, 0xff
	}
	binary.BigEndian.PutUint16(d[0x1002:], 0x100)
	return d
}

// lzw decodes 10-bit MSB-first codes into out[q:end] (Curse 0x3d55e).
func lzw(out []byte, q, end int, src, d []byte) []byte {
	prefix := func(c int) int { return int(int16(binary.BigEndian.Uint16(d[0x802+2*c:]))) }
	var emit func(c int) byte
	emit = func(c int) byte {
		if c < 0x100 {
			out[q] = byte(c)
			q++
			return byte(c)
		}
		first := emit(prefix(c))
		out[q] = d[0x402+c]
		q++
		return first
	}
	buf := uint16(src[0])<<8 | uint16(src[1])
	src = src[2:]
	prev := int(buf >> 6)
	emit(prev)
	next, bits := be16(d, 0x1002), 6
	for q < end {
		if bits != 0 {
			buf = buf<<8 | uint16(src[0])
			src = src[1:]
			bits -= 2
		} else {
			buf = uint16(src[0])<<8 | uint16(src[1])
			src = src[2:]
			bits = 6
		}
		c := int(buf>>bits) & 0x3ff
		first := emit(c)
		if next >= 0x400 {
			next = prune(d, prev, next)
		}
		d[prev] = 1
		d[next] = 0xff
		d[0x402+next] = first
		binary.BigEndian.PutUint16(d[0x802+2*next:], uint16(prev))
		for next++; d[next] != 0; next++ {
		}
		prev = c
	}
	binary.BigEndian.PutUint16(d[0x1002:], uint16(next))
	return src
}

// prune frees leaf codes until at least 0x55 are free; returns the lowest
// (Curse 0x3d472).
func prune(d []byte, keep, low int) int {
	freed := 0
	for freed < 0x55 {
		for i := 0x100; i < 0x400; i++ {
			switch f := int8(d[i]); {
			case f < 0 && i != keep:
				d[i] = 0
				freed++
				low = min(low, i)
			case f > 0:
				d[i] = 0xff
			}
		}
		for i := 0x100; i < 0x400; i++ {
			if d[i] != 0 {
				p := int(binary.BigEndian.Uint16(d[0x802+2*i:]))
				d[p] = 1
			}
		}
	}
	return low
}

// lz77 decodes one chunk up to its end code (Silver Blades 0x454fc).
func lz77(out []byte, q int, src []byte) []byte {
	lit := func(n int) {
		copy(out[q:q+n], src[:n])
		q += n
		src = src[n:]
	}
	ref := func(off, n int) {
		for ; n > 0; n-- {
			out[q] = out[q-off-1]
			q++
		}
	}
	for {
		b := int(src[0])
		src = src[1:]
		switch {
		case b&0x80 == 0:
			c := int(src[0])
			src = src[1:]
			lit(c >> 3 & 3)
			ref(c&0xe0<<2|b, c&7+3)
		case b&0x40 == 0:
			c, e := int(src[0]), int(src[1])
			src = src[2:]
			lit(e & 3)
			ref(c&0xe0<<1|b&0x3f|e<<7&0x7e00, c&0x1f+3)
		case b&0x20 == 0:
			n := b & 0xf
			if b&0x10 == 0 {
				n = (n + 1) * 4
			}
			lit(n)
		default:
			return src
		}
	}
}
