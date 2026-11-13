package importpod

import (
	"embed"
	"encoding/json"
	"gobox/dax"
	"gobox/ecl"
	eclscript "gobox/ecl/script"
	"io/fs"
	"testing/fstest"
)

// scriptFS is the scripts' data (gobox ecl/script Library.Data): the
// layout, made from pod-re's Pascal (scripts/layout/global.txt).
//
//go:embed scripts
var scriptFS embed.FS

func scriptData() fs.FS {
	d, err := fs.Sub(scriptFS, "scripts")
	if err != nil {
		panic(err) // a constant, valid path
	}
	return withEntries{d}
}

// withEntries adds the journal's entries (JournalEntries, made from the
// book, not checked in) to the scripts' data as EntriesFile, the file the
// import writes: every reader of scriptData checks Journal against them.
type withEntries struct{ fs.FS }

func (w withEntries) Open(name string) (fs.File, error) {
	if name != eclscript.EntriesFile {
		return w.FS.Open(name)
	}
	b, err := json.Marshal(JournalEntries())
	if err != nil {
		return nil, err
	}
	return fstest.MapFS{name: {Data: b}}.Open(name)
}

// The ECL script machine: operands, variables, texts, calls (unit O055 of
// the reconstruction). Script addresses are 16-bit:
//
//	1..1024        game block bytes        (p87F8, game*)
//	1025..2048     the selected character's record bytes (p9CEE)
//	32768..40447   the script code, bytes  (p87FC)
//
// Every other address reads 0 and ignores writes. A variable operand says
// its size: type 1 a byte, 3 (and 128) a word.
var eclLayout = ecl.Layout{
	Area:  noRange,
	Aux:   noRange,
	Area2: noRange,
	Code:  ecl.Range{First: 32768, Last: 40447},
}

// noRange holds no address (Pools of Darkness has no area, aux or area2
// records).
var noRange = ecl.Range{First: 1, Last: 0}

const scriptCodeSize = 40447 - 32768 + 1

// ImportScriptsTo decompiles the game's ECL*.DAX in dos into an adventure
// folder's scripts dir, with the script data beside them (the importer's
// step; eclscript.ImportFolder).
func ImportScriptsTo(dos, dir string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	return eclscript.ImportFolder(eclscript.V3, eclscript.DAXSource(l), scriptData(), dir, eclscript.FoldGates, nil)
}
