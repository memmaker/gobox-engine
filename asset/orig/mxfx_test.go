package orig

import (
	"encoding/binary"
	"testing"
)

// TestMXEffectsOnly: no song file (Curse's Amiga release), its GLIB
// effects still play.
func TestMXEffectsOnly(t *testing.T) {
	glb := make([]byte, 0x14+6+64)
	copy(glb, "GLIB")
	binary.BigEndian.PutUint16(glb[8:], 1)
	binary.BigEndian.PutUint32(glb[0x10:], 0x14)
	binary.BigEndian.PutUint16(glb[0x14:], 7418)
	binary.BigEndian.PutUint32(glb[0x16:], 64)
	for i := range 64 {
		glb[0x1a+i] = byte(int8(100 - (i%8)*25))
	}
	m, err := NewMX(nil, glb)
	if err != nil {
		t.Fatal(err)
	}
	m.Song(1)
	m.Effect(0)
	out := make([]byte, 4*2000)
	m.Render(out, 44100)
	loud := false
	for i := 0; i < len(out); i += 2 {
		if v := int16(binary.LittleEndian.Uint16(out[i:])); v > 200 || v < -200 {
			loud = true
		}
	}
	if !loud {
		t.Fatal("the effect is silent")
	}
}
