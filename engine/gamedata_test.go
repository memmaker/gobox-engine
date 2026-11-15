package engine

import (
	"gobox/adventure"
	"gobox/rules1e"
	"slices"
)

// The tests read the game's tables by file name, as the ruleset's source does by table.

// monster is block id of the MON archive file; false: none.
func (d *gameData) monster(file string, id int) (adventure.Monster[rules1e.Sheet, rules1e.AffectRecord, rules1e.Item], bool) {
	var m map[int]adventure.Monster[rules1e.Sheet, rules1e.AffectRecord, rules1e.Item]
	readJSON(d.dir, file, &m)
	r, ok := m[id]
	return r, ok
}

// itemBlock is ITEM archive file's block id, its records (nil: none).
func (d *gameData) itemBlock(file string, id int) []rules1e.Item {
	var m map[int][]rules1e.Item
	readJSON(d.dir, file, &m)
	return m[id]
}

// itemTable is the ITEMS table (nil: the game data has none).
func (d *gameData) itemTable() []rules1e.ItemBase {
	var t []rules1e.ItemBase
	readJSON(d.dir, adventure.ItemTableFile, &t)
	return slices.Clone(t)
}

// readJSON reads dir/name.json into v; a missing file leaves v as it is.
func readJSON(dir, name string, v any) {
	if _, err := adventure.ReadJSON(dir, name, v); err != nil {
		panic(err)
	}
}

// readItems is the base item table, blank when the data has none.
func (e *Engine) readItems() []rules1e.ItemBase {
	if t := e.gdata().itemTable(); t != nil {
		return t
	}
	return make([]rules1e.ItemBase, rules1e.ItemTypes)
}
