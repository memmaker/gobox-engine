package importpor

import (
	"os"
	"testing"

	"gobox/ecl/script"
)

// TestScriptsGoldOnly: no imported script reads, writes or compares a silver,
// electrum or platinum coin (Felix 2026-11-12: gold and copper only); the
// import rewrote them (script.GoldCoins), the area 9 fixes the coin loop.
// Detect the class elsewhere: a layout field holding money has coins=, and each game runs this test.
func TestScriptsGoldOnly(t *testing.T) {
	const dir = "../../../data/por/sources/assets/scripts"
	if err := ImportIfStale("../../../data/por/sources/dos", dir); err != nil {
		t.Fatal(err)
	}
	bad, err := script.MovedCoins(script.V1, dir, os.DirFS(dir), scriptFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}
