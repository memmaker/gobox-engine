package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/rules"
	"path/filepath"
)

// sets are the data sets the engine reads (Set.file).
const sets = 8

// gameData is the imported GEO blocks read so far, by file (data/assets/data).
type gameData struct {
	dir  string
	maps map[string]map[int][]byte
}

// gdata is the engine's imported data, opened on first use.
func (e *Engine) gdata() *gameData {
	if e.gd == nil {
		e.gd = &gameData{dir: filepath.Join(e.adv, adventure.DataDir), maps: map[string]map[int][]byte{}}
	}
	return e.gd
}

// geoBlock is GEO archive file's block id (nil: none).
func (d *gameData) geoBlock(file string, id int) []byte {
	m, ok := d.maps[file]
	if !ok {
		m = readMaps(d.dir, file)
		d.maps[file] = m
	}
	return m[id]
}

// source is the engine's data folder as the ruleset's source (rules.Source).
type source struct{ e *Engine }

func (s source) RulesFile() string { return s.e.rulesFile() }

// Name is the file of table t in the set in play.
func (s source) Name(t rules.SourceTable) string {
	switch t {
	case rules.SourceMonsters:
		return s.e.monsterSet().File("MON")
	case rules.SourceTreasure:
		return s.e.set.File("ITEM")
	case rules.SourceItems:
		return adventure.ItemTableFile
	}
	panic(fmt.Sprintf("data: no table %d", t))
}

// Read reads the table's file into v; false: there is none. A broken file is
// a broken import: it panics.
func (s source) Read(t rules.SourceTable, v any) bool {
	name := s.Name(t)
	ok, err := adventure.ReadJSON(s.e.gdata().dir, name, v)
	if err != nil {
		panic(fmt.Sprintf("data %s: %v (run the import again)", name, err))
	}
	return ok
}

// readMaps reads dir/file.bin and .txt (nil: none); broken ones panic.
func readMaps(dir, file string) map[int][]byte {
	m, err := adventure.ReadBlocks(dir, file)
	if err != nil {
		panic(fmt.Sprintf("data %s: %v (run the import again)", file, err))
	}
	return m
}
