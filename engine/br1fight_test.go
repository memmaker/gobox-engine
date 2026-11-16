package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"testing"

	"gobox/adventure"
	"gobox/game"
	"gobox/importer"
	"gobox/rules"
)

// br1FightCap bounds the steps a played fight may take.
const br1FightCap = 1500

// br1Fought runs block 201 (ClearMonsters, LoadMonster(41, 2, 32), Combat) on a fresh BR1 game and plays the fight
// as a front would: the unit whose turn it is attacks a target the state offers, else walks toward the nearest foe,
// else ends its turn. It returns the engine and the battle log; stop is what ended the play ("" the fight is over).
func br1Fought(t *testing.T, choose func(game.Fighting) game.Action) (e *Engine, log []string, stop string) {
	t.Helper()
	return br1FoughtIn(t, br1AdjacentFight, 0, choose)
}

// br1AdjacentFight is block 201's body: two monsters set up next to the party.
const br1AdjacentFight = "    ClearMonsters()\n    LoadMonster(41, count=2, icon=32)\n    Combat()"

// br1FoughtIn is br1Fought for a block whose load entry is body, the party facing facing (0: north, 2: east).
func br1FoughtIn(t *testing.T, body string, facing int, choose func(game.Fighting) game.Action) (e *Engine, log []string, stop string) {
	t.Helper()
	e = br1Begun(t)
	br1Join(t, e)
	e.facing = facing // SetupMonster counts the free squares ahead of the party (the start square's walls)
	br1OwnBlock(t, e, 201, body)
	var trace []string // the board at each of the party's turns: the fight's events, in order
	defer func() {
		if r := recover(); r != nil {
			stop = fmt.Sprintf("panic: %v\n%s", r, debug.Stack())
			log = trace
		}
	}()
	for range br1FightCap {
		s := e.build()
		f, ok := s.(game.Fighting)
		if !ok { // the fight is over: the results, the treasure, then the block's move entry asks again
			if strings.HasPrefix(paneText(commonOf(s)), moveText) {
				return e, trace, ""
			}
			br1AfterFight(e, s)
			continue
		}
		switch {
		case myTurn(s):
			trace = append(trace, br1Trace(f))
			if _, err := e.doNow(choose(f)); err != nil {
				return e, trace, fmt.Sprintf("the turn's action refused: %v", err)
			}
		case e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	return e, trace, "the fight did not end"
}

// br1Trace is the board as one line: the round, whose turn, and each unit's place and hit points.
func br1Trace(f game.Fighting) string {
	line := fmt.Sprintf("round %d turn %d:", f.Field.Round, f.Field.Turn)
	for _, u := range f.Field.Units {
		line += fmt.Sprintf(" %s@%d,%d hp%d", u.Name, u.At.X, u.At.Y, u.HP)
	}
	return line
}

// br1AfterFight answers what follows a won fight as a front would: the treasure is taken for the first member,
// then Leave; any other page is continued.
func br1AfterFight(e *Engine, s game.State) {
	if l, ok := s.(game.Looting); ok && !playAllows(s, game.ContinueAction{}) { // the announcement first
		for _, g := range l.Items {
			if _, err := e.doNow(game.Take{Item: g.Handle}); err != nil {
				break
			}
		}
		if _, err := e.doNow(game.Leave{}); err != nil {
			panic(fmt.Sprintf("Leave: %v; allowed %+v; items %d money %v", err, l.Allowed.Plain, len(l.Items), l.Money))
		}
		return
	}
	press(e)
}

// br1Join adds the roster's second character to the party through the engine's own join (as TestBR1PickMember).
func br1Join(t *testing.T, e *Engine) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(e.adv, adventure.DataDir, importer.CharactersDir, "*.json"))
	if err != nil || len(files) < 2 {
		t.Fatal("no second imported character", err)
	}
	b, err := os.ReadFile(files[1])
	if err != nil {
		t.Fatal(err)
	}
	id, err := e.ruleset().Import(b)
	if err != nil {
		t.Fatal(err)
	}
	e.join(id)
}

