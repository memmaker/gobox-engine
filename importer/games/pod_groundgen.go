package games

import "gobox/rules1e"

// CombatGround is how PoD lays its combat ground (rules.json combatGround),
// from the constants of combatsetup.go: a dungeon on open floor, a wilderness
// with no terrain kinds (flags 4 throughout, O065:0723) and the river the
// engine asks for.
func PodCombatGround() rules1e.GroundGen {
	return rules1e.GroundGen{
		Fill:    podTileOpenGround,
		Corners: rules1e.CornerTiles{NWDoorOpen: 0, NEWallOpen: 7, NEDoorOpen: 7},
		Wild: &rules1e.WildGen{
			Grass: wildGrass, Open: wildGrassTile, Terrain: rules1e.TerrainNone,
			Strip: &rules1e.StripGen{Style: rules1e.StripRiver,
				Tiles: []rules1e.Ground{wildBank, wildWater, wildFord, wildFordEnd}},
			Trees: &rules1e.TreeGen{Base: wildTreeChance,
				Shrub: wildBushes, ShrubN: 2, Top: wildTreeTop, Trunk: wildTreeTrunk, TreeN: 5},
			Scatter: &rules1e.ScatterGen{
				Kinds: []rules1e.ScatterKind{{Base: wildRocks, N: 2}, {Base: wildBoulder}, {Base: wildBoulder, N: 5},
					{Base: wildShrub}, {Base: wildFlowers, N: 6}},
				Odds: []rules1e.ScatterOdds{{Percent: []int{0, 2, 1, 14, 5}}}},
		},
	}
}
