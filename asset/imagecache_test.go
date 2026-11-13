package asset

import "testing"

// A sprite decodes once: the presenter asks for its portraits every View,
// and its GPU copies are kept by image identity.
func TestImageCached(t *testing.T) {
	n := SpriteNames("pool", "portraits")
	if len(n) == 0 {
		t.Skip("no pool portraits")
	}
	a, b := SpriteImage("pool", "portraits", n[0]), SpriteImage("pool", "portraits", n[0])
	if a == nil || a != b {
		t.Fatalf("Image decoded twice (%p, %p)", a, b)
	}
}

func BenchmarkImage(b *testing.B) {
	n := SpriteNames("pool", "portraits")
	for range b.N {
		SpriteImage("pool", "portraits", n[0])
	}
}

func TestImageAbsentTwice(t *testing.T) {
	if SpriteImage("pool", "portraits", "none") != nil || SpriteImage("pool", "portraits", "none") != nil {
		t.Fatal("absent sprite not nil")
	}
}
