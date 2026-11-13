// Command importstale runs the imports a test's areas need where they are
// stale (importer.ImportIfStale), as the setup does.
// The engine's tests (gobox-engine/pool) run it from their TestMain, so the
// engine's module needs no import of por.
package main

import (
	"log"

	"gobox/importer"
)

func main() {
	if err := importer.ImportIfStale("../../data/por/sources/dos", "../../data/por/scripts"); err != nil {
		log.Fatal(err)
	}
}
