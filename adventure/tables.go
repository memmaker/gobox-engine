package adventure

import (
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
	"slices"
)

// GameTables is the game's data the Engine shows, handed in by the game at New.
type GameTables struct {
	PortraitHead, PortraitBody []rules.PortraitPic  // portrait number -> its picture, 1-based ([0] unused)
	Script                     *ecl.Version         `json:"-"` // the game's ECL dialect; nil: the first engine's (V1)
	ScriptSounds               map[byte]ScriptSound // the Sound command's numbers (pod: 2..14 digitized, 128+n song n); one not listed plays DefaultSound
	DefaultSound               *ScriptSound         // what a number ScriptSounds does not list plays; nil: nothing
	AddNPCIcon                 bool                 // AddNPC loads the NPC's combat icon (sotsb, pod; not cotab)
	Reserve                    Reserve              // the reserve every scripted fight gets (pod); zero: none
	Dialect                    rules.Dialect        // the game's script byte-code dialect (ReadGameTables sets Script from it)
	TreasureSets               [][]int32            // ItemSet's sets: set number -> its library item ids (pod: 1..11; [0] unused)
	ScriptFixes                []ecl.Fix            // the game's decompiler special cases (ecl.Library.Fixes): the import's, for the texts' key
	Quests                     []QuestDef           // the quest journal's quests (POR: the commissions); none: an empty journal
	OriginalSave               rules1e.SaveLayout   // the game's SAVGAMx.DAT format (the original-save import)
	PurseWhos                  []PurseWho           // the Who commands that spend the party's one purse (POR: the inn, the tavern)
	Symbols                    []SymbolLoad         // the 8x8 symbol sets Init loads; nil: the first engine's (DefaultSymbols: POR, sotsb, cotab)
	Art                        *ArtNames            // the pictures the Engine loads by role; nil: none (every import writes its own)
	MapLabels                  []MapLabel           // the maps' fixed labels (POR: New Phlan's buildings); none: none
	Music                      *AreaMusic           // the soundtrack by area (POR's); nil: none
	MonsterSet                 rules.Set            // the set whose MON and CPIC archives hold the monsters (sotsb, pod: 1); 0: the current set (POR, cotab: the disk)
	Walls                      *WallTextures        // the 3D view's remastered wall textures; nil: none
	GameSet                    rules.Set            // the set the party menu's game starts in (DS:5376: POR 3; cotab, sotsb, pod 2)
	FirstBlock                 int                  // the ECL block a new game starts on, with no last block saved (POR 0, cotab 1, sotsb 17, pod 16)
	Logbook                    *Logbook             // the Logbook command's book (BR); nil: the game has none
	PictureSets                map[int]rules.Set    // a PIC block kept in another set's file than the one in play (BR1 pO124_02D0: 62 in PIC9); nil: none
	BigPictures                BigPictures          // the Picture command's whole-screen BIGPIC pictures; zero: none (POR)
	OverlandSteps              rules.Size           // MapX and MapY range over 0..W-1 and 0..H-1 on the overland, which is the whole map (BR1: 38 by 15, Area[195], [196]); zero: the Pool's window of the map (16 by 36, shifted by the view)
	Places                     []Place              // the area blocks' own places (the front's music trigger); none: none
	Space                      *SpaceMap            // the rocket's system map (BR1); nil: the game has none (the Pool games, BR2 yet)
	Overland                   []OverlandBlock      // the script blocks that are the overland's maps (POR's three wildernesses); none: none (the others' scripts say AreaKind)
	NoRemainsBlocks            []int                // the script blocks whose arena clutter has no remains (POR combatst.pas SprinkleRareFeatures: eclScriptId <> 10); none: none
}

// OverlandBlock is a script block that is an overland map: outdoors (not 3D) its start leaves the view
// at ViewMode (2..4, the wilderness's place). Finale: with the boss gone (Boss 0xFE) and the river shut
// the area's start runs the endgame hook (POR's central wilderness).
type OverlandBlock struct {
	Block    int
	ViewMode int
	Finale   bool
}

// OverlandBlock is the overland entry of script block block; nil: it is not an overland map.
func (t GameTables) OverlandBlock(block int) *OverlandBlock {
	for i := range t.Overland {
		if t.Overland[i].Block == block {
			return &t.Overland[i]
		}
	}
	return nil
}

