package engine

import (
	"reflect"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

func br1Walls(t *testing.T, e *Engine) game.MapWalls {
	t.Helper()
	x, ok := e.build().(game.Exploring)
	if !ok {
		t.Fatalf("%T", e.build())
	}
	return x.Map.Walls
}

// TestBR1WallsBackdrop: the tarmac (block 16, set 1, walls 7 then 127: no BACK1 picture 255) shows the
// city's flats without a picture, as the old front did (buckrogers-go rules TestClassStreetOutside); in
// set 4 with BACK4 block 10 loaded last, outdoors (cell without bit 7: block 138) the picture and the mars
// clouds' flats show (rules TestBackdrop, TestBackFlats), indoors (bit 7: block 10) the haze's.
func TestBR1WallsBackdrop(t *testing.T) {
	e := br1Begun(t)
	if w := br1Walls(t, e); w.Floor != "br1_city" || w.Sky != "sky_br1_city" || w.Backdrop != nil {
		t.Errorf("tarmac: floor %q sky %q backdrop %v", w.Floor, w.Sky, w.Backdrop != nil)
	}
	e.set = 4
	e.LoadWalldef(4, 1)
	e.LoadWalldef(10, 2)
	w := br1Walls(t, e)
	if w.Floor != "br1_mars_clouds" || w.Sky != "sky_br1_mars_clouds" || w.Backdrop == nil || w.Backdrop.Block != 138 || w.Backdrop.File != "BACK4" {
		t.Errorf("set 4 outdoors: floor %q sky %q backdrop %+v", w.Floor, w.Sky, w.Backdrop)
	}
	e.gs.geo.Plane[0x200+e.posY<<4+e.posX] |= 0x80
	w = br1Walls(t, e)
	if w.Floor != "br1_mars_haze" || w.Backdrop == nil || w.Backdrop.Block != 10 {
		t.Errorf("set 4 indoors: floor %q backdrop %+v", w.Floor, w.Backdrop)
	}
}

// TestBR1PlaceMusic: Where.Music is the block's place (the old front's musicPlace): Neo-Chicago on the tarmac
// (block 16), Salvation in the base's menu (block 17), whose streets are a town.
func TestBR1PlaceMusic(t *testing.T) {
	e := br1Begun(t)
	if w := commonOf(e.build()).Where; w.Music != "neo-chicago" || w.Mode == "town" {
		t.Errorf("block 16: music %q mode %q", w.Music, w.Mode)
	}
	br1ToBaseMenu(t, e)
	if w := commonOf(e.build()).Where; w.Music != "salvation" || w.Mode != "town" {
		t.Errorf("block 17: music %q mode %q", w.Music, w.Mode)
	}
}

// TestBR1ScriptSound: the intro's Sound(5) (in a For loop) reaches the front as the cue "5", Sound(131) as none (the old engine's
// command 67 reports the numbers below 128).
func TestBR1ScriptSound(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	var heard []string
	note := func(out game.Outcome) {
		for _, ev := range out.Events {
			if s, ok := ev.(game.Sounded); ok {
				heard = append(heard, s.Sound)
			}
		}
	}
	for _, a := range []game.Action{game.AddCharacter{Roster: 0}, game.Begin{}} {
		out, err := e.doNow(a)
		if err != nil {
			t.Fatalf("%T: %v", a, err)
		}
		note(out)
	}
	for n := 0; commonOf(e.build()).Ask != nil && n < 80; n++ {
		note(br1AnswerOut(t, e))
	}
	for _, h := range heard { // Sound(5) is in the intro's For loop, once a pass
		if h != "5" {
			t.Errorf("the intro sounded %q", h)
		}
	}
	if len(heard) == 0 {
		t.Error("the intro sounded nothing")
	}
}

// TestBR1DelayMs: a script's Delay pauses 100 ms a unit of the game speed (BR textrend.pas PauseOnly), none at speed 0.
func TestBR1DelayMs(t *testing.T) {
	for _, c := range []struct{ speed, ms int }{{3, 300}, {0, 0}} {
		e := br1Begun(t)
		e.msgDelay = byte(c.speed)
		br1OwnBlock(t, e, 201, "    Delay()")
		got := 0
		if e.eclWait != nil {
			got = e.eclWait.ms
		}
		if got != c.ms {
			t.Errorf("speed %d: pause %d ms, want %d", c.speed, got, c.ms)
		}
	}
}

// TestBR1ViewKeptAcrossBlocks: a block whose start does not set the view kind keeps the overland (BR's EclInit leaves the
// interface state alone; the Pool's InitAreaScript sets the 3D view, which a Pool game still gets).
func TestBR1ViewKeptAcrossBlocks(t *testing.T) {
	e := br1Begun(t)
	if e.world.In3D != 1 {
		t.Fatalf("a new game's view: In3D %d, want the 3D view", e.world.In3D)
	}
	e.world.In3D = 0 // the overland
	e.InitAreaScript()
	if e.world.In3D != 0 {
		t.Errorf("a block's start set In3D %d over the overland", e.world.In3D)
	}
	v := *e.eclVersion           // a copy: the dialect is shared
	v.Behaviour.ViewKept = false // as the Pool's
	e.eclVersion = &v
	e.InitAreaScript()
	if e.world.In3D != 1 {
		t.Errorf("without ViewKept the start leaves In3D %d", e.world.In3D)
	}
}

// TestBR1OverlandMove mirrors buckrogers-go TestOverlandMove: a block whose script sets the overland view kind
// (AreaKind 4) is the overland, not by its number; the party's square is MapX, MapY (Area[195], [196]) over 0..37 by
// 0..14, a step costs twelve hours, and the edge holds the party (the clock still runs, as the old overland() did).
func TestBR1OverlandMove(t *testing.T) {
	e := br1Begun(t)
	br1OwnBlock(t, e, 201, "    AreaKind(4, 255)")
	br1Pages(t, e) // the template's move entry
	if e.world.In3D != 0 || e.viewMode != overlandViewMode {
		t.Fatalf("In3D %d view mode %d, want the overland", e.world.In3D, e.viewMode)
	}
	x, ok := e.build().(game.Exploring)
	if !ok || !x.Map.Overland {
		t.Fatalf("%T overland %v scripted %d phase %v top %v ask %v", e.build(), ok && x.Map.Overland, e.world.Scripted, e.phase, e.top(), commonOf(e.build()).Ask)
	}
	e.world.MapX, e.world.MapY, e.facing = 10, 5, 2 // a step east through the front's Arrive
	hour := e.clockPart(rules.ClockMinuteOfDay) / 60
	if _, err := e.doNow(game.RunEvent{Before: true, Side: 2}); err != nil { // the template's move entry asks a key
		t.Fatal(err)
	}
	br1Pages(t, e)
	if _, err := e.doNow(game.Arrive{From: rules.MapPos{X: 10, Y: 5}, To: rules.MapPos{X: 11, Y: 5}, Facing: e.facing}); err != nil || e.world.MapX != 11 {
		t.Fatalf("Arrive east: square %d,%d err %v pre %+v exploring %v top %v over %v view %v inv %v", e.world.MapX, e.world.MapY, err, e.ext.pre, e.exploring(), e.top(), e.over, e.view.on, e.snap.invOn)
	}
	if h := e.clockPart(rules.ClockMinuteOfDay) / 60; h != (hour+12)%24 {
		t.Errorf("hour %d after %d, a step costs 12 hours", h, hour)
	}
	e.world.MapX = 37
	hour = e.clockPart(rules.ClockMinuteOfDay) / 60
	e.mapMove()
	if e.world.MapX != 37 || e.world.MapY != 5 {
		t.Errorf("square %d,%d, want 37,5", e.world.MapX, e.world.MapY)
	}
	if h := e.clockPart(rules.ClockMinuteOfDay) / 60; h != (hour+12)%24 {
		t.Errorf("hour %d after %d, want 12 later", h, hour)
	}
	e.facing = 0 // north
	e.world.MapY = 0
	e.mapMove()
	if e.world.MapY != 0 {
		t.Errorf("y %d north of the map", e.world.MapY)
	}
	e.world.MapX, e.world.MapY, e.facing = 36, 14, 2
	e.mapMove()
	if e.world.MapX != 37 {
		t.Errorf("x %d, want 37", e.world.MapX)
	}
	e.facing = 4 // south, off the 15 rows
	e.mapMove()
	if e.world.MapY != 14 {
		t.Errorf("y %d south of the map", e.world.MapY)
	}
	if w := commonOf(e.build()).Where; w.X != 37 || w.Y != 14 {
		t.Errorf("Where %d,%d, want the overland's square 37,14", w.X, w.Y)
	}
}

// TestBR1RestMenu: Program(9) opens the rest menu (explore.pas RestMenu) and holds the script, which goes on after Exit.
// Rest is a day: every okay member one hit point up (and the clock a day on) (gametime.pas RestAndHeal(1), the same as
// rulesxxvc.RestAndHeal).
func TestBR1RestMenu(t *testing.T) {
	e := br1Begun(t)
	hp := func() int { return int(e.selFact(rules.FactHP)) }
	e.setSelFact(rules.FactHP, int64(hp()-3))
	was := hp()
	br1OwnBlock(t, e, 202, "    Program(9)\n    Text(\"AFTER THE REST\")\n    PressKey()")
	menu := func() []string {
		a := commonOf(e.build()).Ask
		if a == nil {
			return nil
		}
		return a.Replies
	}
	if m := menu(); len(m) != 2 || m[0] != "Rest" || m[1] != "Exit" {
		t.Fatalf("the rest menu: %q", m)
	}
	if _, err := e.doNow(game.Reply{N: restMenuRest}); err != nil {
		t.Fatal(err)
	}
	if hp() != was+1 { // the day on the clock (AdvanceClock) has no layout name to read here
		t.Errorf("after a rest: hp %d (was %d)", hp(), was)
	}
	if m := menu(); len(m) != 2 {
		t.Fatalf("the menu is not asked again: %q", m)
	}
	if _, err := e.doNow(game.Reply{N: restMenuExit}); err != nil {
		t.Fatal(err)
	}
	if got := paneText(commonOf(e.build())); !strings.HasPrefix(got, "AFTER THE REST") {
		t.Errorf("the script did not go on after Exit: %q", got)
	}
}

// br1Loaded is a fresh BR1 engine at the party menu that loads blob (Engine.Save's).
func br1Loaded(t *testing.T, blob []byte) *Engine {
	t.Helper()
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if err := e.Load(blob); err != nil {
		t.Fatal(err)
	}
	br1Settle(e)
	return e
}

// br1Settle lets the engine's own steps run until it waits for a decision.
func br1Settle(e *Engine) { e.Advance() }

// TestBR1SaveLoadExploring: a game saved while exploring block 16 after the intro, loaded into a fresh engine, goes on at the
// same square with the same party, purse, script cells, clock and view, and does not play the intro again.
func TestBR1SaveLoadExploring(t *testing.T) {
	a := br1Begun(t)
	blob, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	b := br1Loaded(t, blob)
	wa, wb := commonOf(a.build()), commonOf(b.build())
	if _, ok := b.build().(game.Exploring); !ok {
		t.Fatalf("loaded into %T %q, want exploring", b.build(), paneText(wb))
	}
	if !reflect.DeepEqual(wa.Where, wb.Where) {
		t.Errorf("where %+v, loaded %+v", wa.Where, wb.Where)
	}
	if wb.Ask != nil || paneText(wb) != "" {
		t.Errorf("the intro plays again: %q %+v", paneText(wb), wb.Ask)
	}
	if len(wa.Party) == 0 || len(wa.Party) != len(wb.Party) || wa.Party[0].Name != wb.Party[0].Name || wa.Party[0].HP != wb.Party[0].HP {
		t.Errorf("party %+v, loaded %+v", wa.Party, wb.Party)
	}
	if !reflect.DeepEqual(wa.Purse, wb.Purse) {
		t.Errorf("purse %v, loaded %v", wa.Purse, wb.Purse)
	}
	for _, name := range []string{fieldScriptBlock, "MsgDelay", "LastX", "LastY", "Word4BFB", "Word4BFD", "Word4BFE"} {
		ra, ok := a.layoutField(name)
		if !ok {
			t.Fatalf("no %s", name)
		}
		rb, _ := b.layoutField(name)
		if a.host.Get(ra) != b.host.Get(rb) {
			t.Errorf("%s %d, loaded %d", name, a.host.Get(ra), b.host.Get(rb))
		}
	}
	if a.world.Scripted != b.world.Scripted || a.world.In3D != b.world.In3D || a.clockPart(rules.ClockMinuteOfDay) != b.clockPart(rules.ClockMinuteOfDay) {
		t.Errorf("scripted %d/%d In3D %d/%d clock %d/%d", a.world.Scripted, b.world.Scripted, a.world.In3D, b.world.In3D,
			a.clockPart(rules.ClockMinuteOfDay), b.clockPart(rules.ClockMinuteOfDay))
	}
}

// TestBR1SaveLoadScriptedPlace: a game saved in block 17 (a scripted place: the base's menu) loads back into its menu, the
// walls of the view the same.
func TestBR1SaveLoadScriptedPlace(t *testing.T) {
	a := br1Begun(t)
	br1ToBaseMenu(t, a)
	if !br1AtBaseMenu(a) {
		t.Fatal("not at the base's menu")
	}
	blob, err := a.Save()
	if err != nil {
		t.Fatal(err)
	}
	b := br1Loaded(t, blob)
	if !br1AtBaseMenu(b) {
		c := commonOf(b.build())
		t.Fatalf("loaded into %T %q ask %+v", b.build(), paneText(c), c.Ask)
	}
	if b.gs.areaKind != 17 || b.world.Scripted != a.world.Scripted {
		t.Errorf("block %d scripted %d, want 17 and %d", b.gs.areaKind, b.world.Scripted, a.world.Scripted)
	}
}
