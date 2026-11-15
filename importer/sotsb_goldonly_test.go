package importer

import (
	"os"
	"path/filepath"
	"testing"

	"gobox/ecl"
)

// TestScriptsGoldOnly: no imported script reads, writes or compares a silver,
// electrum or platinum coin (Felix 2026-11-12: gold and copper only); the
// import rewrote them (ecl.GoldCoins).
// Detect the class elsewhere: a layout field holding money has coins=, and each game runs this test.
func TestSotsbScriptsGoldOnly(t *testing.T) {
	dos := filepath.Join("..", "data", "sotsb", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	dir := t.TempDir()
	if err := SotsbImportScriptsTo(dos, dir); err != nil {
		t.Fatal(err)
	}
	bad, err := ecl.MovedCoins(ecl.V2, dir, embeddedScripts("sotsb"), journalFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}

// TestNoScalarTables: no imported script indexes a field of one element
// (ecl.ScalarTables; POR's quest table was split into scalars).
func TestSotsbNoScalarTables(t *testing.T) {
	dos := filepath.Join("..", "data", "sotsb", "sources", "dos")
	if _, err := os.Stat(dos); err != nil {
		t.Skip("no originals:", err)
	}
	dir := t.TempDir()
	if err := SotsbImportScriptsTo(dos, dir); err != nil {
		t.Fatal(err)
	}
	bad, err := ecl.ScalarTables(ecl.V2, dir, embeddedScripts("sotsb"), journalFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}
