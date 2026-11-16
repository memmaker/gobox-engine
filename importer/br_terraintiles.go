package importer

// brTerrainTiles is the arena tile each combat terrain draws: gdconsts.pas combatTerrain[t].tile for t = 1..LAST_COMBAT_TERRAIN
// (154), at index t. Index 0 is the tile of terrain 21, which CbtMap.DrawArenaCell draws for a cell holding 0 or a value past
// LAST_COMBAT_TERRAIN (adventure.ArtNames.TerrainTiles reads any terrain outside the table as index 0).
var brTerrainTiles = []int{14, 0, 1, 2, 3, 4, 5, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 49, 50, 51, 52, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 34, 35, 36, 37, 38, 39, 40, 41, 42, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 34, 35, 36, 37, 38, 39, 40, 41, 42, 30, 31, 42, 33, 34, 35, 36, 37, 38, 39, 40, 41, 32, 43, 44}
