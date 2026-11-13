package importer

import (
	"gobox/adventure"
	"gobox/rules"
	"gobox/rules1e"
)

// Layout1e is the shape of this game's tables (the imported data/rules.json's
// flat slot tables, the race rows), written as data/layout.json.
var SotsbLayout1e = rules1e.Layout{SlotStride: sotsbSpellLevels, Races: rules1e.RaceNumbers(rules1e.Layout439)}

// GameTables is the game's pool.GameTables, written as data/gametables.json:
// the second engine's dialect, an NPC loads its combat icon, the duel sounds
// (callSound); the portrait tables stay empty.
var GameTables = adventure.GameTables{
	Dialect:    rules.DialectCOTAB,
	AddNPCIcon: true,
	ScriptSounds: map[byte]adventure.ScriptSound{
		soundDuelA: {Cue: rules.CueStep},
		soundDuelB: {Cue: rules.CueFireball},
	},
	DefaultSound: &adventure.ScriptSound{Cue: rules.CueStep}, // callSound's default
	ScriptFixes:  journalFixes,
	OriginalSave: SotsbOriginalSave,
	MonsterSet:   1, // LoadCreatureData's MON1, LoadIconAnim's CPIC1
	Art: &adventure.ArtNames{ // LoadCombatTileSet's Load24x24Set (sotsb-re combatst.pas): outdoors it loads none
		DungeonTiles:  []adventure.TileLoad{{File: "DUNGCOM", Block: 1, Last: 31}},
		CombatSprites: "COMSPR",
		Title:         "TITLE",
	},
}

// OriginalSave is savgamX.dat (saveload.go, O0F8:2813/2FAB): the area byte, the area, game
// (area2) and aux records at script addresses 19200, 31744, 31232 (eclLayout), no code image,
// the position and states, the three wall slots.
var SotsbOriginalSave = rules1e.SaveLayout{
	Lead: rules1e.LeadAreaKind,
	Records: []rules1e.SaveRecord{
		{Base: int(sotsbEclLayout.Area.First), Cells: 1024, CellBytes: 2},
		{Base: int(sotsbEclLayout.Area2.First), Cells: 1024, CellBytes: 2},
		{Base: int(sotsbEclLayout.Aux.First), Cells: 512, CellBytes: 2},
	},
	CodeBase:  int(sotsbEclLayout.Code.First),
	CodeSize:  sotsbScriptCodeSize,
	WallBytes: saveWallSlots * 4,
}
