package playtest

import (
	"gobox/importer/games"
	"log"
	"os"
	"testing"
)

// TestMain runs the import a test's areas need (ImportIfStale), as the
// setup does: nothing at load decompiles.
func TestMain(m *testing.M) {
	if err := games.ImportIfStale("../../../data/por/sources/dos", "../../../data/por/sources/assets/scripts"); err != nil {
		log.Fatal(err)
	}
	if err := games.ImportAssetsIfStale("../../../data/por/sources"); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
