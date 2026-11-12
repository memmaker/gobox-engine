package pool

import (
	"fmt"
	"gobox/char"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/ecl/script"
	"gobox/exe/tablespec"
	"os"
	"path/filepath"
	"slices"
)

// Ruleset builds the Engine's ruleset (the one injection point of the rules,
// fold step E2): monster, items and treasure are the game data's monster
// records, base item table and treasure blocks; rules is the adventure's
// rules.json ("": it has none); host is the Engine's side of the creation. The game hands it to New; the types are unnamed so a builder
// needs nothing of this package.
type Ruleset func(
	rules string,
	monster func(core.MonsterID) (char.Sheet, []char.AffectRecord, []char.Item, bool),
	items func() []char.ItemBase,
	treasure func(block int) []char.Item,
	host goldbox.CreationHost,
) (core.Rules, core.CreationRules)

// GameTables is the game's data the Engine shows, handed in by the game at New.
type GameTables struct {
	PortraitHead, PortraitBody []goldbox.PortraitPic // portrait number -> its picture, 1-based ([0] unused)
	Script                     *script.Version       `json:"-"` // the game's ECL dialect; nil: the first engine's (V1)
	ScriptSounds               map[byte]ScriptSound  // the Sound command's numbers (pod: 2..14 digitized, 128+n song n); one not listed plays DefaultSound
	DefaultSound               *ScriptSound          // what a number ScriptSounds does not list plays; nil: nothing
	AddNPCIcon                 bool                  // AddNPC loads the NPC's combat icon (sotsb, pod; not cotab)
	AreaMonsters               AreaMonsters          // the area's own monsters a fight adds (pod); zero: none
	Dialect                    core.Dialect          // the game's script byte-code dialect (ReadGameTables sets Script from it)
	TreasureSets               [][]int32             // ItemSet's sets: set number -> its library item ids (pod: 1..11; [0] unused)
	ScriptFixes                []script.Fix          // the game's decompiler special cases (script.Library.Fixes): the import's, for the texts' key
	Natives                    []NativeAt            // the labels a native script runs at (POR: the tavern game, CAST KNOCK); none: no native
	Quests                     []QuestDef            // the quest journal's quests (POR: the commissions); none: an empty journal
	OriginalSave               char.SaveLayout       // the game's SAVGAMx.DAT format (the original-save import)
	PurseWhos                  []PurseWho            // the Who commands that spend the party's one purse (POR: the inn, the tavern)
	Symbols                    []SymbolLoad          // the 8x8 symbol sets Init loads; nil: POR's, sotsb's and cotab's (PorSymbols)
	Art                        *ArtNames             // the pictures the Engine loads by role; nil: POR's (PorArt)
	MapLabels                  []MapLabel            // the maps' fixed labels (POR: New Phlan's buildings, PorMapLabels); none: none
	Music                      *AreaMusic            // the soundtrack by area (POR: PorMusic); nil: none
	MonsterSet                 goldbox.Set           // the set whose MON and CPIC archives hold the monsters (sotsb, pod: 1); 0: the current set (POR, cotab: the disk)
}

// ArtNames are the picture files the Engine loads by role, each game's own
// (POR's SQRPACI tile set is cotab's TILES; sotsb and pod have none). An
// empty name or list: the game has no such picture.
type ArtNames struct {
	TileSheets    []TileLoad // whole sheets Init takes as the tile set, tile First/0x80's half (POR: SQRPACI 1 and 2); none: a blank set
	Tiles         []TileLoad // laid over the tile set at Init (cotab: TILES 1 and 2, copyprot's Load24x24Set)
	DungeonTiles  []TileLoad // what combat setup lays over the tiles in a dungeon (ovr11:1244, LoadCombatTileSet)
	WildTiles     []TileLoad // and outdoors (sotsb: none, the set stays as it is)
	AreaTiles     []TileLoad // the area map's (BACPAC), laid over the tiles off the 3D view
	AreaIcon      string     // the area map's party icon sheet (ICON)
	CombatSprites string     // the combat sprites (COMSPR)
	Title         string     // the title pictures (TITLE)
	BorderCells   []int      // DrawBorder style 0's corner, side and top symbols (POR: 0x114..0x116); none: no border pictures
}

