package rulesxxvc

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// seamSource serves the tables as the engine's source would: the import's JSON.
type seamSource struct{ tables []byte }

func (seamSource) RulesFile() string             { return "" }
func (seamSource) Name(rules.SourceTable) string { return "" }
func (s seamSource) Read(t rules.SourceTable, v any) bool {
	if t != rules.SourceRuleTables || s.tables == nil {
		return false
	}
	if err := json.Unmarshal(s.tables, v); err != nil {
		panic(err)
	}
	return true
}

// seamHost is a creation host that answers nothing (the builder never calls it).
type seamHost struct{}

func (seamHost) Selected() int                             { return 0 }
func (seamHost) Keep(rules.CharID) error                   { return nil }
func (seamHost) Join(rules.RosterID) (rules.CharID, error) { return rules.NoChar, nil }
func (seamHost) Portrait(rules.Pics) (h, b rules.PicID)    { return }
func (seamHost) IconPics(rules.Pics) (h, b rules.PicID)    { return }
func (seamHost) SetPics(rules.Draft, *rules.Pics)          {}
func (seamHost) QuickParty() []rules.Template              { return nil }

const seamItem rules.ItemID = 1

func seamTables() xxvc.Tables {
	return xxvc.Tables{
		ItemTypes: []xxvc.ItemType{{Slot: xxvc.SlotWeapon, Damage: rules.Dice{N: 1, Sides: 6}}},
		ItemWords: []string{"", "Laser", "Pistol"},
		Items:     map[rules.ItemID]xxvc.ItemTemplate{seamItem: {Type: 0, Words: [3]int32{1, 2, 0}, Weight: 4, Value: 30, Quantity: 1}},
	}
}

func seamRules(t *testing.T) (*Rules, rules.CreationRules) {
	t.Helper()
	b, err := json.Marshal(seamTables())
	if err != nil {
		t.Fatal(err)
	}
	rs, cr := Builder(seamSource{b}, seamHost{})
	return rs.(*Rules), cr
}

