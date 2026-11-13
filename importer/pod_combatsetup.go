package importer

// DungeonTiles is how many tiles at the start of DungCom are wall pieces:
// dungeonMap puts tiles 0..0x18 as walls (0x16 the plain wall, 0x18 the highest).
const PodDungeonTiles = 0x18 + 1

// tileOpenGround: the groundInfo tile of open ground (things may be put there).
const podTileOpenGround = 22

// Wilderness battle grounds (raw ground bytes, rows of DS:3EBA).
const (
	wildGrass      = 69 // the map is filled with it; its tile is 37
	wildGrassTile  = 37
	wildBank       = 52 // the river banks; 53 too
	wildWater      = 55 // 56 too
	wildFord       = 58 // 59, 60 too
	wildFordEnd    = 61
	wildBushes     = 41 // 42 too
	wildTreeTop    = 31 // 32..35: over a trunk 36..40
	wildTreeTrunk  = 36
	wildRocks      = 49 // 50 too
	wildBoulder    = 44 // 45..48 too
	wildShrub      = 68
	wildFlowers    = 61 // 62..66 too
	wildTreeChance = 2  // O065:09A1 at the terrain flags 4 (O065:0723)
)
