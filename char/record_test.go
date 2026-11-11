package char

import (
	"bytes"
	"math/rand/v2"
	"testing"

	"gobox/core"
)

// RoundTrip checks that b decodes and encodes back to itself, blank fields aside.
func RoundTrip[R Record](t *testing.T, b []byte) {
	t.Helper()
	want := Kept[R](b)
	if got := Encode(Decode[R](b)); !bytes.Equal(got, want) {
		t.Fatalf("%T: round trip differs", *new(R))
	}
}

func TestRecordRoundTrip(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	fill := func(l core.RecordLayout) []byte {
		b := make([]byte, l)
		for i := range b {
			b[i] = byte(r.Uint32())
		}
		return b
	}
	RoundTrip[Record285](t, fill(Layout285))
	RoundTrip[Record422](t, fill(Layout422))
	RoundTrip[Record439](t, fill(Layout439))
	RoundTrip[Record510](t, fill(Layout510))
}
