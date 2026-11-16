package engine

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"gobox/adventure"
	"gobox/importer"
	"gobox/rules"
)

// The old engine's goldens (buckrogers-go, branch golden/one-rng, patched to gobox's one generator): see
// rulesxxvc/testdata/oldengine/README. The data is plain JSON; these tests need only BR1's originals for the
// import (they skip without), never the old engine.

const goldenDir = "../rulesxxvc/testdata/oldengine"

func readGolden(t *testing.T, name string, into any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(goldenDir, name))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, into); err != nil {
		t.Fatal(name, err)
	}
}

// Where the old engine keeps what the shared engine's inputs are: the dump's goldenPlace.
type goldenPlace struct {
	Name    string `json:"name"`
	Block   int    `json:"block"`
	Kind    int    `json:"kind"`    // the area kind byte (TravelKind)
	Terrain int    `json:"terrain"` // WildTerrain
	Map     int    `json:"map"`     // the GEO block loaded, 0 none
	X       int    `json:"x"`
	Y       int    `json:"y"`
	Face    int    `json:"face"`
}

// The views the engine sets (viewMode): 1 the 3D view, 3 a wilderness with the map's column shifted (combatsetup).
// goldenDisks is the number of BR1's data sets.
const goldenDisks = 6

// goldenSeed is the dice state the dump seeded (rules.RNG State).
const goldenSeed = 20240601

const (
	goldenView3D       = 1
	goldenViewWildness = 3
)

// goldenAt puts the engine where the golden's place says: the 3D map, the kind and terrain cells, the party's
// square and facing, the view. The ground and the placement take their inputs from there only.
func (e *Engine) goldenAt(t *testing.T, p goldenPlace) {
	t.Helper()
	e.gs.areaKind = byte(p.Block)
	e.viewMode, e.world.In3D = goldenViewWildness, 0
	e.gs.geo = rules.Map{}
	if p.Map != 0 {
		// The data set (disk) that holds the block, as the old engine's dump looked it up: the first with a map of its size.
		for e.set = 1; e.set <= goldenDisks; e.set++ {
			if len(e.gdata().geoBlock(e.set.File("GEO"), p.Map)) == geoSize {
				break
			}
		}
		e.LoadGeo(p.Map)
		e.viewMode, e.world.In3D = goldenView3D, 1
	}
	for name, v := range map[string]int{fieldTravelKind: p.Kind, fieldWildTerrain: p.Terrain} {
		r, ok := e.layoutField(name)
		if !ok {
			t.Fatalf("no %s in the layout", name)
		}
		e.host.Set(r, int64(v))
	}
	e.posX, e.posY, e.facing = p.X, p.Y, p.Face
}

type goldenArena struct {
	goldenPlace
	Seed uint32   `json:"seed"`
	Rows []string `json:"rows"`
	Maps string   `json:"sides"`
}

// goldenSideBytes is what the dump records per map square: WallAt for the directions 0, 2, 4, 6, then the door
// state (the old engine's passable), then the cell: 9 bytes, squares by y then x.
const goldenSideBytes = 9

var goldenDirs = [4]int{0, 2, 4, 6}

// TestBR1GoldenMap: the 3D map the engine hands the ground is the one the old engine's walls came from.
func TestBR1GoldenMap(t *testing.T) {
	var arenas []goldenArena
	readGolden(t, "arenas.json", &arenas)
	e := br1Begun(t)
	for _, a := range arenas {
		if a.Map == 0 {
			continue
		}
		e.goldenAt(t, a.goldenPlace)
		sides, err := hex.DecodeString(a.Maps)
		if err != nil {
			t.Fatal(err)
		}
		for y := range 16 {
			for x := range 16 {
				g := sides[(y*16+x)*goldenSideBytes:]
				for i, d := range goldenDirs {
					if w := e.WallAt(x, y, d); w != g[i] {
						t.Errorf("%s: WallAt(%d,%d,%d) = %d, old engine %d", a.Name, x, y, d, w, g[i])
					}
					if w := e.DoorAt(x, y, d); w != g[4+i] {
						t.Errorf("%s: DoorAt(%d,%d,%d) = %d, old engine passable %d", a.Name, x, y, d, w, g[4+i])
					}
				}
				if c := e.MapCell(x, y); c != g[8] {
					t.Errorf("%s: MapCell(%d,%d) = %#x, old engine %#x", a.Name, x, y, c, g[8])
				}
			}
		}
	}
}

