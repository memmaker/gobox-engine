package engine

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
)

// br1JourneyMoves bounds the steps taken looking for the first fight.
const br1JourneyMoves = 600

// br1Stage logs a journey stage, so a failure says where the walk was.
func br1Stage(t *testing.T, format string, args ...any) {
	t.Helper()
	t.Logf("journey: "+format, args...)
}

// TestBR1Journey: buckrogers-go rules TestJourney on the shared engine, by the actions a front sends: a new character
// made, added, Begin, the opening's pages, the walk to the dying officer, the first ground fight of block 16 won
// on Quick, the loot taken, the DEPOT of block 17 with a buy, a rest, a save, a load into a FRESH engine and the
// game resumed with the same party, credits, square and script state.
func TestBR1Journey(t *testing.T) {
	e := br1Engine(t)
	e.Open("")
	e.titleDone()
	e.Advance()

	// a new character, kept, added, Begin
	if err := e.createCharacter(); err != nil {
		t.Fatal(err)
	}
	if err := e.rename("Wilma"); err != nil {
		t.Fatal(err)
	}
	if err := e.EndCreation(true); err != nil {
		t.Fatal(err)
	}
	list := e.BuildCharList()
	if len(list) == 0 || list[len(list)-1] != "Wilma" {
		t.Fatalf("roster %v", list)
	}
	if _, err := e.doNow(game.AddCharacter{Roster: len(list) - 1}); err != nil {
		t.Fatal(err)
	}
	if e.partySize() != 1 || commonOf(e.build()).Party[0].Name != "Wilma" {
		t.Fatalf("team: %d", e.partySize())
	}
	if _, err := e.doNow(game.Begin{}); err != nil {
		t.Fatal(err)
	}
	br1Pages(t, e) // the intro
	x, ok := e.build().(game.Exploring)
	if !ok || x.Where.X != 5 || x.Where.Y != 6 || x.Map.Walk == nil {
		t.Fatalf("after the intro: %T %+v", e.build(), x.Where)
	}

	if playAllows(x, game.MakeCamp{}) { // explore.pas MapMainMenu's dungeon menu: Move, Area, View, Save, Change, Look; the rest is Program(9)'s
		t.Fatal("BR's dungeon menu offers a camp")
	}

	// the dying officer
	if err := br1North(e); err != nil {
		t.Fatal(err)
	}
	if c := commonOf(e.build()); c.Where.Y != 5 || !strings.HasPrefix(paneText(c), "AN OFFICER IS COLLAPSED") {
		t.Fatalf("officer: %+v %q", c.Where, paneText(c))
	}
	br1Pages(t, e)
	if x, ok := e.build().(game.Exploring); !ok || x.Where.Y != 5 {
		t.Fatalf("opening: %T", e.build())
	}
	br1Stage(t, "the opening done at %+v", commonOf(e.build()).Where)

	// the first fight of block 16, walked into: back and forth along the corridor, the step entry's chance of an
	// encounter (area16 L_87BE) takes its time
	br1Pages(t, e)
	fought := false
	rng := &rules.RNG{State: 5}
	for n := 0; n < br1JourneyMoves && !fought; n++ {
		switch s := e.build().(type) {
		case game.Fighting:
			fought = true
		case game.Exploring:
			if rng.N(3) == 0 {
				e.doNow(game.Face{Facing: rng.N(4) * 2})
			}
			br1North(e)
		default:
			_ = s
			br1Answer(t, e)
		}
	}
	if !fought {
		t.Fatalf("no fight in %d moves: %T", br1JourneyMoves, e.build())
	}
	br1Stage(t, "in a fight at %+v, block %d", commonOf(e.build()).Where, e.gs.areaKind)
	// the party's credits are what its members hold: a NEO warrior killed in the fight takes his with him, Wilma's are her own
	purse := func() rules.Copper {
		return rules.Copper(e.fact(e.partyIDs()[0], rules.PurseFact{}).(rules.Purse).Copper)
	}
	for n := 0; ; n++ {
		if n > br1FightCap {
			t.Fatalf("fight stuck: %T %q", e.build(), paneText(commonOf(e.build())))
		}
		s := e.build()
		if _, ok := s.(game.Fighting); !ok {
			break
		}
		switch {
		case myTurn(s):
			if _, err := e.doNow(game.Quick{}); err != nil {
				t.Fatal(err)
			}
		case e.waiting(PendFight):
			e.Step()
		default:
			press(e)
		}
	}
	if _, over := e.build().(game.Assembling); over {
		t.Fatal("the party lost the fight")
	}
	// the treasure: the announcement, then every item taken
	carried := func() (n int) {
		for _, id := range e.partyIDs() {
			n += len(e.listedItems(id))
		}
		return n
	}
	had, cash0 := carried(), purse()
	var money int
	offered := 0
	for n := 0; ; n++ {
		l, ok := e.build().(game.Looting)
		if !ok && (offered > 0 || n > 20) {
			break
		}
		if !ok { // the results' pages
			br1Answer(t, e)
			continue
		}
		if n > 20 {
			t.Fatalf("loot stuck: %+v", l.Allowed.Plain)
		}
		if playAllows(l, game.ContinueAction{}) { // "You found some loot!"
			br1Answer(t, e)
			continue
		}
		offered, money = len(l.Items), l.Money[game.Credits]
		if _, err := e.doNow(game.Take{All: true}); err != nil {
			t.Fatal(err)
		}
		if _, err := e.doNow(game.Leave{}); err != nil {
			t.Fatal(err)
		}
	}
	if offered == 0 || carried() != had+offered || len(e.goods()) != 0 {
		t.Fatalf("loot: %d offered, party carries %d (had %d), %d left", offered, carried(), had, len(e.goods()))
	}
	if purse() != cash0+rules.Copper(money) {
		t.Fatalf("credits %d, were %d with %d offered", purse(), cash0, money)
	}
	if _, ok := e.build().(game.Exploring); !ok {
		t.Fatalf("after the loot: %T %q", e.build(), paneText(commonOf(e.build())))
	}
	br1Stage(t, "fight won, %d items taken, credits %d", offered, purse())

	square := commonOf(e.build()).Where
	// save; a FRESH engine loads and resumes
	wa := commonOf(e.build())
	if _, ok := e.build().(game.Exploring); !ok || wa.Where.X != square.X || wa.Where.Y != square.Y {
		t.Fatalf("back: %T %+v", e.build(), wa.Where)
	}
	blob, err := e.Save()
	if err != nil {
		t.Fatal(err)
	}
	b := br1Loaded(t, blob)
	wb := commonOf(b.build())
	if _, ok := b.build().(game.Exploring); !ok {
		t.Fatalf("loaded into %T %q", b.build(), paneText(wb))
	}
	if !reflect.DeepEqual(wa.Where, wb.Where) {
		t.Errorf("square %+v, loaded %+v", wa.Where, wb.Where)
	}
	if len(wb.Party) != len(wa.Party) || wb.Party[0].Name != "Wilma" || !reflect.DeepEqual(wa.Purse, wb.Purse) {
		t.Errorf("party %d, loaded %d (%q); purse %v, loaded %v", len(wa.Party), len(wb.Party), wb.Party[0].Name, wa.Purse, wb.Purse)
	}
	for i := range wa.Party {
		if wa.Party[i].Name != wb.Party[i].Name || wa.Party[i].HP != wb.Party[i].HP || len(wa.Party[i].Items) != len(wb.Party[i].Items) {
			t.Errorf("member %d: %q hp %d items %d, loaded %q hp %d items %d", i, wa.Party[i].Name, wa.Party[i].HP, len(wa.Party[i].Items), wb.Party[i].Name, wb.Party[i].HP, len(wb.Party[i].Items))
		}
	}
	if wb.Ask != nil || paneText(wb) != "" {
		t.Errorf("the game resumed on a question: %q %+v", paneText(wb), wb.Ask)
	}
	for _, name := range []string{fieldScriptBlock, "MsgDelay", "PrevX", "PrevY", "Word4BFB", "Word4BFD", "Word4BFE", "EntryBytes", "Word4C2E"} {
		ra, ok := e.layoutField(name)
		if !ok {
			continue
		}
		rb, _ := b.layoutField(name)
		if e.host.Get(ra) != b.host.Get(rb) {
			t.Errorf("script state %s %d, loaded %d", name, e.host.Get(ra), b.host.Get(rb))
		}
	}
	if e.world.Scripted != b.world.Scripted || e.gs.areaKind != b.gs.areaKind {
		t.Errorf("block %d scripted %d, loaded %d/%d", e.gs.areaKind, e.world.Scripted, b.gs.areaKind, b.world.Scripted)
	}
	if err := br1North(b); err != nil { // the game goes on
		t.Fatal(err)
	}
	br1Stage(t, "saved, loaded into a fresh engine, resumed at %+v", wb.Where)

	// a rest on the loaded game: BR's dungeon menu has no camp, its rest is the script's Program(9) (explore.pas
	// RestMenu); a block of the test's own calls it: Rest once, Exit
	hp := func() int { return int(b.selFact(rules.FactHP)) }
	b.setSelFact(rules.FactHP, int64(hp()-1))
	was := hp()
	br1OwnBlock(t, b, 202, "    Program(9)\n    Text(\"AFTER THE REST\")\n    PressKey()")
	if m := commonOf(b.build()).Ask; m == nil || !slices.Equal(m.Replies, []string{"Rest", "Exit"}) {
		t.Fatalf("the rest menu: %+v", m)
	}
	if _, err := b.doNow(game.Reply{N: restMenuRest}); err != nil {
		t.Fatal(err)
	}
	if hp() != was+1 {
		t.Fatalf("rest: hp %d, was %d", hp(), was)
	}
	if _, err := b.doNow(game.Reply{N: restMenuExit}); err != nil {
		t.Fatal(err)
	}
	if got := paneText(commonOf(b.build())); !strings.HasPrefix(got, "AFTER THE REST") {
		t.Fatalf("after the rest: %q", got)
	}

	// the DEPOT of block 17 (the story's route to it is long: DevAt), on the loaded game: one buy
	bpurse := func() rules.Copper {
		return rules.Copper(b.fact(b.partyIDs()[0], rules.PurseFact{}).(rules.Purse).Copper)
	}
	br1ToBaseMenu(t, b)
	if _, err := b.doNow(game.Reply{N: 2}); err != nil { // Depot
		t.Fatal(err)
	}
	for n := 0; n < 10; n++ {
		if _, ok := b.build().(game.Trading); ok {
			break
		}
		br1Answer(t, e)
	}
	tr, ok := b.build().(game.Trading)
	if !ok || tr.Kind != game.TradeShop || len(tr.Goods) == 0 {
		t.Fatalf("no shop: %T %q", b.build(), paneText(commonOf(b.build())))
	}
	good, cash, items := tr.Goods[0], bpurse(), len(b.listedItems(b.partyIDs()[0]))
	if rules.Copper(good.Price) > cash {
		t.Fatalf("a %d credit good, %d credits", good.Price, cash)
	}
	if _, err := b.doNow(game.Buy{Member: 0, Item: good.Handle}); err != nil {
		t.Fatal("buy:", err)
	}
	if got := bpurse(); got != cash-rules.Copper(good.Price) || len(b.listedItems(b.partyIDs()[0])) != items+1 {
		t.Fatalf("after the buy: credits %d (had %d, price %d), items %d (had %d)", got, cash, good.Price, len(b.listedItems(b.partyIDs()[0])), items)
	}
	if _, err := b.doNow(game.Leave{}); err != nil {
		t.Fatal("leave:", err)
	}
	for n := 0; n < 10 && !br1AtBaseMenu(b); n++ {
		br1Answer(t, b)
	}
	if !br1AtBaseMenu(b) {
		t.Fatalf("after the shop: %T %q", b.build(), paneText(commonOf(b.build())))
	}
	br1Stage(t, "bought %s for %d, credits %d", good.Name, good.Price, bpurse())

}
