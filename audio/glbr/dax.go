package glbr

import (
	"encoding/binary"
	"errors"
)

// GlbrBlocks unpacks a DAX file into blocks indexed by id.
func GlbrBlocks(f []byte) ([][]byte, error) {
	if len(f) < 2 {
		return nil, errors.New("dax: short")
	}
	n := int(binary.LittleEndian.Uint16(f))
	var out [][]byte
	for i := 0; i+9 <= n && 2+i+9 <= len(f); i += 9 {
		e := f[2+i:]
		id, off := int(e[0]), int(binary.LittleEndian.Uint32(e[1:]))
		raw := int(binary.LittleEndian.Uint16(e[5:]))
		comp := int(binary.LittleEndian.Uint16(e[7:]))
		s := 2 + n + off
		if s+comp > len(f) {
			return nil, errors.New("dax: block outside file")
		}
		src, b := f[s:s+comp], make([]byte, 0, raw)
		for p := 0; p < len(src) && len(b) < raw; {
			c := int(src[p])
			p++
			if c < 0x80 {
				c++
				if p+c > len(src) {
					return nil, errors.New("dax: truncated")
				}
				b = append(b, src[p:p+c]...)
				p += c
			} else if p < len(src) {
				for range 256 - c {
					b = append(b, src[p])
				}
				p++
			}
		}
		for len(out) <= id {
			out = append(out, nil)
		}
		out[id] = b[:min(len(b), raw)]
	}
	return out, nil
}
