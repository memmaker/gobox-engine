// Package por holds what the port adds to the game: the area scripts' names
// and layout (scripts/, the decompiler's input, see MODERNIZED.md §6 in
// data/por/docs). They are built into the binary, so the import reads them
// from wherever it runs and an installation is complete without the repository.
package por

import "embed"

//go:embed scripts
var Assets embed.FS
