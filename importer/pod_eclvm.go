package importer

import (
	"encoding/json"
	"gobox/asset/orig"
	"gobox/ecl"
	"io/fs"
	"testing/fstest"
)

// podScriptData is the scripts' data (gobox ecl Library.Data): the
// layout, made from pod-re's Pascal (scripts/layout/global.txt), and the
// journal's entries (JournalEntries, made from the book, not checked in) as
// EntriesFile. Every file is listed, so the import copies and keys the same
// files the engine reads from the folder.
func podScriptData() fs.FS {
	data := fstest.MapFS{}
	err := fs.WalkDir(embeddedScripts("pod"), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(embeddedScripts("pod"), p)
		data[p] = &fstest.MapFile{Data: b}
		return err
	})
	if err != nil {
		panic(err) // the embedded files
	}
	b, err := json.MarshalIndent(JournalEntries(), "", " ")
	if err != nil {
		panic(err) // plain data
	}
	data[ecl.EntriesFile] = &fstest.MapFile{Data: b}
	return data
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
var podEclLayout = ecl.AddrLayout{
	Area:  noRange,
	Aux:   noRange,
	Area2: noRange,
	Code:  ecl.Range{First: 32768, Last: 40447},
}

// noRange holds no address (Pools of Darkness has no area, aux or area2
// records).
var noRange = ecl.Range{First: 1, Last: 0}

const podScriptCodeSize = 40447 - 32768 + 1

// ImportScriptsTo decompiles the game's ECL*.DAX in dos into an adventure
// folder's scripts dir, with the script data beside them (the importer's
// step; orig.ImportDAXFolder).
func PodImportScriptsTo(dos, dir string) error {
	return orig.ImportDAXFolder(ecl.V3, dos, podScriptData(), dir, podScriptFixes)
}
