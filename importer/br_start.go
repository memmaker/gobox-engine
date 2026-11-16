package importer

import (
	"gobox/adventure"
	"gobox/game"
	"gobox/rules"
)

// brSeries is presentdata/data.go's Series (buckrogers-go), the same in buckrogers2-go rules/data.go:14.
const brSeries = "buck-rogers"

// brStart is what one game's start-up files hold: gametables.json's values, each from its evidence. Both games' sources are
// buckroger-re and buckrogers2-re (the Pascal), buckrogers-go and buckrogers2-go (the old engines), buckrogers-core (shared).
type brStart struct {
	// FirstBlock: eclvm.pas START_ECL_SCRIPT = 16 ("ECL1.dax's first record", both games); rules/adv.go begin() (m.Block = 16).
	FirstBlock int
	// GameSet: the disk a new game starts on (rules/adv.go begin(): e.disk = 1, both games).
	GameSet rules.Set
	// FirstBigPicture: Picture numbers from here show BIGPIC (eclvm.pas FIRST_BIG_PICTURE = 112; rules/eclops.go picture).
	FirstBigPicture int
	// CodeBase and CodeSize: the loaded block's script addresses (buckrogers-core ecl.go AddrCode = 32768, CodeSize = 9217).
	CodeBase, CodeSize int
	// Symbols: Load8x8Set(0, 6) is BORDERS block 0 (BR1 sym8x8.pas BORDER_SLOT = 6), Load8x8Set(202, 4) the data set's 8X8D<set>
	// block 202 (gameinit.pas:243-244 in BR1, :248-249 in BR2).
	Symbols []adventure.SymbolLoad
	// Art: gameinit.pas LoadIconAnim('COMSPR') (BR1 :253, BR2 :253); the title file the game's start sequence shows.
	Art adventure.ArtNames
	// PictureSets: pO124_02D0 reads these PIC blocks from PIC7..9 whatever the disk (rules/eclops.go picOf in BR1, smallPic in BR2).
	PictureSets map[int]rules.Set
	// Walls: the game has a wall texture table (importer/walls/<id>.txt); false: none yet (BR2).
	Walls bool
	// Data: the front's data (gamedata.json) beyond the series and id; nil: not moved over from the old front yet (BR2).
	Data func(id string) game.Data
	// Logbook: the journal's titles (importer/br1/journal, from the printed Log Book); nil: none yet.
	Logbook *adventure.Logbook
	// Overland: the squares MapX, MapY range over on the overland (rules.Size{}: none yet).
	Overland rules.Size
	// Places: the blocks' music triggers (br_places.go); nil: none yet.
	Places []adventure.Place
}

const (
	brBorderSlot, brGameSlot   = 6, 4
	brBorderBlock, brGameBlock = 0, 202
)

var (
	brSymbols = []adventure.SymbolLoad{
		{Set: brBorderSlot, File: "BORDERS", Block: brBorderBlock},
		{Set: brGameSlot, Block: brGameBlock},
	}
	// brPictureSets is BR1's and BR2's (buckrogers-go eclops.go picOf; buckrogers2-go eclops.go:254-260 smallPic: the same six).
	brPictureSets = map[int]rules.Set{105: 7, 64: 7, 72: 8, 92: 8, 74: 8, 62: 9}
)

// brArt is the art BR1 and BR2 name: the combat sprites COMSPR blocks 1..25 in icon rows 13..37 (gameinit.pas
// LoadIconAnim('COMSPR', index, index + 13 - 1)) and the shots' icons of cbtacts.pas AnimateMissile (the same in both games:
// the pace 10; the directed ones are Animate(firstKind), the rest the two-frame icons 30 and 37, which also fly for a grenade).
var brArt = adventure.ArtNames{CombatSprites: "COMSPR", Title: "TITLE", IconMask: brIconMask, PlainIcons: true, MemberIcons: "CHARS", ActionIcons: true, Marks: brMarks, OverlandCursor: adventure.MarkPicture{File: "CURSOR", Block: brCursorBlock, Mask: brCursorMask}, ShipPictures: "SHIPS",
	PlaceTiles: brCombatTiles, TerrainTiles: brTerrainTiles,
	SpriteSlots: []adventure.SpriteLoad{{Block: 1, Slot: 13, Count: 25}},
	Missiles: []adventure.MissileArt{
		{Kind: rules.FlightDefault, Slot: 37, Pace: brMissilePace},
		{Kind: rules.FlightWave, Slot: 30, Pace: brMissilePace},
		{Kind: rules.FlightHeatBolt, Slot: 13, Directed: true, Pace: brMissilePace},
		{Kind: rules.FlightLaser, Slot: 16, Directed: true, Pace: brMissilePace},
		{Kind: rules.FlightCrossbowBolt, Slot: 19, Directed: true, Pace: brMissilePace},
		{Kind: rules.FlightGunBolt, Slot: 22, Directed: true, Pace: brMissilePace},
		{Kind: rules.FlightRocket, Slot: 25, Directed: true, Pace: brMissilePace},
	}}

