package rules1e

// GroundGen is how a game lays its combat ground (rules.json combatGround):
// the tables, tiles and terrain sets where the games differ; the passes
// themselves are shared. POR's is PorGroundGen, the zero value of a game
// that carries none.

// FlagValue is Value where the terrain's flags hold Flag; in a list the
// later entries override the earlier ones.
type FlagValue struct {
	Flag  TerrainFlag `json:"flag"`
	Value int         `json:"value"`
}

// TerrainFlag is what a wilderness square holds, the bits below.
type TerrainFlag byte

// Terrain flags (the bits of Tables.Terrain's entries).
const (
	TerrainSparse   TerrainFlag = 2   // fewer trees and rocks
	TerrainThin     TerrainFlag = 4   // a few fewer
	TerrainForest   TerrainFlag = 8   // more trees
	TerrainHills    TerrainFlag = 16  // a cliff or road now and then
	TerrainMountain TerrainFlag = 32  // more often
	TerrainRough    TerrainFlag = 64  // more of everything
	TerrainBare     TerrainFlag = 128 // no trees, much less of anything
)

type GroundGen struct {
	Fill    Ground      `json:"fill"`    // dungeon: the ground the map starts as, 0: none (PoD: open floor)
	Corners CornerTiles `json:"corners"` // dungeon: the corner tiles where the games differ
	Clutter *ClutterGen `json:"clutter"` // what lies on open floor now and then; nil: nothing
	Rubble  *RubbleGen  `json:"rubble"`  // dungeon: heaps in flagged rooms; nil: none
	Wild    *WildGen    `json:"wild"`    // outdoors; nil: the map stays bare (SotSB fights only in dungeons)
}

// CornerTiles are the tiles of the cases a game's corner code decides
// differently (the numbers are floor-tile numbers, the ground is that + 1).
type CornerTiles struct {
	NWDoorOpen byte `json:"nwDoorOpen"` // north-west a: a door to the west, open to the north, the corner open (POR 0xD, PoD 0)
	NEWallOpen byte `json:"neWallOpen"` // north-east b: a wall to the north-east, nothing east or beyond (POR 3, the others 7)
	NEDoorOpen byte `json:"neDoorOpen"` // north-east b: a door there, the same (POR 1, PoD 7, the others 1)
}

// ClutterGen is POR's floor clutter (ovr11:0000), on open floor Tile.
type ClutterGen struct {
	Tile    byte      `json:"tile"`
	Rare    [2]Ground `json:"rare"`    // 99 and 100 on 1d100 (the second: remains first)
	Remains [2]Ground `json:"remains"` // the remains, 6 in 10 the first
}

// RubbleGen is the dungeon rubble of CotAB, SotSB and PoD: in a room whose
// map cell has CellFlag, on open floor Tile, a heap with Edge round it.
type RubbleGen struct {
	CellFlag byte   `json:"cellFlag"`
	Tile     byte   `json:"tile"`
	Heap     Ground `json:"heap"`
	Edge     Ground `json:"edge"`
	HeapOdds int    `json:"heapOdds"` // in 10
	EdgeOdds int    `json:"edgeOdds"`
}

// WildGen is the wilderness: the ground it starts as, where the terrain
// comes from and the passes, run in this order; a nil pass is skipped.
type WildGen struct {
	Grass     Ground        `json:"grass"`
	Open      byte          `json:"open"` // the tile things may be put on
	Terrain   TerrainSource `json:"terrain"`
	BossHides bool          `json:"bossHides"` // POR: where the area's boss byte is 0xFF the buildings kind (0x80) is plain (1)
	Strip     *StripGen     `json:"strip"`
	Groves    *GroveGen     `json:"groves"`
	Trees     *TreeGen      `json:"trees"`
	Plants    *PlantGen     `json:"plants"`
	Scatter   *ScatterGen   `json:"scatter"`
}

// TerrainSource says where a fight's terrain number comes from.
type TerrainSource string

const (
	TerrainFromMap   TerrainSource = "map"   // the area's map square (POR: TerrainMap)
	TerrainFromGiven TerrainSource = "given" // the engine's (BuildGround.Terrain)
	TerrainNone      TerrainSource = "none"  // none: the terrain flags are 0
)

// StripStyle is the shape of the strip down the map.
type StripStyle string

const (
	StripRoad  StripStyle = "road"  // pairs with stones (POR)
	StripCliff StripStyle = "cliff" // a left and right face with a foot now and then (CotAB)
	StripRiver StripStyle = "river" // banks, water and fords, when the engine says so (PoD)
)

// StripGen is the road, cliff or river that runs down the map, leaning with
// the perspective.
type StripGen struct {
	Style  StripStyle  `json:"style"`
	Chance []FlagValue `json:"chance"` // percent by flag; 0 where none (river: always, if BuildGround.River)
	Tiles  []Ground    `json:"tiles"`  // road: pair, stone pair; cliff: left, right, foot, foot2 (the first two + 1d2); river: bank, water, ford, ford end
}

// GroveGen is POR's trees: chance in 100 by flag (1 where none), a size roll
// per square: 1d10 over 8 is 5 or 6, else (roll+1)/2 (RoughRoll: 1d3+3).
type GroveGen struct {
	Default   int         `json:"default"` // where no flag
	Chance    []FlagValue `json:"chance"`
	RoughFlag TerrainFlag `json:"roughFlag"` // flags with it roll RoughRoll
	Base      Ground      `json:"base"`      // + size: the trunk
	Top       Ground      `json:"top"`       // + size: over it, for sizes under 5
}