// NoRemains reports whether script block block's arena clutter has no remains.
func (t GameTables) NoRemains(block int) bool { return slices.Contains(t.NoRemainsBlocks, block) }

// Place is area block Block's own place: the music trigger the front plays its own music for
// (game.Where.Music, a gobox data/music folder) and whether it is a town's peaceful streets
// (Where.Mode "town"). A block's number is the game's AreaKind, one place a block (BR1's
// blocks are unique over its disks).
type Place struct {
	Block int
	Music string
	Town  bool
}

// Place is the place of area block block; nil: none.
func (t GameTables) Place(block int) *Place {
	for i := range t.Places {
		if t.Places[i].Block == block {
			return &t.Places[i]
		}
	}
	return nil
}

// Logbook is the printed book the Logbook command names entries of by number: entry n is the journal
// text titled Title followed by n, read under the box title Box.
type Logbook struct{ Box, Title string }

// BigPictures are the Picture command's numbers that show BIGPIC<set> block n
// instead of PIC<set>'s view-window picture.
type BigPictures struct {
	First       int  // the first big picture number (sotsb 112, cotab 120, pod 240); 0: none
	Framed      int  // a big picture that keeps the screen's frame, not whole-screen (sotsb 121, the overland map); 0: none
	NotOverland bool // on the overland a big picture keeps the frame (pod O040:07DB)
}

// WallTextures are the remastered textures of the WALLDEF walls the 3D view
// draws, and its flats (floor_<Floor>, ceiling_<Sky> in the pool series).
type WallTextures struct {
	Floor, Sky string
	Walls      []WallTexture
	Areas      []AreaLook // the areas' own flats and scenery
	Backs      []BackLook // the flats by BACK picture (a game whose view paints a picture over the flats)
}

// BackNone is the BackLook block of a view with no BACK picture (the sky fill over the ground).
const BackNone = -1

// BackLook is the flats (pool floor_<Floor>, ceiling_<Sky>) of BACK<Set>.DAX
// picture Block, the one the view shows (view3d.pas LoadBackdrop: the last
// loaded wall set's block, + 128 where the party's map cell has no bit 7);
// BackNone: the set's look where no picture exists. A sky_* ceiling is an
// outdoor sky over the horizon; any other is tiled overhead.
type BackLook struct {
	Set        rules.Set
	Block      int
	Floor, Sky string
}

// Back is the look of BACK picture block of set set; nil: none.
func (w *WallTextures) Back(set rules.Set, block int) *BackLook {
	for i := range w.Backs {
		if w.Backs[i].Set == set && w.Backs[i].Block == block {
			return &w.Backs[i]
		}
	}
	return nil
}

// AreaLook is GEO block Geo of set Set's own 3D view look: its flats
// (pool floor_<Floor>, ceiling_<Sky>; "": the game's) and scenery.
type AreaLook struct {
	Set        rules.Set
	Geo        int
	Floor, Sky string
	Scenery    game.Scenery
}

// Area is the look of GEO block geo of set set; nil: the game's.
func (w *WallTextures) Area(set rules.Set, geo int) *AreaLook {
	for i := range w.Areas {
		if w.Areas[i].Set == set && w.Areas[i].Geo == geo {
			return &w.Areas[i]
		}
	}
	return nil
}

// WallTexture is wall Wall (1..rules.SetWalls) of chunk Chunk (1-based) of
// block Block of set Set's WALLDEF archive: the pool series texture Texture.
// Solid: the texture looks like a solid wall, so a side with it the map
// lets the party through is a secret door (an illusionary wall).
type WallTexture struct {
	Set                rules.Set
	Block, Chunk, Wall int
	Texture            string
	Solid              bool
}