// seamFile is a saved character as the importer writes it.
func seamFile(t *testing.T, name string, npc bool) rules.CharacterFile {
	t.Helper()
	c := member(name, 12, 9, 0, 250)
	c.Race, c.Level, c.HP, c.MaxHP, c.Exp, c.Side = xxvc.Martian, 3, 17, 20, 900, rules.SideParty
	c.ArmorClass, c.Movement, c.Damage = [2]int{7, 9}, 12, [2]rules.Dice{{N: 2, Sides: 4, Bonus: 1}}
	c.Items = []Carried{{Ref: rules.ItemRef{ID: seamItem, Count: 1}, Readied: true}}
	if npc {
		c.NonPlayer, c.Morale = true, 40
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func seamJoin(t *testing.T, r *Rules, name string) rules.CharID {
	t.Helper()
	id, err := r.Import(seamFile(t, name, false))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := r.Resolve(nil, rules.JoinParty{Who: id}, nil); err != nil {
		t.Fatal(err)
	}
	return id
}

func TestBuilder(t *testing.T) {
	r, cr := seamRules(t)
	if len(r.tables.Items) != 1 || r.tables.ItemWords[1] != "Laser" {
		t.Errorf("tables did not reach Begin: %+v", r.tables)
	}
	if cr == nil {
		t.Fatal("no creation")
	}
	// The creation is ported (creation.go); a member that is not there cannot be modified.
	if d, err := cr.Draft(rules.NoChar); err != nil || len(cr.Fields(d)) != 5 || cr.QuickParty() != nil {
		t.Errorf("creation: %v %v", d, err)
	}
	if _, err := cr.Draft(1); !errors.Is(err, ErrNoMember) {
		t.Errorf("creation Draft of a member: %v", err)
	}
	defer func() {
		if recover() == nil {
			t.Error("a source without tables did not panic")
		}
	}()
	Builder(seamSource{}, seamHost{})
}

// TestSessionActions is the specification of the start-up and party Actions.
func TestSessionActions(t *testing.T) {
	r, _ := seamRules(t)
	for _, a := range []rules.Action{rules.InitRules{}, rules.LoadItemTable{}, rules.NewAdventure{}, rules.SetSeed{Seed: 5},
		rules.EnterPhase{Phase: rules.PhaseMenu}, rules.ActFor{Member: 0}, rules.SetMemberCount{N: 0}, rules.EmptyParty{}, rules.ForgetParty{}} {
		if _, err := r.Resolve(nil, a, nil); err != nil {
			t.Errorf("%T: %v", a, err)
		}
	}
	if _, err := r.Resolve(nil, rules.Toughen{Who: 9}, nil); err != nil {
		t.Errorf("Toughen: %v", err)
	}
	if _, err := r.Resolve(nil, rules.JoinParty{Who: 99}, nil); err == nil {
		t.Error("JoinParty of no one")
	}
	if _, err := r.Resolve(nil, rules.Intimidate{}, nil); !errors.Is(err, ErrNotPorted) {
		t.Errorf("an action XXVc does not know: %v", err)
	}
	o, err := r.Resolve(nil, rules.EmptyMember{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	first := o.Effects[0].(rules.EmptyMemberIs).Who
	if o, _ = r.Resolve(nil, rules.EmptyMember{}, nil); o.Effects[0].(rules.EmptyMemberIs).Who != first || r.Sheet(first).Name != "" {
		t.Error("the placeholder is one id, emptied")
	}
}

func TestNewAdventureResets(t *testing.T) {
	r, _ := seamRules(t)
	id := seamJoin(t, r, "Buck")
	r.Resolve(nil, rules.SetClock{Part: rules.ClockHours, Value: 7}, nil)
	r.purse = 9
	r.Resolve(nil, rules.NewAdventure{}, nil)
	if r.Check(rules.MemberCount{}, nil).Value != 0 || r.Check(rules.ClockIs{Part: rules.ClockHours}, nil).Value != 0 || r.purse != 0 {
		t.Error("new adventure keeps its state")
	}
	if r.chars[id] == nil || len(r.party) != 1 {
		t.Error("new adventure forgets the party list as well")
	}
}

func TestSeededRolls(t *testing.T) {
	r, _ := seamRules(t)
	r.Resolve(nil, rules.SetSeed{Seed: 12345}, nil)
	direct := rules.RNG{State: 12345}
	for i := 0; i < 5; i++ {
		if got, want := r.Check(rules.RollRandom{N: 100}, nil).Value, int32(direct.N(100)); got != want {
			t.Fatalf("RollRandom %d: %d, want %d", i, got, want)
		}
		if got, want := r.Check(rules.RollDice{N: 3, Sides: 6}, nil).Value, int32(rules.RNGDice(&direct, 3, 6)); got != want {
			t.Fatalf("RollDice %d: %d, want %d", i, got, want)
		}
	}
	rng := rules.RNG{State: 1} // a call that brings its own stream rolls from it
	own := rules.RNG{State: 1}
	if r.Check(rules.RollRandom{N: 10}, &rng).Value != int32(own.N(10)) || rng != own {
		t.Error("the passed stream is not the one rolled from")
	}
}

func TestPartyCountActingClock(t *testing.T) {
	r, _ := seamRules(t)
	a, b := seamJoin(t, r, "Buck"), seamJoin(t, r, "Wilma")
	if got := r.Check(rules.MemberCount{}, nil); !got.Yes || got.Value != 2 {
		t.Errorf("MemberCount %+v", got)
	}
	if got := r.Check(rules.PartySize{}, nil); got.Value != 2 {
		t.Errorf("PartySize %+v", got)
	}
	if got := r.Check(rules.ActingFor{}, nil); !got.Yes || got.Value != 1 {
		t.Errorf("ActingFor %+v: the last joined acts", got)
	}
	r.Resolve(nil, rules.ActFor{Member: 0}, nil)
	if r.Check(rules.ActingFor{}, nil).Value != 0 {
		t.Error("ActFor")
	}
	r.Resolve(nil, rules.SetMemberCount{N: 5}, nil)
	if r.Check(rules.MemberCount{}, nil).Value != 5 || r.Check(rules.PartySize{}, nil).Value != 2 {
		t.Error("the record's count is its own, not the list's length")
	}
	v, err := r.Fact(rules.NoChar, rules.PartyFact{})
	if p := v.(rules.Party); err != nil || len(p) != 2 || p[0] != a || p[1] != b {
		t.Errorf("PartyFact %v %v", v, err)
	}

	r.Resolve(nil, rules.SetClock{Part: rules.ClockHours, Value: 13}, nil)
	r.Resolve(nil, rules.SetClock{Part: rules.ClockTenMinutes, Value: 3}, nil)
	for part, want := range map[rules.ClockPart]int32{rules.ClockHours: 13, rules.ClockTenMinutes: 3, rules.ClockMinuteOfDay: 13*60 + 30} {
		if got := r.Check(rules.ClockIs{Part: part}, nil); !got.Yes || got.Value != want {
			t.Errorf("ClockIs %d: %+v, want %d", part, got, want)
		}
	}
	if _, err := r.Resolve(nil, rules.SetClock{Part: rules.ClockMinuteOfDay, Value: 1}, nil); err == nil {
		t.Error("the minute of the day is read only")
	}
}

// TestEngineQuestions is the specification of what XXVc answers the state build.
func TestEngineQuestions(t *testing.T) {
	r, _ := seamRules(t)
	id := seamJoin(t, r, "Buck")
	npc, err := r.Import(seamFile(t, "Hired", true))
	if err != nil {
		t.Fatal(err)
	}
	r.Resolve(nil, rules.EnterPhase{Phase: rules.PhaseExploring}, nil)

	facts := []struct {
		name string
		f    rules.Fact
		who  rules.CharID
		want rules.Value
	}{
		{"MemberCard", rules.MemberCardFact{}, id, rules.MemberCard{Party: true, Class: "Warrior"}},
		{"PoolPics", rules.PoolPicsFact{}, id, rules.PoolPics{}},
		{"Portrait", rules.PortraitFact{}, id, rules.Portrait{}},
		{"IconPics", rules.IconPicsFact{}, id, rules.IconPics{Size: 1}}, // the small combat icon, body 0
		{"Book", rules.BookFact{}, id, rules.Book{Levels: []int32{3}}},
		{"SpellMenu", rules.SpellMenuFact{}, id, rules.SpellMenu(nil)},
		{"Owed", rules.OwedFact{}, id, rules.PowerOwed{}},
		{"Figures", rules.FiguresFact{}, id, rules.Figures(nil)},
		{"Train", rules.TrainFact{}, id, rules.TrainCard{Why: hallMaxLevel}}, // the seam's tables name no level to train to
		{"Damage", rules.DamageFact{}, id, rules.Text("1d6+6")},              // the readied pistol's dice as RecalcStats has them, not the c.Damage set above
		{"Listed", rules.ListedFact{}, id, rules.Handles{1}},
		{"Word status", rules.WordFact{Of: rules.FactStatus}, id, rules.Text("Okay")},
		{"Word race", rules.WordFact{Of: rules.FactRace}, id, rules.Text("Martian")},
		{"Word class", rules.WordFact{Of: rules.FactClasses}, id, rules.Text("Warrior")},
		{"Word sex", rules.WordFact{Of: rules.FactSex}, id, rules.Text("")},
		{"Word alignment", rules.WordFact{Of: rules.FactAlignment}, id, rules.Text("")},
		{"Word exceptional", rules.WordFact{Of: rules.FactExceptional}, id, rules.Text("")},
		{"Control PC", rules.FactControl, id, rules.Number(0)},
		{"Control hireling", rules.FactControl, npc, rules.Number(0x80 | 40)},
		{"LevelReady", rules.FactLevelReady, id, rules.Yes(false)},
		{"Active", rules.FactActive, id, rules.Yes(true)},
		{"Side", rules.FactSide, id, rules.SideParty},
		{"Credits", rules.CoinFact{Coin: xxvc.Credits}, id, rules.Number(250)},
		{"Purse", rules.PurseFact{}, rules.NoChar, rules.Purse{Copper: 250}}, // the party's: Buck's own; the hireling has not joined
		{"Loot", rules.PurseFact{Loot: true}, rules.NoChar, rules.Purse{}},
		{"ItemCard", rules.ItemCardFact{Item: 1}, id, rules.ItemCard{Item: 1, Name: "Laser Pistol", Count: 1, Price: 30, Sale: rules.Purse{Copper: 15}, Kind: seamItem, Weight: 4,
			Readied: true, Identified: true, Tab: string(rules.KindWeapon), Slot: "weapon"}}, // a readied weapon names its doll slot
	}
	for _, c := range facts {
		got, err := r.Fact(c.who, c.f)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if g, w := jsonOf(t, got), jsonOf(t, c.want); g != w {
			t.Errorf("%s: %s, want %s", c.name, g, w)
		}
	}

	checks := []struct {
		name string
		q    rules.Check
		want rules.Answer
	}{
		{"PoolCap", rules.PoolCap{}, rules.Answer{}},
		{"PoolCount", rules.PoolCount{}, rules.Answer{}},
		{"CarryCount", rules.CarryCount{Who: id}, rules.Answer{Value: 1}},
		{"ItemTableMissing", rules.ItemTableMissing{}, rules.Answer{}},
		{"IsDay", rules.IsDay{}, rules.Answer{}},
		{"SkyDay day", rules.SkyDay{Sky: 3}, rules.Answer{Yes: true}},
		{"SkyDay night", rules.SkyDay{Sky: 1, SkyInside: 3}, rules.Answer{}},
		{"SkyDay inside", rules.SkyDay{Inside: true, Sky: 1, SkyInside: 3}, rules.Answer{Yes: true}},
		{"Fit", rules.Fit{Who: id}, rules.Answer{Yes: true}},
		{"Modifiable", rules.Modifiable{Who: id}, rules.Answer{}},
		{"PaladinCan", rules.PaladinCan{Who: id}, rules.Answer{}},
		{"InFight", rules.InFight{}, rules.Answer{}},
		{"HallOpen", rules.HallOpen{}, rules.Answer{}},
	}
	for _, c := range checks {
		if got := r.Check(c.q, nil); got != c.want {
			t.Errorf("%s: %+v, want %+v", c.name, got, c.want)
		}
	}

	sh := r.Sheet(id)
	if sh.Name != "Buck" || sh.HP != 17 || sh.MaxHP != 20 || sh.AC != 53 || sh.RearAC != 51 || sh.THAC0 != sheetBase || sh.Move != 12 || sh.Exp != 900 ||
		sh.Classes != (rules.ClassSet{xxvc.Warrior}) || sh.Health != rules.Hale || sh.Levels[xxvc.Warrior] != 3 {
		t.Errorf("Sheet %+v", sh)
	}
	if r.Purse(id) != 250 {
		t.Errorf("Purse %d", r.Purse(id))
	}
}

func jsonOf(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestFileFormat: the record is the character's JSON; Pack and Unpack are
// byte-identical, the roster number rides in it, a joined character's file
// imports back.
func TestFileFormat(t *testing.T) {
	r, _ := seamRules(t)
	var ff rules.FileFormat = r
	file := seamFile(t, "Buck", false)
	if ff.RosterOf(rules.SavedRecord(file)) != rules.NoRoster {
		t.Error("a new file has a roster number")
	}
	files, err := ff.Unpack(file, false)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(files.Record, file) || files.Items != nil || files.Portrait != nil {
		t.Error("Unpack: the record is the one file")
	}
	back, err := ff.Pack(files)
	if err != nil || !bytes.Equal(back, file) {
		t.Fatalf("Pack: %v, identical %v", err, bytes.Equal(back, file))
	}
	if _, err := ff.Pack(rules.CharacterFiles{Record: file, Items: []byte{1}}); err == nil {
		t.Error("Pack takes a file the format has none of")
	}
	if _, err := ff.Pack(rules.CharacterFiles{Record: []byte("nope")}); err == nil {
		t.Error("Pack takes a record that is no character")
	}
	numbered := ff.WithRoster(rules.SavedRecord(file), 7)
	if ff.RosterOf(numbered) != 7 {
		t.Error("RosterOf after WithRoster")
	}
	if got := ff.WithRoster(numbered, rules.NoRoster); !bytes.Equal(got, file) {
		t.Error("the roster number is the only change WithRoster makes")
	}
	if lv := ff.SavedLevels(numbered); len(lv) != 1 || lv[0] != 3 {
		t.Errorf("SavedLevels %v", lv)
	}
	// Import, SetRoster, Export, Unpack: what the engine's SaveCharacter writes.
	id, err := r.Import(rules.CharacterFile(numbered))
	if err != nil {
		t.Fatal(err)
	}
	r.Resolve(nil, rules.SetRoster{Who: id, Roster: 9}, nil)
	out, err := r.Export(id)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := ff.Unpack(out, false)
	if err != nil || ff.RosterOf(saved.Record) != 9 {
		t.Errorf("exported roster %d, %v", ff.RosterOf(saved.Record), err)
	}
	id2, _ := r.Import(out)
	if out2, _ := r.Export(id2); !bytes.Equal(out, out2) {
		t.Error("Export is not stable through Import")
	}
}

func TestSessionSaveLoad(t *testing.T) {
	r, _ := seamRules(t)
	id := seamJoin(t, r, "Buck")
	r.Resolve(nil, rules.SetSeed{Seed: 77}, nil)
	r.Resolve(nil, rules.SetClock{Part: rules.ClockHours, Value: 5}, nil)
	b, err := r.Save()
	if err != nil {
		t.Fatal(err)
	}
	q, _ := seamRules(t)
	if err := q.Load(b); err != nil {
		t.Fatal(err)
	}
	if v, _ := q.Fact(rules.NoChar, rules.PartyFact{}); len(v.(rules.Party)) != 1 || v.(rules.Party)[0] != id ||
		q.Check(rules.MemberCount{}, nil).Value != 1 || q.Check(rules.ClockIs{Part: rules.ClockHours}, nil).Value != 5 ||
		q.Check(rules.RollRandom{N: 1000}, nil) != r.Check(rules.RollRandom{N: 1000}, nil) {
		t.Error("the session is not saved whole")
	}
}

// AdvanceClock carries up the units and runs the party's effects down by the minutes passed
// (BR1 pO0E3_0388, pO0E3_0000).
func TestAdvanceClock(t *testing.T) {
	r, _ := seamRules(t)
	id := seamJoin(t, r, "TICK")
	r.chars[id].Effects = []Effect{{ID: 1, Minutes: 25}, {ID: 2, Minutes: 5}}
	for range 12 {
		if _, err := r.Resolve(nil, rules.AdvanceClock{Count: 1, Unit: clockMinute}, nil); err != nil {
			t.Fatal(err)
		}
	}
	if r.clock[clockMinute] != 2 || r.clock[clockTurn] != 1 {
		t.Errorf("clock %v after 12 minutes", r.clock)
	}
	if fx := r.chars[id].Effects; len(fx) != 1 || fx[0].ID != 1 || fx[0].Minutes != 13 {
		t.Errorf("effects %+v, want effect 1 with 13 minutes", fx)
	}
	if _, err := r.Resolve(nil, rules.AdvanceClock{Count: 1, Unit: clockTurn}, nil); err != nil {
		t.Fatal(err)
	}
	if fx := r.chars[id].Effects; r.clock[clockTurn] != 2 || len(fx) != 1 || fx[0].Minutes != 3 {
		t.Errorf("after a turn: clock %v effects %+v", r.clock, fx)
	}
}

// The clock's last carry (years at clockUnitScale 100) ages each party member a year (gametime.pas ValidateClock).
func TestYearsAgeTheParty(t *testing.T) {
	r, _ := seamRules(t)
	id := seamJoin(t, r, "OLD")
	r.chars[id].Age = 30
	r.clock[clockYear] = clockUnitScale[clockYear] - 1
	r.clock[clockMonth] = clockUnitScale[clockMonth] - 1
	if _, err := r.Resolve(nil, rules.AdvanceClock{Count: 1, Unit: clockMonth}, nil); err != nil {
		t.Fatal(err)
	}
	if r.chars[id].Age != 31 || r.clock[clockYear] != 0 {
		t.Errorf("age %d, years %d", r.chars[id].Age, r.clock[clockYear])
	}
}