// br1TurnAction is the turn's choice: an Attack the state offers, else a move to the offered square nearest a foe
// that is nearer than the unit stands, else EndTurn.
func br1TurnAction(f game.Fighting) game.Action {
	for _, a := range f.Allowed.Plain {
		if _, ok := a.(game.Attack); ok {
			return a
		}
	}
	var me rules.BoardPos
	var foes []rules.BoardPos
	for _, u := range f.Field.Units {
		if u.ID == f.Field.Turn {
			me = u.At
		} else if u.Foe && !u.Dead && !u.Gone { // Foe: against the unit the player acts for
			foes = append(foes, u.At)
		}
	}
	near := func(p rules.BoardPos) int {
		d := 1 << 30
		for _, q := range foes {
			d = min(d, max(rules.Abs(p.X-q.X), rules.Abs(p.Y-q.Y)))
		}
		return d
	}
	best, bestD := rules.BoardPos{}, near(me)
	for _, m := range f.Allowed.Moves {
		if d := near(m.At); d < bestD {
			best, bestD = m.At, d
		}
	}
	if bestD < near(me) {
		return game.MoveTo{To: best}
	}
	return game.EndTurn{}
}

// TestBR1Fight: a script's fight on the shared engine with the XXVc rules (buckrogers-go TestFight). Block 201 runs
// ClearMonsters, LoadMonster(41, 2, 32), Combat. Combat asks BuildGround (the dungeon arena from the block, the
// place's traits and the walls) and PlaceFight (units sized by their icon dimensions, the monsters from the sides
// SetupMonster named); the state is Fighting, played through the engine's public actions.
func TestBR1Fight(t *testing.T) {
	e := br1Begun(t)
	br1Join(t, e)
	members := e.partySize()
	br1OwnBlock(t, e, 201, "    ClearMonsters()\n    LoadMonster(41, count=2, icon=32)\n    Combat()")
	ids := e.partyIDs()
	if len(ids) != members+2 {
		t.Fatalf("party of %d, want %d members and 2 monsters", len(ids), members)
	}
	for _, id := range ids[members:] {
		if e.ruleset().Sheet(id).Name == "" || e.fact(id, rules.FactSide) != rules.SideMonsters {
			t.Errorf("monster %d: %+v", id, e.ruleset().Sheet(id))
		}
	}
	g := e.groundTerms() // what the engine hands the ruleset's ground: the block, the place's traits, the wilderness terrain
	kind := byte(br1Cell(t, e, "TravelKind"))
	if g.Block != 201 || g.Place != adventure.BRPlaceTraits(kind) || g.Terrain != rules.TerrainID(br1Cell(t, e, "WildTerrain")) {
		t.Errorf("BuildGround %+v: want block 201, traits of area kind %#x, the WildTerrain cell", g, kind)
	}
	if _, ok := e.build().(game.Fighting); !ok {
		t.Fatalf("after Combat: %T, want game.Fighting", e.build())
	}

	e, log, stop := br1Fought(t, br1TurnAction)
	// Where the play stops is reported, not hidden: the fight is played to its end or this says what stops it.
	if stop != "" {
		t.Fatalf("the fight did not play to its end: %s\nbattle log so far: %q", stop, log)
	}
	if _, ok := e.build().(game.Fighting); ok {
		t.Error("still Fighting after the play")
	}
	if len(log) < 2 {
		t.Errorf("the fight left a trace of %d turns", len(log))
	}
	if got := len(e.partyIDs()); got != 2 {
		t.Errorf("after the fight the party list has %d, want the 2 members: the monsters are gone", got)
	}
	t.Logf("fight trace (%d turns): %q", len(log), log)
}

// TestBR1FightSeed: the same seed gives the same fight twice.
func TestBR1FightSeed(t *testing.T) {
	_, a, stopA := br1Fought(t, br1TurnAction)
	_, b, stopB := br1Fought(t, br1TurnAction)
	if stopA != stopB || !slices.Equal(a, b) {
		t.Errorf("two plays differ:\n%q (%s)\n%q (%s)", a, stopA, b, stopB)
	}
}

// TestBR1QuickFight (buckrogers-go TestQuickFight): the party computer-played (the Quick the state offers, at each turn); the fight ends and the monsters are gone.
func TestBR1QuickFight(t *testing.T) {
	e, trace, stop := br1Fought(t, func(f game.Fighting) game.Action {
		for _, a := range f.Allowed.Plain {
			if q, ok := a.(game.Quick); ok {
				return q
			}
		}
		return game.EndTurn{}
	})
	if stop != "" {
		t.Fatalf("the quick fight did not play to its end: %s\ntrace so far: %q", stop, trace)
	}
	if got := len(e.partyIDs()); got != e.partySize() || got != 2 {
		t.Errorf("after the quick fight the party list has %d (party size %d), want the 2 members", got, e.partySize())
	}
}