// ArtNames are the picture files the Engine loads by role, each game's own
// (POR's SQRPACI tile set is cotab's TILES; sotsb and pod have none). An
// empty name or list: the game has no such picture.
type ArtNames struct {
	TileSheets     []TileLoad   // whole sheets Init takes as the tile set, tile First/0x80's half (POR: SQRPACI 1 and 2); none: a blank set
	Tiles          []TileLoad   // laid over the tile set at Init (cotab: TILES 1 and 2, copyprot's Load24x24Set)
	DungeonTiles   []TileLoad   // what combat setup lays over the tiles in a dungeon (ovr11:1244, LoadCombatTileSet)
	WildTiles      []TileLoad   // and outdoors (sotsb: none, the set stays as it is)
	AreaTiles      []TileLoad   // the area map's (BACPAC), laid over the tiles off the 3D view
	AreaIcon       string       // the area map's party icon sheet (ICON)
	CombatSprites  string       // the combat sprites (COMSPR)
	Title          string       // the title pictures (TITLE)
	BorderCells    []int        // DrawBorder style 0's corner, side and top symbols (POR: 0x114..0x116); none: no border pictures
	SpriteSlots    []SpriteLoad // the combat sprites' blocks and the icon slots they fill at Init; none: PoolSpriteSlots
	Missiles       []MissileArt // what flies for each FlightKind; none: PoolMissiles
	PlaceTiles     []TileLoad   // laid over a fresh tile set at a fight's setup, the sheets whose When and Unless fit the place (BR1: marscom, venuscom, lotek, hitek)
	TerrainTiles   []int        // the tile each ground terrain value draws, by value (BR1's combatTerrain[t].tile; index 0 for any value outside the table); none: the value is the tile
	IconMask       byte         // the colour the combat icons (sprites, monsters, members) are masked on (the Pool's 0; BR1's 8)
	PlainIcons     bool         // the monsters' icons are loaded as drawn (VGA art): not recoloured with the Pool's EGA palette
	OverlandCursor MarkPicture  // the picture that marks the party's square on an overland drawn as one whole picture (BR1: CURSOR block 1); File "": none
	ShipPictures   string       // the file of the foe ships' pictures, block by the ruleset's Ship.Picture (BR: SHIPS); none: no picture
	Marks          []MarkArt    // the picture over a square for each of the ruleset's SquareMarks; none: no marker is drawn
	ActionIcons    bool         // a unit shows its icon's second frame (block+0x80) while it attacks (BR1: cbtmap.pas SetCreatureFacing frame 1 in cbtacts.pas MakeAttack)
	MemberIcons    string       // the sheet of a party member's whole-picture icon (BR1: CHARS, block = the record's icon, +0x80 the action pose); none: composed from CHEAD and CBODY
}

// MarkPicture is a picture of a file's block, masked on one colour.
type MarkPicture struct {
	File        string
	Block, Mask int
}

// MarkArt is the picture of one SquareMark: the combat sprite slot Slot's frame Frame (0 ready, 1 action).
type MarkArt struct {
	Mark        rules.SquareMark
	Slot, Frame int
}

// SpriteLoad is a run of Count blocks of the combat sprites from block Block into the icon slots from Slot (Init's
// LoadIconPic of COMSPR: POR's blocks 0..0xB fill slots 0xD.., BR's gameinit.pas LoadIconAnim loop blocks 1..25 fill 13..37).
type SpriteLoad struct{ Block, Slot, Count int }

// PoolSpriteSlots are the Pool games' (ovr33 Init: blocks 0..0xB into slots 0xD.., block 0x19 into slot 0x19).
var PoolSpriteSlots = []SpriteLoad{{Block: 0, Slot: 0xD, Count: 0xC}, {Block: 0x19, Slot: 0x19, Count: 1}}

// MissileArt is the picture of what flies for one FlightKind (a Spell's when Spell): the combat sprite slot Slot
// (the slot's first frame; Directed: the slot's eight directions as AnimateMissile chooses them, from Slot on) and the
// flight's pace in ms per third of a square, the original's.
type MissileArt struct {
	Spell    bool
	Kind     rules.FlightKind
	Slot     int
	Directed bool
	Pace     int
}

// PoolMissiles are the Pool games' (ovr14:26B0 attackAnimation and its spells' callers; slot 0xD is the first sprite).
var PoolMissiles = []MissileArt{
	{Spell: true, Kind: rules.FlightDefault, Slot: 0x12, Pace: 0x1E},
	{Spell: true, Kind: rules.FlightGaze, Slot: 0x12, Pace: 0x2D},
	{Spell: true, Kind: rules.FlightSpit, Slot: 0x17, Pace: 0x1E},
	{Spell: true, Kind: rules.FlightBolt, Slot: 0x13, Pace: 0x32},
	{Kind: rules.FlightDefault, Slot: 0x14, Pace: 0x14},
	{Kind: rules.FlightSlingStone, Slot: 0x15, Pace: 0x14},
	{Kind: rules.FlightPointed, Slot: 0xD, Directed: true, Pace: 0xA},
	{Kind: rules.FlightTumbling, Slot: 0x10, Pace: 0x32},
	{Kind: rules.FlightFlask, Slot: 0x11, Pace: 0x32},
}

