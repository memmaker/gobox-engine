package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/rules1e"
	"path/filepath"
)

// monsterData is a monster's records (MON<n>CHA, SPC, ITM block id).
type monsterData = adventure.Monster[rules1e.Sheet, rules1e.AffectRecord, rules1e.Item]

// sets are the data sets the engine reads (Set.file).
const sets = 8

// gameData is the imported data read so far, by file (data/assets/data).
type gameData struct {
	dir      string
	maps     map[string]map[int][]byte
	monsters map[string]map[int]monsterData
	items    map[string]map[int][]rules1e.Item
	table    []rules1e.ItemBase
}

// gdata is the engine's imported data, opened on first use.
func (e *Engine) gdata() *gameData {
	if e.gd == nil {
		e.gd = &gameData{dir: filepath.Join(e.adv, adventure.DataDir), maps: map[string]map[int][]byte{}, monsters: map[string]map[int]monsterData{}, items: map[string]map[int][]rules1e.Item{}}
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

// monster is block id of the MON archives of file ("MON<n>"); false: none.
func (d *gameData) monster(file string, id int) (monsterData, bool) {
	m, ok := d.monsters[file]
	if !ok {
		readJSON(d.dir, file, &m)
		d.monsters[file] = m
	}
	r, ok := m[id]
	return r, ok
}

// itemBlock is ITEM archive file's block id, its records (nil: none).
func (d *gameData) itemBlock(file string, id int) []rules1e.Item {
	m, ok := d.items[file]
	if !ok {
		readJSON(d.dir, file, &m)
		d.items[file] = m
	}
	return m[id]
}

// itemTable is the ITEMS table (nil: the game data has none).
func (d *gameData) itemTable() []rules1e.ItemBase {
	if d.table == nil {
		readJSON(d.dir, adventure.ItemTableFile, &d.table)
	}
	return d.table
}

// readJSON reads dir/name.json into v; a missing file leaves v as it is, a
// broken one is a broken import: it panics.
func readJSON(dir, name string, v any) {
	if _, err := adventure.ReadJSON(dir, name, v); err != nil {
		panic(fmt.Sprintf("data %s: %v (run the import again)", name, err))
	}
}

// readMaps reads dir/file.bin and .txt (nil: none); broken ones panic.
func readMaps(dir, file string) map[int][]byte {
	m, err := adventure.ReadBlocks(dir, file)
	if err != nil {
		panic(fmt.Sprintf("data %s: %v (run the import again)", file, err))
	}
	return m
}
