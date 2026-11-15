package rules1e

import (
	"gobox/rules"
	"log"
	"slices"
)

// ItemTypes is the size of the base item table (0x810 bytes).
const ItemTypes = 0x81

// monsterData is a monster's records (MON<n>CHA, SPC, ITM block id), as
// MON<n>.json holds them (adventure.Monster's shape).
type monsterData struct {
	Record  Sheet
	Affects []AffectRecord `json:",omitempty"`
	Items   []Item         `json:",omitempty"`
}

// tables are the adventure's tables rules1e reads, by file (the source's
// Name), read on first use.
type tables struct {
	src      rules.Source
	monsters map[string]map[int]monsterData
	items    map[string]map[int][]Item
}

func newTables(src rules.Source) *tables {
	return &tables{src: src, monsters: map[string]map[int]monsterData{}, items: map[string]map[int][]Item{}}
}

// monsterSet is the monsters of the set in play.
func (t *tables) monsterSet() map[int]monsterData {
	file := t.src.Name(rules.SourceMonsters)
	m, ok := t.monsters[file]
	if !ok {
		t.src.Read(rules.SourceMonsters, &m)
		t.monsters[file] = m
	}
	return m
}

// monster is monster id of the set in play; false: none.
func (t *tables) monster(id rules.MonsterID) (Monster, bool) {
	d, ok := t.monsterSet()[int(id)]
	return Monster{Sheet: d.Record, Affects: d.Affects, Items: d.Items}, ok
}

// firstMonster is the lowest monster id of the set in play; false: none.
func (t *tables) firstMonster() (int, bool) {
	first, ok := 0, false
	for id := range t.monsterSet() {
		if !ok || id < first {
			first, ok = id, true
		}
	}
	return first, ok
}

// itemBlock is the treasure archive's block n of the set in play, its records (nil: none).
func (t *tables) itemBlock(n int) []Item {
	file := t.src.Name(rules.SourceTreasure)
	m, ok := t.items[file]
	if !ok {
		t.src.Read(rules.SourceTreasure, &m)
		t.items[file] = m
	}
	return m[n]
}

// itemTable reads the base item table: a short or missing file leaves blank records.
func (t *tables) itemTable() []ItemBase {
	var items []ItemBase
	t.src.Read(rules.SourceItems, &items)
	if items == nil {
		// extra: without the table no class may use any item and nothing has stats
		log.Println("ITEMS (the base item table) is missing from the game data: import the game again")
		return make([]ItemBase, ItemTypes)
	}
	return slices.Clone(items)
}
