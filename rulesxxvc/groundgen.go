package rulesxxvc

import (
	"fmt"
	"image"
	"math"

	"gobox/rules"
)

// c_terr.pas: the arena's unset ground and the profile flags.
const (
	unsetTerrain       rules.TerrainID = 6    // UNSET_TERRAIN: wilderness ground not yet given a tile; as a pairTile, no pair
	profPairBelow                      = 0x01 // PROFILE_PAIR_BELOW: the pair tile goes below, not right
	profOneRoll                        = 0x02 // PROFILE_ONE_ROLL: tiles rolled once, for the first plot
	profPairFollows                    = 0x04 // PROFILE_PAIR_FOLLOWS: the pair takes the tile's roll
	profDoubleVariance                 = 0x08 // PROFILE_DOUBLE_VARIANCE
	profKeepCenter                     = 0x10 // PROFILE_KEEP_CENTER
	profFixedScatter                   = 0x20 // PROFILE_FIXED_SCATTER: exactly scatter cells from the centre
)

// terrainProfile is types.pas TTerrainProfile.
type terrainProfile struct {
	chance, diceCount, diceSides, clearance, scatter, scatterBonus int
	flags                                                          int
	tile                                                           rules.TerrainID
	pairTile                                                       rules.TerrainID
	tileVariance, pairVariance                                     int
}

// terrainProfiles is gdconsts.pas terrainProfiles[1..22], in order (index = profile - 1).
var terrainProfiles = [22]terrainProfile{
	{50, 6, 12, 0, 12, 12, 0, 112, unsetTerrain, 3, 0},
	{50, 4, 10, 0, 30, 0, 0, 112, unsetTerrain, 3, 0},
	{80, 4, 10, 0, 10, 10, 0, 115, unsetTerrain, 3, 0},
	{90, 3, 8, 0, 30, 0, 0, 115, unsetTerrain, 3, 0},
	{80, 3, 8, 0, 12, 0, profPairBelow, 118, 130, 0, 0},
	{80, 4, 10, 0, 40, 0, profPairBelow, 119, 127, 4, 3},
	{80, 4, 20, 0, 10, 10, profPairBelow + profOneRoll, 119, 127, 8, 3},
	{80, 2, 12, 0, 20, 0, 0, 104, unsetTerrain, 0, 0},
	{90, 4, 10, 0, 10, 5, profOneRoll, 102, unsetTerrain, 6, 0},
	{80, 2, 5, 0, 8, 10, 0, 110, 111, 0, 0},
	{80, 2, 10, 1, 40, 0, profPairFollows + profDoubleVariance, 69, 70, 5, 0},
	{80, 3, 8, 0, 15, 20, 0, 79, unsetTerrain, 3, 0},
	{80, 4, 12, 0, 15, 20, 0, 82, unsetTerrain, 7, 0},
	{100, 10, 20, 0, 9, 4, 0, 91, unsetTerrain, 0, 0},
	{80, 5, 10, 0, 10, 10, profPairBelow + profOneRoll, 100, 98, 2, 2},
	{100, 1, 6, 0, 5, 0, profKeepCenter, 92, 93, 0, 0},
	{80, 1, 10, 1, 5, 10, 0, 92, 93, 0, 0},
	{80, 5, 10, 0, 15, 15, profOneRoll, 94, unsetTerrain, 4, 0},
	{100, 10, 8, 0, 0, 0, profKeepCenter + profFixedScatter, 79, unsetTerrain, 10, 0},
	{80, 3, 10, 0, 10, 10, profPairBelow + profOneRoll, 137, 135, 2, 2},
	{80, 3, 10, 0, 15, 15, profOneRoll, 131, unsetTerrain, 4, 0},
	{80, 10, 10, 0, 15, 10, 0, 154, unsetTerrain, 0, 0},
}

// profileRun is one BuildWilderness(profile, count) of a terrain's recipe.
type profileRun struct{ profile, count int }

