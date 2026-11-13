// Command importstale runs the imports a test's areas need where they are
// stale (games.ImportIfStale, ImportAssetsIfStale), as the setup does.
// The engine's tests (gobox-engine/pool) run it from their TestMain, so the
// engine's module needs no import of por.
package main

import (
	"log"

	"gobox/importer/games"
)

func main() {
	if err := games.ImportIfStale("../../data/por/sources/dos", "../../data/por/sources/assets/scripts"); err != nil {
		log.Fatal(err)
	}
	if err := games.ImportAssetsIfStale("../../data/por/sources"); err != nil {
		log.Fatal(err)
	}
}
