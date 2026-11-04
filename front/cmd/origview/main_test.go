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
	dir := os.Getenv("HOME") + "/Projects/pod-go/data/dos"
	l, err := dax.NewLoader(dir)
	if err != nil || !l.Exists("GEO1.DAX") {
		t.Skip("no pod-go data")
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
