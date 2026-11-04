package walls

import (
	"bytes"
	"image"
	"image/png"
	"testing"
	"testing/fstest"
)

// The lookups over a registered series (the real series test their files in their own modules).
func TestLookups(t *testing.T) {
	var b bytes.Buffer
	png.Encode(&b, image.NewGray(image.Rect(0, 0, 1, 1)))
	Register("test", fstest.MapFS{"floor_water.png": {Data: b.Bytes()}})
	if Names("test")[0] != "floor_water" || Image("test", "nope") != nil || Image("none", "floor_water") != nil ||
		!Has("test", "floor_water") || Has("test", "nope") || len(Textures("test", "floor_water", "nope")) != 1 {
		t.Error("lookups wrong")
	}
}

// Every series under Dir reads from disk, each texture decoding under a unique name.
func TestSeriesOnDisk(t *testing.T) {
	if len(Series()) < 2 || len(Names("pool")) == 0 || Image("buck-rogers", "floor_water") != nil || Has("../walls", "walls") {
		t.Fatalf("series %v", Series())
	}
	for _, s := range Series() {
		for _, n := range Names(s) {
			if Image(s, n) == nil {
				t.Errorf("%s/%s does not decode", s, n)
			}
		}
	}
}
