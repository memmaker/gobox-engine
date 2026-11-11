package goldbox

import "testing"

type seeded struct{ s uint32 }

func (r *seeded) Random(n int) int { return Random(&r.s, n) }

// Pinned on the games' engines: seed 0x12345678, the first Random(20),
// Random(6), Random(6) are 19, 1, 0.
func TestRandomPinned(t *testing.T) {
	s := uint32(0x12345678)
	for i, c := range []struct{ n, want int }{{20, 19}, {6, 1}, {6, 0}} {
		if got := Random(&s, c.n); got != c.want {
			t.Fatalf("roll %d: Random(%d) = %d, want %d", i, c.n, got, c.want)
		}
	}
	if s != 0xd9a484b7 {
		t.Fatalf("seed %08x", s)
	}
	r := &seeded{0x12345678}
	if got := Dice(r, 1, 20); got != 20 {
		t.Fatalf("Dice(1,20) = %d, want 20", got)
	}
	fixed := uint32(7)
	Randomize(&s, &fixed)
	if s != 7 {
		t.Fatalf("Randomize fixed: %d", s)
	}
}

// Dice's n and sides are bytes in the originals: 257 dice are one die.
func TestDiceByteArgs(t *testing.T) {
	a, b := &seeded{0x12345678}, &seeded{0x12345678}
	if x, y := Dice(a, 257, 262), Dice(b, 1, 6); x != y || a.s != b.s {
		t.Fatalf("Dice(257,262) = %d, Dice(1,6) = %d", x, y)
	}
}
