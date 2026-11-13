package asset

import (
	"testing"
)

func upperOracle(c byte) byte {
	if c >= 'a' && c <= 'z' {
		return c - 32
	}
	return c
}

// oracle is por-go's engine glyph() before the move (font = chosen pool font bits, nil: the game's).
func oracle(game, pc []byte, ch byte, upcase bool) ([8]byte, bool) {
	var g [8]byte
	if pc != nil && upcase && ch >= 0x20 && ch < 0x7F {
		copy(g[:], pc[int(ch)*8:])
		return g, true
	}
	i := int(ch)
	if upcase {
		switch ch {
		case '[':
			ch = '<'
		case ']':
			ch = '>'
		}
		i = int(upperOracle(ch)) % 0x40
	}
	if game == nil || i >= 177 {
		return g, false
	}
	copy(g[:], game[(0x40+i)*8:])
	return g, true
}

func TestGlyphOracle(t *testing.T) {
	game := Get(GameFont)
	for _, pc := range [][]byte{nil, Get("IBM_BIOS"), Get("Dark_Queen_of_Krynn")} {
		for c := 0; c < 256; c++ {
			for _, up := range []bool{false, true} {
				wg, wok := oracle(game, pc, byte(c), up)
				g, ok := Glyph(pc, byte(c), up)
				if g != wg || ok != wok {
					t.Fatalf("ch %d up %v", c, up)
				}
			}
		}
	}
}
