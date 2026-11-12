package pool_test

import (
	"log"
	"os"
	"testing"

	"por/importpor"
)

// TestMain runs the import a test's areas need (importpor.ImportIfStale), as the
// setup does: nothing at load decompiles. It is an external test file, so the
// package's tests can use the importer that imports the package.
func TestMain(m *testing.M) {
	if err := importpor.ImportIfStale("../../por-go/data/dos", "../../por-go/data/assets/scripts"); err != nil {
		log.Fatal(err)
	}
	if err := importpor.ImportAssetsIfStale("../../por-go/data"); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