// wildRecipes is combatst.pas BuildWildernessArena's case: the runs of each wilderness terrain, in order. Terrains
// 14 and 21 also run a loop of three runs with a scatter override first (wildRecipe14).
var wildRecipes = map[int][]profileRun{
	1:  {{6, 6}, {1, 8}, {3, 1}, {4, 2}, {5, 2}, {8, 2}},
	8:  {{6, 6}, {1, 8}, {3, 1}, {4, 2}, {5, 2}, {8, 2}},
	9:  {{6, 6}, {1, 8}, {3, 1}, {4, 2}, {5, 2}, {8, 2}},
	2:  {{1, 6}, {3, 1}, {4, 1}, {5, 1}, {7, 5}, {8, 1}},
	3:  {{2, 1}, {4, 3}, {7, 10}, {10, 1}, {9, 5}, {22, 1}},
	4:  {{10, 10}, {2, 1}, {3, 1}, {4, 1}, {7, 3}, {9, 10}, {22, 2}},
	5:  {{10, 10}, {2, 1}, {3, 1}, {4, 1}, {7, 10}, {9, 15}, {22, 3}},
	6:  {{1, 1}, {2, 2}, {3, 1}, {4, 1}, {7, 2}, {8, 1}, {9, 1}},
	7:  {{1, 1}, {2, 2}, {3, 1}, {4, 1}, {7, 2}, {8, 1}, {9, 1}},
	11: {{13, 2}, {14, 1}, {16, 1}, {14, 2}, {18, 3}, {12, 2}},
	20: {{13, 2}, {14, 1}, {16, 1}, {14, 2}, {18, 3}, {12, 2}},
	12: {{11, 4}, {12, 4}, {13, 2}, {18, 2}, {15, 12}},
	18: {{11, 4}, {12, 4}, {13, 2}, {18, 2}, {15, 12}},
	13: {{11, 3}, {12, 3}, {13, 2}, {14, 2}, {18, 4}, {15, 8}},
	19: {{11, 3}, {12, 3}, {13, 2}, {14, 2}, {18, 4}, {15, 8}},
	14: {{20, 4}, {21, 4}, {12, 1}, {13, 1}},
	21: {{20, 4}, {21, 4}, {12, 1}, {13, 1}},
	15: {{12, 2}, {13, 4}, {17, 3}, {18, 7}, {11, 1}},
	16: {{13, 1}, {17, 1}, {18, 8}},
	17: {{11, 2}, {12, 4}, {13, 2}, {18, 4}, {15, 6}},
}

// Terrain 14 and 21 run profiles 14, 16, 19 six times, each with a scatter override of d4 + 2 (rolled first), before
// their recipe.
const (
	forestRuns         = 6
	forestScatterBonus = 2
	forestScatterDice  = 4
	loopTerrainA       = 14
	loopTerrainB       = 21
)

// Terrain below firstHighland takes ground 108 + d2 - 1; 15 and 16 take 91; the others 89 + d2 - 1
// (BuildWildernessArena's end).
const (
	highlandTerrain  = 10
	groundLow        = 108
	groundHighland   = 89
	groundLowSpread  = 2
	groundHighSpread = 2
	groundDunes      = 91
	duneTerrainA     = 15
	duneTerrainB     = 16
)

// wildArena is the arena's cells while a wilderness is laid: GDVars.combatArena^.cells[row][col].
type wildArena struct {
	cells    [arenaRows][arenaCols]rules.TerrainID
	rng      *rules.RNG
	override int // GDVars.scatterOverride: > 0 a forced scatter
	centerX  int // GDVars.randomCenterX, Y
	centerY  int
}

func (w *wildArena) dice(n, sides int) int { return rules.RNGDice(w.rng, n, sides) }

// randomReal is Turbo Pascal's Random: the advanced seed over 2^32.
func (w *wildArena) randomReal() float64 {
	rules.Random(&w.rng.State, 0)
	return float64(w.rng.State) / (1 << 32)
}

// open is IsTerrainOpen: an unset cell or ground 89, 90, 91, 108, 109. x and y arrive as the ShortInts the original
// passes: past 127 they wrap negative.
func (w *wildArena) open(x, y int) bool {
	x, y = int(int8(x)), int(int8(y))
	if x < 0 || x >= arenaCols || y < 0 || y >= arenaRows {
		return false
	}
	switch t := w.cells[y][x]; t {
	case unsetTerrain, 89, 90, 91, 108, 109:
		return true
	}
	return false
}

// build is combatst.pas BuildWilderness(profile, count).
func (w *wildArena) build(profile, count int) {
	p := terrainProfiles[profile-1]
	for range count {
		if !PercentileRoll(w.rng, p.chance) {
			continue
		}
		remaining := w.dice(p.diceCount, p.diceSides)
		first := true
		scatter := p.scatter
		if p.flags&profKeepCenter == 0 {
			w.centerX = w.dice(1, arenaCols) - 1
			w.centerY = w.dice(1, arenaRows) - 1
		}
		if p.scatterBonus > 0 {
			scatter += w.dice(1, p.scatterBonus)
		}
		if w.override > 0 {
			scatter = w.override
		}
		var tile, pair rules.TerrainID
		var cx, cy int // Bytes
		for ; remaining > 0; remaining-- {
			if first || p.flags&profOneRoll == 0 {
				variance := 0
				if p.tileVariance > 0 {
					variance = w.dice(1, p.tileVariance) - 1
				}
				if p.flags&profDoubleVariance != 0 {
					variance *= 2
				}
				tile = p.tile + rules.TerrainID(variance)
				pair = p.pairTile
				if p.pairTile != unsetTerrain {
					if p.flags&profPairFollows != 0 {
						pair += rules.TerrainID(variance)
					} else {
						variance = 0
						if p.pairVariance > 0 {
							variance = w.dice(1, p.pairVariance) - 1
						}
						if p.flags&profDoubleVariance != 0 {
							variance *= 2
						}
						pair += rules.TerrainID(variance)
					}
				}
				first = false
			}
			if scatter > 0 {
				distance := float64(scatter)
				if p.flags&profFixedScatter == 0 {
					distance = float64(w.dice(1, scatter+1) - 1)
				}
				angle := w.randomReal() * 2 * math.Pi
				cx = int(byte(int(math.Cos(angle)*distance) + w.centerX)) // Byte: Trunc into it, Inc wraps
				cy = int(byte(int(math.Sin(angle)*distance) + w.centerY))
			} else {
				cx, cy = w.centerX, w.centerY
			}
			if w.placementOpen(p, cx, cy, pair) {
				w.cells[int8(cy)][int8(cx)] = tile // in range: open said so
				if pair != unsetTerrain {
					if p.flags&profPairBelow != 0 {
						cy = int(byte(cy + 1))
					} else {
						cx = int(byte(cx + 1))
					}
					if cy <= arenaRows-1 && cx <= arenaCols-1 {
						w.cells[cy][cx] = pair
					}
				}
			}
		}
	}
}