// TestBR1GoldenArena: the ground BuildGround lays from the engine's inputs, cell by cell against the old engine's
// fight map (cmap.go buildMap: a tile byte a cell, 25 rows of 50).
func TestBR1GoldenArena(t *testing.T) {
	var arenas []goldenArena
	readGolden(t, "arenas.json", &arenas)
	e := br1Begun(t)
	for _, a := range arenas {
		e.goldenAt(t, a.goldenPlace)
		e.do(rules.SetSeed{Seed: a.Seed})
		e.combatMapSetup()
		v, err := e.ruleset().Fact(rules.NoChar, rules.GroundFact{})
		if err != nil {
			t.Fatal(a.Name, err)
		}
		g := v.(rules.Ground)
		diff, first := 0, ""
		for y, row := range a.Rows {
			b, err := hex.DecodeString(row)
			if err != nil || len(b) != len(g.Squares[y]) {
				t.Fatalf("%s row %d: %v, %d cells", a.Name, y, err, len(b))
			}
			for x, want := range b {
				if got := g.Squares[y][x].Tile; got != int(want) {
					if diff++; first == "" {
						first = fmt.Sprintf("(%d,%d) = %d, old engine %d", x, y, got, want)
					}
				}
			}
		}
		if diff > 0 {
			t.Errorf("%s: %d of %d cells differ; first %s", a.Name, diff, len(a.Rows)*50, first)
		}
	}
}