// TileLoad is a block of File laid into the tile set from tile First: its
// frames 0..Last (LoadBackpacIcons' n, Load24x24Set's count-1). The block is
// Block, or for From a value of the game's state.
type TileLoad struct {
	File               string
	Block, First, Last int
	From               TileFrom `json:",omitempty"`
}

// TileFrom is where a TileLoad's block number comes from.
type TileFrom int

const (
	TileBlockFixed   TileFrom = iota // TileLoad.Block
	TileBlockWalls                   // the low byte of the loaded wall set's id (pod's DungCom)
	TileBlockAreaSet                 // the area's tile set when 1 or 2, else 0 (pod's WildCom)
)

// PorArt is POR's ArtNames (Init, ovr11:1244, ovr33:0022).
var PorArt = ArtNames{
	TileSheets:    []TileLoad{{File: "SQRPACI", Block: 1}, {File: "SQRPACI", Block: 2, First: 0x80}},
	DungeonTiles:  []TileLoad{{File: "DUNGCOM", Block: 1, Last: 0x18}, {File: "RANDCOM", Block: 1, First: 0x22, Last: 5}},
	WildTiles:     []TileLoad{{File: "WILDCOM", Block: 1, Last: 0x21}, {File: "RANDCOM", Block: 1, First: 0x22, Last: 5}},
	AreaTiles:     []TileLoad{{File: "BACPAC", Block: 1, Last: 0x28}},
	AreaIcon:      "ICON",
	CombatSprites: "COMSPR",
	Title:         "TITLE",
	BorderCells:   []int{0x114, 0x115, 0x116},
}

// Files are every file a names (TestArtNamesImported).
func (a ArtNames) Files() []string {
	var out []string
	for _, l := range slices.Concat(a.TileSheets, a.Tiles, a.DungeonTiles, a.WildTiles, a.AreaTiles) {
		out = append(out, l.File)
	}
	for _, n := range []string{a.AreaIcon, a.CombatSprites, a.Title} {
		if n != "" {
			out = append(out, n)
		}
	}
	return out
}

// monsterSet is the set the monsters' MON and CPIC archives are read from.
func (e *Engine) monsterSet() goldbox.Set {
	if s := e.game().MonsterSet; s != 0 {
		return s
	}
	return e.set
}

// artNames is the game's ArtNames.
func (e *Engine) artNames() ArtNames {
	if a := e.game().Art; a != nil {
		return *a
	}
	return PorArt
}

// SymbolLoad is one 8x8 symbol set Init loads: block Block of File into set
// Set; an empty File is the content set's 8X8D<n>.
type SymbolLoad struct {
	Set   int
	File  string
	Block int
}

// PorSymbols are the sets POR, sotsb and cotab load (ovr35:003B).
var PorSymbols = []SymbolLoad{{Set: 4, Block: 202}, {Set: 0, Block: 203}}

// AreaMonsters are the monsters a fight adds by area screen (Behaviour.AreaMonstersJoin):
// Count of ByScreen[the area screen], standing aside with Status and Control.
type AreaMonsters struct {
	ByScreen        []AreaMonster
	Count           int
	Status, Control int
}

// AreaMonster is one area screen's monster and its combat icon.
type AreaMonster struct{ Monster, Icon int }

// scriptSound is what the game's number n plays: its ScriptSounds entry, else DefaultSound.
func (g GameTables) scriptSound(n byte) (ScriptSound, bool) {
	if s, ok := g.ScriptSounds[n]; ok {
		return s, true
	}
	if g.DefaultSound != nil {
		return *g.DefaultSound, true
	}
	return ScriptSound{}, false
}

// ScriptSound is what one Sound command number plays: a song or a cue.
type ScriptSound struct {
	Music bool          // Song, not Cue
	Song  int           // the soundtrack's song
	Cue   core.SoundCue // the sound pack's event
}

// bareRuleset and bareData are what an Engine made without New (a test's
// &Engine{}) builds from; the tests set them (r1e_test.go).
var (
	bareRuleset Ruleset
	bareData    GameTables
)

