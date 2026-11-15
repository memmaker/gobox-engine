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
	BigPictures                BigPictures          // the Picture command's whole-screen BIGPIC pictures; zero: none (POR)
}

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