type goldenUnit struct {
	Name   string `json:"name"`
	Mon    int    `json:"mon"`
	Size   int    `json:"size"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Facing int    `json:"facing"`
	Gone   bool   `json:"gone"`
}

type goldenPlacement struct {
	Place    string       `json:"place"`
	Members  int          `json:"members"`
	Group    string       `json:"group"`
	Approach int          `json:"approach"`
	Operand  int          `json:"distance"`
	Dist     int          `json:"monsterDist"` // what the fight is placed with
	Units    []goldenUnit `json:"units"`
}

// goldenSecondMember is the shipped character the dump's second member was loaded from (CHRDATA2.SAV): several
// shipped characters share a name, so the file names it.
const goldenSecondMember = "CHRDATA2.json"

// goldenJoin adds the imported character in file to the party through the engine's own join; its name must be
// the golden's.
func goldenJoin(t *testing.T, e *Engine, file, name string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(e.adv, adventure.DataDir, importer.CharactersDir, file))
	if err != nil {
		t.Fatal(err)
	}
	id, err := e.ruleset().Import(b)
	if err != nil {
		t.Fatal(file, err)
	}
	if got := e.ruleset().Sheet(id).Name; got != name {
		t.Fatalf("%s is %q, the golden's second member %q", file, got, name)
	}
	e.join(id)
}

// goldenMonsterIcon is the icon block LoadMonster names: the picture is not a placement input.
const goldenMonsterIcon = 32

// TestBR1GoldenPlacement: the squares and facings PlaceFight gives the party and the monsters, unit by unit
// against the old engine's place() (cmap.go), for the golden's places, party sizes, groups, approach masks and
// distances. A monster the old engine left off the map is one the engine's Left removes from the party list.
func TestBR1GoldenPlacement(t *testing.T) {
	var cases []goldenPlacement
	readGolden(t, "placement.json", &cases)
	var places []goldenArena
	readGolden(t, "arenas.json", &places)
	at := map[string]goldenPlace{}
	for _, p := range places {
		at[p.Name] = p.goldenPlace
	}
	var groups map[string][][2]int
	readGolden(t, "groups.json", &groups)

	deviations := 0
	fixtures := goldenFixtures{}
	for _, c := range cases {
		e := fixtures.ready(t, c.Members, c.Units)
		e.goldenFight(t, at[c.Place], groups[c.Group], c.Approach, c.Dist)
		v, err := e.ruleset().Fact(rules.NoChar, rules.UnitsFact{})
		if err != nil {
			t.Fatal(err)
		}
		placed := map[rules.CharID]rules.Unit{}
		for _, u := range v.(rules.Units) {
			placed[u.ID] = u
		}
		ids := e.partyIDs()
		name := fmt.Sprintf("%s members=%d %s approach=%d dist=%d", c.Place, c.Members, c.Group, c.Approach, c.Dist)
		// The old engine's units in list order, those it left off the map dropped: the engine's list after Left.
		var want []goldenUnit
		for _, u := range c.Units {
			if !u.Gone {
				want = append(want, u)
			}
		}
		if len(ids) != len(want) {
			t.Errorf("%s: %d units in the party list, old engine %d placed", name, len(ids), len(want))
			continue
		}
		// The old engine's deviation: its placement test (Combat.Who, the last unit on a square wins, and the unit
		// being placed is last) cannot see a unit already on a square, so a team's grid laid again for a second pass
		// from one side (monsterApproach with more monsters than sides) puts two units on one square. The Pascal
		// sees them (cbtmap.pas BuildArenaArrays: cellCombatant of the units placed so far; ProbeDirection: a slot
		// but the unit's own is a combatant there). From the first such unit on the old engine's squares are not the
		// original's; the units before it are compared, and the engine's must never share a square.
		upto := len(want)
		seen := map[[2]int]bool{}
		for i, w := range want {
			if seen[[2]int{w.X, w.Y}] {
				upto = i
				break
			}
			seen[[2]int{w.X, w.Y}] = true
		}
		overlap := map[rules.BoardPos]bool{}
		for i, id := range ids {
			u := placed[id]
			if overlap[u.At] {
				t.Errorf("%s: unit %d %q shares square %d,%d", name, i, u.Name, u.At.X, u.At.Y)
			}
			overlap[u.At] = true
			if i >= upto {
				deviations++
				continue
			}
			w := want[i]
			if u.Name != w.Name || u.Size != w.Size || u.At.X != w.X || u.At.Y != w.Y || u.Facing != w.Facing {
				t.Errorf("%s: unit %d %q size %d at %d,%d facing %d; old engine %q size %d at %d,%d facing %d",
					name, i, u.Name, u.Size, u.At.X, u.At.Y, u.Facing, w.Name, w.Size, w.X, w.Y, w.Facing)
			}
		}
	}
	t.Logf("%d units not compared: the old engine placed them on a taken square (Combat.Who)", deviations)
}

// goldenFixtures is one engine per party size, each with its members only: ready gives the engine back with the last
// case's monsters gone (the rules' state reloaded).
type goldenFixtures struct {
	engines map[int]*Engine
	saved   map[int][]byte
}

// ready is an engine with members members: the first is the roster's, the second (units[1].Name) the imported
// character of that name.
func (g *goldenFixtures) ready(t *testing.T, members int, units []goldenUnit) *Engine {
	t.Helper()
	if g.engines == nil {
		g.engines, g.saved = map[int]*Engine{}, map[int][]byte{}
	}
	e := g.engines[members]
	if e == nil {
		e = br1Begun(t)
		if members == 2 {
			goldenJoin(t, e, goldenSecondMember, units[1].Name)
		}
		var err error
		if g.saved[members], err = e.ruleset().Save(); err != nil {
			t.Fatal(err)
		}
		g.engines[members] = e
	}
	if err := e.ruleset().Load(g.saved[members]); err != nil { // the members only: the last case's monsters go
		t.Fatal(err)
	}
	e.clearTreasure()
	return e
}

// goldenFight is a fight begun as combatSetup does it for the golden's inputs, but for the music: the monsters
// loaded, the ground laid, everyone placed, FightPlaced.
func (e *Engine) goldenFight(t *testing.T, p goldenPlace, group [][2]int, approach, dist int) {
	t.Helper()
	e.goldenAt(t, p)
	e.do(rules.SetSeed{Seed: goldenSeed})
	e.run.monsterSides, e.world.MonsterDist = byte(approach), uint16(dist)
	for _, m := range group {
		e.addMonsters(m[0], m[1], goldenMonsterIcon)
	}
	e.ruleset().BeginFight(e.fight(), nil)
	e.combatMapSetup()
	e.placeCombatants()
	if _, err := e.ruleset().Resolve(e.fight(), rules.FightPlaced{}, nil); err != nil {
		t.Fatal(err)
	}
}
