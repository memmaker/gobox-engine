package main

import (
	"gobox/importer"
	"gobox/ui"
	"log"
	"os"
	"slices"
	"testing"
)

// TestMain runs the import a test's areas need (ImportIfStale), as the
// setup does: nothing at load decompiles.
func TestMain(m *testing.M) {
	if err := importer.ImportIfStale("../../data/por/sources/dos", "../../data/por/scripts"); err != nil {
		log.Fatal(err)
	}
	var stuck []string // the dead screens the run's frames drew, each once: a question of the game's on no screen
	ui.OnStuck = func(what string) {
		if !slices.Contains(stuck, what) {
			stuck = append(stuck, what)
		}
	}
	code := m.Run()
	for _, s := range stuck {
		log.Println("FAIL: a frame drawn stuck:", s)
	}
	if len(stuck) > 0 && code == 0 {
		code = 1
	}
	os.Exit(code)
}
