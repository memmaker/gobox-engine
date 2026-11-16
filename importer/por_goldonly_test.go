package importer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gobox/ecl"
)

// TestScriptsGoldOnly: no imported script reads, writes or compares a silver,
// electrum or platinum coin (Felix 2026-11-12: gold and copper only); the
// import rewrote them (ecl.GoldCoins), the area 9 fixes the coin loop.
// Detect the class elsewhere: a layout field holding money has coins=, and each game runs this test.
func TestPorScriptsGoldOnly(t *testing.T) {
	const dir = "../data/por/scripts"
	if err := ImportIfStale("../data/por/sources/dos", dir); err != nil {
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
	if err := ImportIfStale("../data/por/sources/dos", dir); err != nil {
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

// POR's overland blocks are its data (eclvm.pas WILD_*_SCRIPT 25..27), not the engine's.
func TestPorOverlandTables(t *testing.T) {
	for b, mode := range map[int]int{25: 2, 26: 3, 27: 4} {
		if o := porGameTables.OverlandBlock(b); o == nil || o.ViewMode != mode {
			t.Errorf("block %d: %v", b, o)
		}
	}
	if !porGameTables.NoRemains(10) {
		t.Error("block 10 has no remains")
	}
}

// The engine no longer rewrites the platinum prices: the imported text says them
// (the old replacer's inputs are gone from POR's scripts, its outputs are there).
func TestPorPlatinumPricesInScripts(t *testing.T) {
	const dir = "../data/por/scripts"
	if err := ImportIfStale("../data/por/sources/dos", dir); err != nil {
		t.Fatal(err)
	}
	all := ""
	for _, a := range []string{"area0.txt", "area11.txt"} {
		src, err := os.ReadFile(filepath.Join(dir, a))
		if err != nil {
			t.Fatal(err)
		}
		all += string(src)
	}
	for _, gone := range []string{"1 PLATINUM PIECE", "ENOUGH PLATINUM", "AND FIGHTERS AND 'ROGUES' TO GO EAST."} {
		if strings.Contains(all, gone) {
			t.Errorf("%q still in the scripts", gone)
		}
	}
	for _, want := range []string{"5 GOLD PIECES", "ENOUGH GOLD", "AND FIGHTERS, PALADINS, RANGERS AND 'ROGUES' TO GO EAST."} {
		if !strings.Contains(all, want) {
			t.Errorf("%q missing from the scripts", want)
		}
	}
}
