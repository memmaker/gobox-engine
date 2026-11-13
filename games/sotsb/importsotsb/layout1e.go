package importsotsb

import (
	"gobox/char"
	"gobox/core"
	"gobox/core/dnd"
	"gobox/pool"
	"rules1e"
)

// races is the race of each race row of the game's tables ("": the
// placeholder row 0).
var races = [...]core.Race{"", dnd.Elf, dnd.HalfElf, dnd.Dwarf, dnd.Gnome, dnd.Halfling, dnd.Human, dnd.Monster}

// Layout1e is the shape of this game's tables (the imported data/rules.json's
// flat slot tables, the race rows), written as data/layout.json.
var Layout1e = rules1e.Layout{SlotStride: spellLevels, Races: races[:]}

// GameTables is the game's pool.GameTables, written as data/gametables.json:
// the second engine's dialect, an NPC loads its combat icon, the duel sounds
// (callSound); the portrait tables stay empty.
var GameTables = pool.GameTables{
	Dialect:    core.DialectCOTAB,
	AddNPCIcon: true,
	ScriptSounds: map[byte]pool.ScriptSound{
		soundDuelA: {Cue: core.CueStep},
		soundDuelB: {Cue: core.CueFireball},
	},
	DefaultSound: &pool.ScriptSound{Cue: core.CueStep}, // callSound's default
	ScriptFixes:  journalFixes,
	OriginalSave: OriginalSave,
	MonsterSet:   1, // LoadCreatureData's MON1, LoadIconAnim's CPIC1
	Art: &pool.ArtNames{ // LoadCombatTileSet's Load24x24Set (sotsb-re combatst.pas): outdoors it loads none
		DungeonTiles:  []pool.TileLoad{{File: "DUNGCOM", Block: 1, Last: 31}},
		CombatSprites: "COMSPR",
		Title:         "TITLE",
	},
}

// OriginalSave is savgamX.dat (saveload.go, O0F8:2813/2FAB): the area byte, the area, game
// (area2) and aux records at script addresses 19200, 31744, 31232 (eclLayout), no code image,
// the position and states, the three wall slots.
var OriginalSave = char.SaveLayout{
	Lead: char.LeadAreaKind,
	Records: []char.SaveRecord{
		{Base: int(eclLayout.Area.First), Cells: 1024, CellBytes: 2},
		{Base: int(eclLayout.Area2.First), Cells: 1024, CellBytes: 2},
		{Base: int(eclLayout.Aux.First), Cells: 512, CellBytes: 2},
	},
	CodeBase:  int(eclLayout.Code.First),
	CodeSize:  scriptCodeSize,
	WallBytes: saveWallSlots * 4,
}