// brMarks are the markers over the ground (cbtmap.pas DrawArenaCell: iconFrames[(tile shr 1) + 33, tile and 1] over the cell, where
// tile is gdconsts.pas combatTerrain[t].tile = t - 1 for the arena values 1..5 (chaff, gas, 3, aerosol mist, corpse)); rows 33..35
// are COMSPR blocks 21..23 (LoadIconAnim loads block n into row n + 12).
var brMarks = []adventure.MarkArt{
	{Mark: 1, Slot: 33, Frame: 0}, {Mark: 2, Slot: 33, Frame: 1}, {Mark: 3, Slot: 34, Frame: 0},
	{Mark: 4, Slot: 34, Frame: 1}, {Mark: 5, Slot: 35, Frame: 0},
}

// The wilderness cursor (gameinit.pas: LoadPic('cursor', 1, True, C_Pic.KEY_COLOR, ..); mapcurs.pas DrawCursor draws it over the party's 8x8 cell).
const (
	brCursorBlock = 1
	brCursorMask  = 13
)

// brIconMask is the colour COMSPR, CHARS and CPIC icons are masked on (maptiles.pas LoadIconAnim: LoadPic(.., True, 8, ..)).
const brIconMask = 8

// brCombatTiles is combatst.pas LoadCombatTileSet: the arena's tile sheet by the area kind, Load24x24Set(name, 1, slot, n) laying
// the sheet's block 1 from tile slot: marscom (Mars wilderness $A8: a wilderness with the terrain shift) and venuscom (the
// other wilderness) 52 tiles from 1, lotek ($88, $82) and hitek (all else) 53 from 0.
var brCombatTiles = []adventure.TileLoad{
	{File: "MARSCOM", Block: 1, First: 1, Last: 51, When: rules.TraitWilderness | rules.TraitTerrainShift},
	{File: "VENUSCOM", Block: 1, First: 1, Last: 51, When: rules.TraitWilderness, Unless: rules.TraitTerrainShift},
	{File: "LOTEK", Block: 1, First: 0, Last: 52, When: rules.TraitLowTech, Unless: rules.TraitWilderness},
	{File: "HITEK", Block: 1, First: 0, Last: 52, Unless: rules.TraitWilderness | rules.TraitLowTech},
}

// brMissilePace is AnimateMissile's local.pace := 10.
const brMissilePace = 10

// brStarts by game id.
var brStarts = map[string]brStart{
	"br1": {FirstBlock: 16, GameSet: 1, FirstBigPicture: 112, CodeBase: 32768, CodeSize: 9217, Symbols: brSymbols, PictureSets: brPictureSets,
		Art:     brArt,                                                                                                                                           // br1Start's TITLE
		Logbook: &adventure.Logbook{Box: "Log Book", Title: "Log Entry "}, Data: br1GameData, Walls: true, Places: brPlaces, Overland: rules.Size{W: 38, H: 15}}, // buckrogers-go adv.go overland(): Area[195] 0..37, [196] 0..14
	// BR2: buckrogers2-go rules/adv.go:147,161 (block 16), :158 (disk 1), eclops.go:239 (112), buckrogers2-re eclvm.pas:42,48. The
	// Overland: buckrogers2-go rules/adv.go overland(): area 195 0..37, 196 0..14. Logbook: buckrogers2-go rules/journal.go logBox "Log Book", logTitle "Log Entry N" over data/journal (importer/br2/journal).
	"br2": {FirstBlock: 16, GameSet: 1, FirstBigPicture: 112, CodeBase: 32768, CodeSize: 9217, Symbols: brSymbols, PictureSets: brPictureSets,
		Art: brArt, Walls: true, Data: br2GameData, Places: br2Places, Overland: rules.Size{W: 38, H: 15}, Logbook: &adventure.Logbook{Box: "Log Book", Title: "Log Entry "}}, // br2_importbr2.go br2Start's TITLE
}