// ruleset is the Engine's ruleset, owner of the rules' state (fold step E-R6), made
// at first use by the game's Ruleset: the Engine reaches the rules through it.
func (e *Engine) ruleset() core.Rules {
	if e.rules == nil {
		build := e.newRules
		if build == nil {
			build = bareRuleset
		}
		monster := func(id core.MonsterID) (char.Sheet, []char.AffectRecord, []char.Item, bool) {
			m, ok := e.gdata().monster(e.monsterSet().File("MON"), int(id))
			return m.Record, m.Affects, m.Items, ok
		}
		treasure := func(n int) []char.Item {
			if n == vaultBlock {
				return e.vault.items
			}
			return e.gdata().itemBlock(e.set.File("ITEM"), n)
		}
		e.rules, e.create = build(e.rulesFile(), monster, e.readItems, treasure, creationHost{e})
	}
	return e.rules
}

// game is the game's GameTables (bareData for an Engine made without New).
func (e *Engine) game() GameTables {
	if e.given == nil {
		return bareData
	}
	return *e.given
}

// selFact is the selected member's Fact f as a number; a Fact the ruleset
// does not hold is a bug of the script table.
func (e *Engine) selFact(f core.Fact) int64 {
	return int64(e.selValue(f).(core.Number))
}

// setSelFact writes the selected member's Fact f.
func (e *Engine) setSelFact(f core.Fact, v int64) {
	e.setSelValue(f, core.Number(v))
}

// scriptSide is the Side word's number of each side (SpecialReadZ1 268);
// writing any other number changes nothing.
var scriptSide = map[core.Side]uint16{core.SideParty: 0, core.SideAuto: 0x80, core.SideMonsters: 0x81}

// selValue is the selected member's Fact f.
func (e *Engine) selValue(f core.Fact) core.Value {
	v, err := e.ruleset().Fact(e.selected(), f)
	if err != nil {
		panic(fmt.Sprintf("script: fact %v: %v", f, err))
	}
	return v
}

// setSelValue writes the selected member's Fact f.
func (e *Engine) setSelValue(f core.Fact, v core.Value) {
	if err := e.ruleset().SetFact(e.selected(), f, v); err != nil {
		panic(fmt.Sprintf("script: fact %v: %v", f, err))
	}
}

// selIndex is the selected member's typed Fact f as the script's number: its
// place in the record's table of f's values. A record byte outside the table
// comes back as a core.Number and is read as it stands (SpecialReadZ1).
func selIndex[V comparable](e *Engine, f core.Fact, table []V) int {
	switch v := e.selValue(f).(type) {
	case core.Number:
		return int(v)
	case V:
		return slices.Index(table, v)
	}
	panic(fmt.Sprintf("script: fact %v: unexpected value", f))
}

// worldWord is the world word w's value (core.WordIs); a word the rules do not
// keep is a bug of the Engine.
func (e *Engine) worldWord(w core.WorldWord) int {
	a := e.ruleset().Check(core.WordIs{Word: w}, nil)
	if !a.Yes {
		panic(fmt.Sprintf("rules keep no world word %d", w))
	}
	return int(a.Value)
}

// setWorldWord stores v in the world word w (core.SetWord).
func (e *Engine) setWorldWord(w core.WorldWord, v int) { e.do(core.SetWord{Word: w, Value: int32(v)}) }

// rollDice is n dice of sides from the rules' generator (core.RollDice).
func (e *Engine) rollDice(n, sides int) int {
	return int(e.ruleset().Check(core.RollDice{N: int32(n), Sides: int32(sides)}, nil).Value)
}

// rollRandom is the rules' Random(n) (core.RollRandom).
func (e *Engine) rollRandom(n int) int {
	return int(e.ruleset().Check(core.RollRandom{N: int32(n)}, nil).Value)
}

// rulesFile is the adventure's rules.json (tablespec.RulesFile in its data folder), "" when it has none.
func (e *Engine) rulesFile() string {
	p := filepath.Join(e.adv, DataDir, tablespec.RulesFile)
	if _, err := os.Stat(p); err != nil {
		return ""
	}
	return p
}
