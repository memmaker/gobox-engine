package main

import (
	"gobox/importer"
	"log"
	"os"
	"testing"
)

// TestMain runs the import a test's areas need (ImportIfStale), as the
// setup does: nothing at load decompiles.
func TestMain(m *testing.M) {
	if err := importer.ImportIfStale("../../data/por/sources/dos", "../../data/por/sources/assets/scripts"); err != nil {
		log.Fatal(err)
	}
	if err := importer.ImportAssetsIfStale("../../data/por/sources"); err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	os.RemoveAll(brAdvs()) // the br1/br2 tests' imported folders
	os.Exit(code)
}
