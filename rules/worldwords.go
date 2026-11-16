package rules

// World is the engine's own world words: state the engine and the scripts
// read and write, which no ruleset reads. The engine keeps one (it was the 1e
// records' cells) and the canonical save holds it; it lives here, not in
// adventure or engine, because the original-save reader of a ruleset fills it.
// Each cell is 16 bits wide, as the records' words were.
type World struct {
	AreaKind   uint16 // the area's ECL block number; also the block NewECL left (POR keeps it here, as the original's DS:5D7D copy)
	PrevX      uint16 // the square before the step (the step sound)
	PrevY      uint16
	RiverOpen  uint16 // nonzero: the river can be crossed
	Search     uint16 // bit 0 sweeping, bit 1 search this square
	HoldMove   uint16 // >= 0xFF: the next move is skipped
	LeftMap    uint16 // the step left the map
	Geo        uint16 // the GEO block of the 3D map
	WallSplitA uint16 // with WallSplitB set, opLoadWalls loads into slots 1 and 3
	WallSplitB uint16
	NoCoords   uint16 // nonzero: no x,y on the status line, no area map
	MsgDelay   uint16 // text delay, the save file's copy
	ViewFlags  uint16 // the save file's copy of the view flags (bit 1 and bit 0)
	MaxDist    uint16 // opSetupMonster's distance
	Portrait   uint16 // HEAD/BODY pair shown instead of a picture, 0xFF none
	Temple     uint16 // 1: opCombat opens the temple
	Shop       uint16 // 1: opCombat opens the shop
	SymbolSet  uint16 // the 8X8D<n> symbol set

	// The fight's words: the engine hands the inputs to the rules in FightTerms
	// and sets the outputs from FightResult.
	Morale       uint16 // 0..100 (the engine caps it when a fight begins), the monsters' flee threshold
	Surprise     uint16 // bit side+1: that side is surprised, the first round only
	MonsterHit   uint16 // low byte: signed to-hit bonus of the monsters
	PartyHit     uint16 // low byte: signed to-hit bonus of the party
	PartyMove    uint16 // movement bonus of the party
	NoLoot       uint16 // 1: the monsters drop no items
	Duel         uint16 // DuelSetup's n
	MonsterDist  uint16 // the encounter's distance, counted down by opApproach
	CombatEnd    uint16 // 0x80 the party died, 0x81 it fled
	Beaten       uint16 // monsters beaten
	AttackedAlly uint16 // 1: the party attacked an ally this fight
	// The place, a rest and a hall: the 1e rules read them, the engine hands them in with the
	// question that needs them (AreaMarks, BuildGround, SkyDay, HallTerms, CampOpen ...).
	NoMagic    uint16 // nonzero: no spellcasting here
	In3D       uint16 // nonzero: a 3D area; zero: the overland map
	Scripted   uint16 // nonzero: the place is run by its script: no walking, the move entry runs again until it asks (BR view kind 3)
	MapX, MapY uint16 // the party's square on the overland map

	// The rocket's flight (BR view kind 2, the system map): the engine owns where the party is. The bodies' squares are derived from
	// OrbitStep (adventure.SpaceMap.Body), never kept.
	Flying        uint16 // nonzero: the party flies the rocket on the system map (the overland's map data is the system map's)
	RocketX       uint16 // the ship's square on the system map
	RocketY       uint16
	RocketHeading uint16 // the way the ship points: an index of SpaceMap.Steps (counterclockwise from north)
	OrbitTick     uint16 // time on the map: a move adds SpaceMove.Ticks, a wrap of SpaceMove.TickCycle moves the bodies on
	OrbitStep     uint16 // where the bodies stand on their orbits, counted down through 0..SpaceMap.Cycle-1

	Boss        uint16 // > 0xFD: map marks 8 and 9; 0xFE: the river ending; 0xFF: the combat terrain kind
	MarkFlags4  uint16 // bit 3: area map mark 4 shown
	Mark3       uint16 // 0xFF: area map mark 3 hidden
	Mark4Off    uint16 // zero: area map mark 4 shown
	MarkFlags57 uint16 // bit 1 clear: mark 5, bit 0 clear: mark 7
	Mark4       uint16 // nonzero: area map mark 4 shown
	Sky         uint16 // the 3D view's sky colour; a write redraws the portrait
	SkyInside   uint16 // the sky colour on map cells > 0x7F
	RestEvery   uint16 // a rest's encounter check every n steps
	RestChance  uint16 // its percent chance
	PartyExp    uint16 // the classes a training hall teaches (the ruleset's bits: ClassBits)
}

// Cells are the world's cells by their script field names (a layout names a
// field so when it has it).
func (w *World) Cells() map[string]*uint16 {
	return map[string]*uint16{
		"AreaKind": &w.AreaKind, "PrevX": &w.PrevX, "PrevY": &w.PrevY, "RiverOpen": &w.RiverOpen,
		"Search": &w.Search, "HoldMove": &w.HoldMove, "LeftMap": &w.LeftMap, "Geo": &w.Geo,
		"WallSplitA": &w.WallSplitA, "WallSplitB": &w.WallSplitB, "NoCoords": &w.NoCoords,
		"MsgDelay": &w.MsgDelay, "ViewFlags": &w.ViewFlags, "MaxDist": &w.MaxDist,
		"Portrait": &w.Portrait, "Temple": &w.Temple, "Shop": &w.Shop, "SymbolSet": &w.SymbolSet,
		"Morale": &w.Morale, "Surprise": &w.Surprise, "MonsterHit": &w.MonsterHit, "PartyHit": &w.PartyHit,
		"PartyMove": &w.PartyMove, "NoLoot": &w.NoLoot, "Duel": &w.Duel, "MonsterDist": &w.MonsterDist,
		"CombatEnd": &w.CombatEnd, "Beaten": &w.Beaten, "AttackedAlly": &w.AttackedAlly,
		"NoMagic": &w.NoMagic, "In3D": &w.In3D, "MapX": &w.MapX, "MapY": &w.MapY, "RocketX": &w.RocketX, "RocketY": &w.RocketY, "Boss": &w.Boss,
		"MarkFlags4": &w.MarkFlags4, "Mark3": &w.Mark3, "Mark4Off": &w.Mark4Off, "MarkFlags57": &w.MarkFlags57,
		"Mark4": &w.Mark4, "Sky": &w.Sky, "SkyInside": &w.SkyInside, "RestEvery": &w.RestEvery,
		"RestChance": &w.RestChance, "PartyExp": &w.PartyExp,
	}
}
