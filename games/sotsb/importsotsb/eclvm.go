package importsotsb

import (
	"embed"
	"gobox/dax"
	"gobox/ecl"
	eclscript "gobox/ecl/script"
	"io/fs"
)

// scriptFS is the scripts' data (gobox ecl/script Library.Data): the
// layout, made from sotsb-re's Pascal (scripts/layout/global.txt).
//
//go:embed scripts
var scriptFS embed.FS

func scriptData() fs.FS {
	d, err := fs.Sub(scriptFS, "scripts")
	if err != nil {
		panic(err) // a constant, valid path
	}
	return d
}

// The ECL script machine: operands, variables, strings, comparisons, calls
// (unit O04C of the reconstruction). Script addresses are 16-bit:
//
//	19200..20223  area record words      (p4FB2)
//	31232..31743  aux record words       (p4FBA)
//	31744..32767  area2 record words     (p4FB6; some are character values)
//	32768..40447  the script code, bytes (p4FBE)
//	others        special values (direction, position, ...)
var eclLayout = ecl.Layout{
	Area:  ecl.Range{First: 19200, Last: 20223},
	Aux:   ecl.Range{First: 31232, Last: 31743},
	Area2: ecl.Range{First: 31744, Last: 32767},
	Code:  ecl.Range{First: 32768, Last: 40447},
}

const (
	scriptCodeSize    = 40447 - 32768 + 1
	specialBase       = 49000 // writes above it: position and direction
	specialReadBase   = 49227 // reads above it: position, direction, ...
	areaWordRemainder = 831   // area2 word DIVIDE leaves its remainder in
)

// ScriptLayout is the game's script fields (scripts/layout/global.txt).
func ScriptLayout() (*eclscript.Layout, error) { return eclscript.ReadLayout(scriptData(), 0) }

// ImportScriptsTo decompiles the game's ECL*.DAX in dos into an adventure
// folder's scripts dir, with the script data beside them (the importer's
// step; eclscript.ImportFolder).
func ImportScriptsTo(dos, dir string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	return eclscript.ImportFolder(eclscript.V2, eclscript.DAXSource(l), scriptData(), dir, eclscript.FoldGates, journalFixes)
}
