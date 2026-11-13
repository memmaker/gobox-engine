package games

import (
	"os"
	"testing"

	"gobox/ecl"
)

// TestScriptsGoldOnly: no imported script reads, writes or compares a silver,
// electrum or platinum coin (Felix 2026-11-12: gold and copper only); the
// import rewrote them (ecl.GoldCoins), the area 9 fixes the coin loop.
// Detect the class elsewhere: a layout field holding money has coins=, and each game runs this test.
func TestPorScriptsGoldOnly(t *testing.T) {
	const dir = "../../data/por/sources/assets/scripts"
	if err := ImportIfStale("../../data/por/sources/dos", dir); err != nil {
		t.Fatal(err)
	}
	bad, err := ecl.MovedCoins(ecl.V1, dir, os.DirFS(dir), porScriptFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}

// TestNoScalarTables: no imported script indexes a field of one element
// (ecl.ScalarTables: the quest table 4AA6-4ABF was 26 scalars, the clerk's
// QuestBandits[Scratch1] ran past it). Each game runs this test.
func TestPorNoScalarTables(t *testing.T) {
	dir := t.TempDir()
	if err := ImportIfStale("../../data/por/sources/dos", dir); err != nil {
		t.Fatal(err)
	}
	bad, err := ecl.ScalarTables(ecl.V1, dir, os.DirFS(dir), porScriptFixes)
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range bad {
		t.Error(b)
	}
}
