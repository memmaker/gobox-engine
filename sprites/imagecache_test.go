package sprites

import "testing"

// A sprite decodes once: the presenter asks for its portraits every View,
// and its GPU copies are kept by image identity.
func TestImageCached(t *testing.T) {
	n := Names("pool", "portraits")
	if len(n) == 0 {
		t.Skip("no pool portraits")
	}
	a, b := Image("pool", "portraits", n[0]), Image("pool", "portraits", n[0])
	if a == nil || a != b {
		t.Fatalf("Image decoded twice (%p, %p)", a, b)
	}
}

func BenchmarkImage(b *testing.B) {
	n := Names("pool", "portraits")
	for range b.N {
		Image("pool", "portraits", n[0])
	}
}

func TestImageAbsentTwice(t *testing.T) {
	if Image("pool", "portraits", "none") != nil || Image("pool", "portraits", "none") != nil {
		t.Fatal("absent sprite not nil")
	}
}
