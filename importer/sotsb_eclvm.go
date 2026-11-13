package importer

import (
	"gobox/ecl"
)

// The ECL script machine: operands, variables, strings, comparisons, calls
// (unit O04C of the reconstruction). Script addresses are 16-bit:
//
//	19200..20223  area record words      (p4FB2)
//	31232..31743  aux record words       (p4FBA)
//	31744..32767  area2 record words     (p4FB6; some are character values)
//	32768..40447  the script code, bytes (p4FBE)
//	others        special values (direction, position, ...)
var sotsbEclLayout = ecl.AddrLayout{
	Area:  ecl.Range{First: 19200, Last: 20223},
	Aux:   ecl.Range{First: 31232, Last: 31743},
	Area2: ecl.Range{First: 31744, Last: 32767},
	Code:  ecl.Range{First: 32768, Last: 40447},
}

const (
	sotsbScriptCodeSize  = 40447 - 32768 + 1
	sotsbSpecialBase     = 49000 // writes above it: position and direction
	sotsbSpecialReadBase = 49227 // reads above it: position, direction, ...
	areaWordRemainder    = 831   // area2 word DIVIDE leaves its remainder in
)

// ScriptLayout is the game's script fields (scripts/layout/global.txt).
func SotsbScriptLayout() (*ecl.Layout, error) {
	return ecl.ReadLayout(embeddedScripts("sotsb"), 0)
}

// ImportScriptsTo decompiles the game's ECL*.DAX in dos into an adventure
// folder's scripts dir, with the script data beside them (the importer's
// step; ecl.ImportDAXFolder).
func SotsbImportScriptsTo(dos, dir string) error {
	return ecl.ImportDAXFolder(ecl.V2, dos, embeddedScripts("sotsb"), dir, journalFixes)
}