// TreeGen is the trees of CotAB and PoD: chance in 100 = Base + the
// flags' values (under 0: 1); a Bare flag has none. A hit is a shrub or,
// then, a tree (Top over Trunk), each + 1dN.
type TreeGen struct {
	Base   int         `json:"base"`
	Adjust []FlagValue `json:"adjust"`
	Bare   TerrainFlag `json:"bare"`
	Shrub  Ground      `json:"shrub"`
	ShrubN int         `json:"shrubN"`
	Top    Ground      `json:"top"`
	Trunk  Ground      `json:"trunk"`
	TreeN  int         `json:"treeN"`
}

// PlantGen is POR's plants: 1d(Die) against the running sum of Odds (a
// tree or a lone one, bushes, ...); the flags set Odds entries.
type PlantGen struct {
	Die   int         `json:"die"`
	Odds  [5]int      `json:"odds"`
	Flags []PlantFlag `json:"flags"`
	Tiles PlantTiles  `json:"tiles"`
}

// PlantFlag sets Odds[Index] to Value where the terrain has Flag.
type PlantFlag struct {
	Flag  TerrainFlag `json:"flag"`
	Index int         `json:"index"`
	Value int         `json:"value"`
}

// PlantTiles are the plant grounds: Tall is a tall plant of 1d4 (4: two
// squares high, Top over Bottom), Low + 1d3, Mid + 1d4, Bush + (1d10-1)/3 + 1, Fruit
// + 1d4 (a 4 there is 1d3 unless Rough).
type PlantTiles struct {
	Tall, Top, Bottom, Low, Mid, Bush, Fruit Ground
	RoughFlag                                TerrainFlag
}

// ScatterGen is the small things on open ground of CotAB and PoD: 1d(Die)
// against the running sum of the Kinds' percent of the odds row whose
// roughness range holds the terrain's (Rough: its start and the flags' values).
type ScatterGen struct {
	Rough []FlagValue   `json:"rough"`
	Start int           `json:"start"`
	Kinds []ScatterKind `json:"kinds"`
	Odds  []ScatterOdds `json:"odds"`
}

// ScatterKind is a ground Base + 1dN (N 0: Base alone).
type ScatterKind struct {
	Base Ground `json:"base"`
	N    int    `json:"n"`
}

// ScatterOdds are the percents of the Kinds where the roughness is Lo..Hi.
type ScatterOdds struct {
	Lo      int   `json:"lo"`
	Hi      int   `json:"hi"`
	Percent []int `json:"percent"`
}

// PorGroundGen is POR's (ovr11:0810..11B4).
var PorGroundGen = GroundGen{
	Corners: CornerTiles{NWDoorOpen: 0xD, NEWallOpen: 3, NEDoorOpen: 1},
	Clutter: &ClutterGen{Tile: 0x16, Rare: [2]Ground{0x1A, 0x1B}, Remains: [2]Ground{0x1C, 0x1D}},
	Wild: &WildGen{
		Grass: 0x17, Open: 0x16, Terrain: TerrainFromMap, BossHides: true,
		Strip: &StripGen{Style: StripRoad,
			Chance: []FlagValue{{TerrainMountain, 0x23}, {TerrainHills, 0x4B}},
			Tiles:  []Ground{0x32, 0x33, 0x34, 0x35}},
		Groves: &GroveGen{Default: 1,
			Chance:    []FlagValue{{TerrainSparse, 1}, {TerrainThin, 3}, {TerrainRough, 4}, {TerrainForest, 7}, {TerrainBare, 0}},
			RoughFlag: TerrainRough, Base: 0x23, Top: 0x1F},
		Plants: &PlantGen{Die: 0xFF, Odds: [5]int{0, 6, 0xF, 0x28, 0},
			Flags: []PlantFlag{
				{TerrainSparse, 2, 0x2D}, {TerrainSparse, 3, 0xA},
				{TerrainThin, 2, 0x19},
				{TerrainRough, 3, 5}, {TerrainRough, 4, 0x1E},
				{TerrainBare, 0, 0xF}, {TerrainBare, 3, 0x19}, {TerrainBare, 2, 0xA}},
			Tiles: PlantTiles{Tall: 0x3C, Top: 0x41, Bottom: 0x40, Low: 0x36, Mid: 0x29, Bush: 0x38, Fruit: 0x2D, RoughFlag: TerrainRough}},
	},
}

// terrainTable is kinds as the flags by terrain number: the first range that holds it.
func terrainTable(kinds []struct {
	Lo, Hi Terrain
	kind   TerrainFlag
}) []TerrainFlag {
	t := make([]TerrainFlag, 256)
	for n := range t {
		for _, k := range kinds {
			if Terrain(n) >= k.Lo && Terrain(n) <= k.Hi {
				t[n] = k.kind
				break
			}
		}
	}
	return t
}

// groveTopSizes: a grove of size up to this has a top over its trunk.
const groveTopSizes = 4

// crownOver: the generator puts top over bottom as one two-square thing (a
// grove's top over its trunk of the same size, a tree's top over a trunk, a
// tall plant's Top over its Bottom). Nil w (no wild ground): never.
func (w *WildGen) crownOver(top, bottom Ground) bool {
	in := func(g, base Ground, n int) bool { return g > base && g <= base+Ground(n) }
	switch {
	case w == nil:
		return false
	case w.Groves != nil && in(top, w.Groves.Top, groveTopSizes) && bottom == w.Groves.Base+(top-w.Groves.Top):
		return true
	case w.Trees != nil && in(top, w.Trees.Top, w.Trees.TreeN) && in(bottom, w.Trees.Trunk, w.Trees.TreeN):
		return true // top and trunk roll 1d(TreeN) each: any top over any trunk
	case w.Plants != nil && top == w.Plants.Tiles.Top && bottom == w.Plants.Tiles.Bottom:
		return true
	}
	return false
}
