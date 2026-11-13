package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"gobox/dax"
)

// Pins one view (wall-compare POD-M1-01): Pools of Darkness GEO1 block 16,
// walls block 1, square 3,0 facing west.
func TestPinnedView(t *testing.T) {
	dir := "../../../data/pod/sources/dos"
	l, err := dax.NewLoader(dir)
	if err != nil || !l.Exists("GEO1.DAX") {
		t.Skip("no data/pod/sources/dos")
	}
	im, err := Render(l, mapKey{1, 16, [3]int{1, 127, 127}}, 3, 0, 6)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(im.Pix)); got != want {
		t.Fatalf("hash %s", got)
	}
}

const want = "046f2dfc2b665b1810621fad37d4416bd71d0d021fa70ecdfc75fd527da77d98"

// Pins a Buck Rogers view under the open sky (wall-compare BR-M9-01): the
// BACK4 block 10+128 backdrop behind GEO4 block 65, walls block 10, square 1,5 facing north.
func TestPinnedBuckRogersView(t *testing.T) {
	l, err := dax.NewLoader(os.Getenv("HOME") + "/Projects/buckrogers-go/data/dos")
	if err != nil || !l.Exists("BACK4.DAX") {
		t.Skip("no buckrogers-go data")
	}
	im, err := Render(l, mapKey{4, 65, [3]int{10, 127, 127}}, 1, 5, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := fmt.Sprintf("%x", sha256.Sum256(im.Pix)); got != wantBR {
		t.Fatalf("hash %s", got)
	}
}

const wantBR = "0d438267fbf0ecf9e083e3c35d5673070b4bbfd4462262fe77d5f34956ec3cd1"
