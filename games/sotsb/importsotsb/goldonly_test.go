package importsotsb

import (
	"os"
	"path/filepath"
	"testing"

	eclscript "gobox/ecl/script"
)

// TestScriptsGoldOnly: no imported script reads, writes or compares a silver,
// electrum or platinum coin (Felix 2026-11-12: gold and copper only); the
// import rewrote them (eclscript.GoldCoins).
// Detect the class elsewhere: a layout field holding money has coins=, and each game runs this test.
func TestScriptsGoldOnly(t *testing.T) {
	dos := filepath.Join("..", "..", "..", "data", "sotsb", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	dir := t.TempDir()
	if err := ImportScriptsTo(dos, dir); err != nil {
		t.Fatal(err)
	}
	bad, err := eclscript.MovedCoins(eclscript.V2, dir, scriptData(), journalFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}