// TileLoad is a block of File laid into the tile set from tile First: its
// frames 0..Last (LoadBackpacIcons' n, Load24x24Set's count-1). The block is
// Block, or for From a value of the game's state.
type TileLoad struct {
	File               string
	Block, First, Last int
	From               TileFrom          `json:",omitempty"`
	When, Unless       rules.PlaceTraits `json:",omitempty"` // PlaceTiles: laid when the place has all of When and none of Unless
}

// TileOf is the tile terrain value t draws.
func (a ArtNames) TileOf(t int) int {
	if len(a.TerrainTiles) == 0 {
		return t
	}
	if t < 1 || t >= len(a.TerrainTiles) {
		t = 0
	}
	return a.TerrainTiles[t]
}

// Fits is whether a place with traits t takes the load (PlaceTiles).
func (l TileLoad) Fits(t rules.PlaceTraits) bool { return t.Has(l.When) && t&l.Unless == 0 }

// TileFrom is where a TileLoad's block number comes from.
type TileFrom int

const (
	TileBlockFixed   TileFrom = iota // TileLoad.Block
	TileBlockWalls                   // the low byte of the loaded wall set's id (pod's DungCom)
	TileBlockAreaSet                 // the area's symbol set when 1 or 2, else 0 (pod's WildCom, by its tileSet)
)

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

// SymbolLoad is one 8x8 symbol set Init loads: block Block of File into set
// Set; an empty File is the content set's 8X8D<n>.
type SymbolLoad struct {
	Set   int
	File  string
	Block int
}

// DefaultSymbols are the sets the first engine's Init loads (ovr35:003B; POR, sotsb, cotab).
var DefaultSymbols = []SymbolLoad{{Set: 4, Block: 202}, {Set: 0, Block: 203}}

// Reserve is a fight's reserve, any adventure's to fill: Count of the
// monster of the area's symbol set join every scripted fight off the map
// (rules.HoldInReserve), with Control, until an effect reinforces a side
// with them. Empty: the adventure has none. (pod-re CharFightsAlone.)
type Reserve struct {
	BySet   []ReserveUnit
	Count   int
	Control int
}

// ReserveUnit is one symbol set's monster and its combat icon.
type ReserveUnit struct{ Monster, Icon int }

// ScriptSound is what one Sound command number plays: a song or a cue.
type ScriptSound struct {
	Music bool           // Song, not Cue
	Song  int            // the soundtrack's song
	Cue   rules.SoundCue // the sound pack's event
}

// ScriptAt is a label of an area script.
type ScriptAt struct {
	Area  int    // the area script
	Label string // its label
}

// PurseWho is a Who command that spends the party's one purse: the label's
// instruction plus By (GameTables.PurseWhos).
type PurseWho struct {
	ScriptAt
	By int
}

// QuestDef is a quest (GameTables.Quests, the import's): its id (Quest's value), journal
// name (at most 19 letters: the list's width), giver and place, the giver's own words (the
// scripts' texts that hand it over), its done and its offered tests.
type QuestDef struct {
	ID                 int
	Name, Giver, Place string
	Brief              []string
	Done, Given        QuestTest
}

// QuestTest holds when any of its fields is at least its Min; empty never holds.
type QuestTest []FieldAtLeast

// FieldAtLeast is the game's saved field (layout/) at least Min.
type FieldAtLeast struct {
	Field string
	Min   int64
}

// Sound is what the game's number n plays: its ScriptSounds entry, else DefaultSound.
func (g GameTables) Sound(n byte) (ScriptSound, bool) {
	if s, ok := g.ScriptSounds[n]; ok {
		return s, true
	}
	if g.DefaultSound != nil {
		return *g.DefaultSound, true
	}
	return ScriptSound{}, false
}

// For is the reserve's monster in an area of symbol set set; none past the table.
func (t Reserve) For(set int) (ReserveUnit, bool) {
	if set < 0 || set >= len(t.BySet) {
		return ReserveUnit{}, false
	}
	return t.BySet[set], true
}