// placementOpen is IsPlacementOpen: the cells within clearance of the centre, or the centre and its pair, are open.
func (w *wildArena) placementOpen(p terrainProfile, cx, cy int, pair rules.TerrainID) bool {
	switch {
	case p.clearance > 0:
		all := true
		for dx := -p.clearance; dx <= p.clearance; dx++ {
			for dy := -p.clearance; dy <= p.clearance; dy++ {
				all = all && w.open(cx+dx, cy+dy) // no short cut: IsTerrainOpen has no effect, but the loop is the original's
			}
		}
		return all
	case !w.open(cx, cy):
		return false
	case pair != unsetTerrain:
		if p.flags&profPairBelow != 0 {
			return w.open(cx, cy+1)
		}
		return w.open(cx+1, cy)
	}
	return true
}

// fillUnset is FillUnsetTerrain(base, spread): every unset cell, column by column, gets base plus a roll.
func (w *wildArena) fillUnset(base, spread int) {
	for col := range arenaCols {
		for row := range arenaRows {
			if w.cells[row][col] == unsetTerrain {
				v := base
				if spread > 0 {
					v += w.dice(1, spread) - 1
				}
				w.cells[row][col] = rules.TerrainID(v)
			}
		}
	}
}

// buildWildernessArena is combatst.pas BuildWildernessArena for the wilderness terrain (zone0 wildernessTerrain, 10
// higher when the area kind's low nibble is 8): the cells of the arena, or the error the original shows ("Cant
// identify wild terrain").
func buildWildernessArena(terrain int, rng *rules.RNG) (*[arenaRows][arenaCols]rules.TerrainID, error) {
	w := &wildArena{rng: rng, override: -1}
	for y := range w.cells {
		for x := range w.cells[y] {
			w.cells[y][x] = unsetTerrain
		}
	}
	runs, ok := wildRecipes[terrain]
	if !ok {
		return nil, fmt.Errorf("Cant identify wild terrain %d", terrain)
	}
	if terrain == loopTerrainA || terrain == loopTerrainB {
		for range forestRuns {
			w.override = w.dice(1, forestScatterDice) + forestScatterBonus
			w.build(14, 1)
			w.build(16, 1)
			w.build(19, 1)
		}
		w.override = -1
	}
	for _, run := range runs {
		w.build(run.profile, run.count)
	}
	switch {
	case terrain < highlandTerrain:
		w.fillUnset(groundLow, groundLowSpread)
	case terrain == duneTerrainA || terrain == duneTerrainB:
		w.fillUnset(groundDunes, 0)
	default:
		w.fillUnset(groundHighland, groundHighSpread)
	}
	return &w.cells, nil
}

// buildGround is the rules' half of combatMapSetup (combatst.pas LoadCombatTileSet): the wilderness arena from the
// terrain the game names, drawn from the call's stream (the original draws from the game's Random).
// A dungeon arena (BuildDungeonArena) is not ported: it needs the area kind byte and the barrier plane, which
// BuildGround does not carry.
func (r *Rules) buildGround(g rules.BuildGround, rng *rules.RNG) error {
	if g.Dungeon {
		return fmt.Errorf("BuildGround dungeon: %w (BuildDungeonArena needs the area kind and the barrier plane)", ErrNotPorted)
	}
	cells, err := buildWildernessArena(int(g.Terrain), r.stream(rng))
	if err != nil {
		return err
	}
	r.arena = arena{}
	b := rules.Board{Size: image.Pt(arenaCols, arenaRows), Terrain: make([]rules.TerrainID, arenaCols*arenaRows)}
	for y := range cells {
		copy(b.Terrain[y*arenaCols:], cells[y][:])
	}
	r.ground = b
	return nil
}
