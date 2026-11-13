package asset

import (
	"image"
	"image/color"
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

func TestDrawText(t *testing.T) {
	m := image.NewRGBA(image.Rect(0, 0, 16, 8))
	fg, bg := color.RGBA{255, 0, 0, 255}, color.RGBA{0, 0, 9, 255}
	DrawText(m, nil, "Ab", 0, 0, fg, &bg)
	for i, ch := range "AB" {
		g, _ := oracle(Get(GameFont), nil, byte(ch), true)
		for r := 0; r < 8; r++ {
			for b := 0; b < 8; b++ {
				want := bg
				if g[r]&(0x80>>b) != 0 {
					want = fg
				}
				if m.RGBAAt(i*8+b, r) != want {
					t.Fatal(i, r, b)
				}
			}
		}
	}
}

func TestChooser(t *testing.T) {
	cfg := t.TempDir() + "/FONT.CFG"
	c := NewChooser(t.TempDir(), "Dark_Queen_of_Krynn", cfg)
	if c.List[0].Name != "Dark_Queen_of_Krynn" || len(c.List) != 60 {
		t.Fatal("order")
	}
	c.Step(-1)
	if c.Cur != 60 {
		t.Fatal(c.Cur)
	}
	c.Set(2)
	if NewChooser(t.TempDir(), "Dark_Queen_of_Krynn", cfg).Cur != 2 {
		t.Fatal("cfg")
	}
}
