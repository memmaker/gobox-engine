package importcotab

import "gobox/rules1e"

// CombatGround is how CotAB lays its combat ground (rules.json combatGround),
// from the constants of combatsetup.go.
func CombatGround() rules1e.GroundGen {
	return rules1e.GroundGen{
		Corners: rules1e.CornerTiles{NWDoorOpen: 0xD, NEWallOpen: 7, NEDoorOpen: 1},
		Rubble: &rules1e.RubbleGen{CellFlag: cellRubble, Tile: tileOpenGround,
			Heap: groundRubble, Edge: groundRubbleEdge, HeapOdds: 5, EdgeOdds: 9},
		Wild: &rules1e.WildGen{
			Grass: groundGrass, Open: tileOpenGround, Terrain: rules1e.TerrainFromGiven,
			Strip: &rules1e.StripGen{Style: rules1e.StripCliff,
				Chance: []rules1e.FlagValue{{Flag: terrainMountain, Value: 35}, {Flag: terrainHills, Value: 75}},
				Tiles:  []rules1e.Ground{groundCliffLeft, groundCliffRight, groundCliffFoot, groundCliffFoot2}},
			Trees: &rules1e.TreeGen{Base: 10, Bare: terrainBare,
				Adjust: []rules1e.FlagValue{{Flag: terrainSparse, Value: -5}, {Flag: terrainThin, Value: -2},
					{Flag: terrainRough, Value: 5}, {Flag: terrainForest, Value: 10}},
				Shrub: groundShrub, ShrubN: 2, Top: groundTreeTop, Trunk: groundTreeTrunk, TreeN: 5},
			Scatter: &rules1e.ScatterGen{Start: 50,
				Rough: []rules1e.FlagValue{{Flag: terrainHills, Value: 10}, {Flag: terrainMountain, Value: 30},
					{Flag: terrainRough, Value: 20}, {Flag: terrainThin, Value: -10},
					{Flag: terrainSparse, Value: -20}, {Flag: terrainBare, Value: -50}},
				Kinds: []rules1e.ScatterKind{{Base: groundBoulder, N: 2}, {Base: groundBush, N: 2},
					{Base: groundRocks, N: 4}, {Base: groundStones, N: 3}, {Base: groundFlowers, N: 4}},
				Odds: []rules1e.ScatterOdds{
					{Lo: -30, Hi: 9, Percent: []int{0, 0, 0, 30, 15}},
					{Lo: 10, Hi: 29, Percent: []int{0, 1, 5, 20, 10}},
					{Lo: 30, Hi: 69, Percent: []int{0, 2, 5, 10, 5}},
					{Lo: 70, Hi: 89, Percent: []int{10, 2, 10, 10, 1}},
					{Lo: 90, Hi: 110, Percent: []int{15, 5, 15, 10, 1}}}},
		},
	}
}
