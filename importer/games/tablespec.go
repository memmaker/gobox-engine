package games

import "embed"

// tableSpecs are the games' tables.json: which data tables the importer
// reads out of the executable into data/rules.json (tablespec).
//
//go:embed cotab/tables.json pod/tables.json sotsb/tables.json
var tableSpecs embed.FS

// tableSpec is game's tables.json.
func tableSpec(game string) []byte {
	b, err := tableSpecs.ReadFile(game + "/tables.json")
	if err != nil {
		panic(err) // embedded: a game named here without its file is a bug
	}
	return b
}

// PorAssets are the area scripts' names and layout (por/scripts, the
// decompiler's input, see MODERNIZED.md §6 in data/por/docs). They are built
// into the binary, so the import reads them from wherever it runs and an
// installation is complete without the repository.
//
//go:embed por/scripts
var PorAssets embed.FS
