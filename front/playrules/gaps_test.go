package playrules

import (
	"fmt"
	"image"
	"image/color"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// walker is fake's map: an open 3x3 square with a wall east of (0,0).
type walker struct{}

func (walker) Name() string      { return "yard" }
func (walker) Size() image.Point { return image.Pt(3, 3) }
func (walker) Edge(at image.Point, side int) game.Edge {
	if at == (image.Point{}) && side == 2 {
		return game.Edge{Kind: game.EdgeWall}
	}
	return game.Edge{}
}
func (walker) Departure(image.Point, int) bool { return false }
func (walker) Event(image.Point) bool          { return false }

func allow(a ...play.Action) play.Common { return play.Common{Allowed: play.Allowed{Plain: a}} }

// TestGaps: each closed gap's state shows as the old State, and its
// decision reaches Do as the play action.
func TestGaps(t *testing.T) {
	door := &game.Door{Locked: true, Bash: true}
	intro := &game.Intro{Full: true}
	res := &game.Results{Outcome: "won"}
	man := game.TurnManeuver{ID: 3, Name: "Sweep"}
	sf := game.SpaceFight{Status: "Closing", Range: 4}
	cases := []struct {
		name string
		st   play.State
		d    game.Decision
		sent play.Action // nil: nothing reaches Do
		ok   func(s game.State) bool
	}{
		{"door", play.Exploring{Door: door, Day: true, Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{}}}, Outside: true, Sun: true, Mountains: true}}},
			game.TryDoor{Way: 1}, play.TryDoor{Way: play.DoorBash},
			func(s game.State) bool { return s.Door == door && s.Walls.Sun && s.Walls.Mountains && s.Flags.Day }},
		{"night sky", play.Exploring{Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{}}}, Outside: true}}}, game.Search{}, play.Search{},
			func(s game.State) bool { return !s.Walls.Sun && !s.Flags.Day }},
		{"vault", play.Trading{Kind: play.TradeVault, Vault: &game.Vault{Money: map[string]int{"gold": 5}}}, game.Leave{}, play.Leave{},
			func(s game.State) bool { return s.Mode == game.ModeVault && s.Vault.Money["gold"] == 5 }},
		{"shop identify", play.Trading{Identify: 20, Unknown: []game.ItemRef{{Member: 0, Item: 1}}}, game.Identify{Item: 1}, play.Identify{Item: 1},
			func(s game.State) bool { return s.Shop.IdentifyPrice == 20 && len(s.Shop.Unknown) == 1 }},
		{"training learn", play.Trading{Kind: play.TradeTraining, Training: &game.Training{Learning: true}}, game.Learn{Spell: 4}, play.Learn{Spell: 4},
			func(s game.State) bool { return s.Mode == game.ModeTraining && s.Training.Learning }},
		{"handout esc leaves", play.Scene{Kind: play.SceneHandout, Handout: "Journal", Layers: []game.Layer{{}}, Common: allow(play.Leave{})}, game.Cancel{}, play.Leave{},
			func(s game.State) bool { return s.Read != nil && len(s.Page.Layers) == 1 && s.Holding == game.HoldNone }},
		{"handout read out", play.Scene{Kind: play.SceneHandout, Common: allow(play.Leave{})}, game.Answer{N: 2}, play.Leave{}, nil},
		{"story page", play.Scene{Text: []string{"hi"}, Common: allow(play.Continue{})}, game.Answer{}, play.Continue{},
			func(s game.State) bool {
				r := s.Page.Pane[0][0]
				return s.Holding == game.HoldNone && r.Upcase && r.Fg == (color.RGBA{})
			}},
		{"forget entry", play.Assembling{Common: allow(play.OpenPicker{Kind: 5})}, game.OpenPicker{Kind: 5}, play.OpenPicker{Kind: 5},
			func(s game.State) bool {
				return slices.ContainsFunc(s.Pending.Options, func(o game.Option) bool { return o.Label == "Forget Character" })
			}},
		{"rest heals", play.Camping{Heals: true, Common: allow(play.AdvanceTimeBy{})}, game.LeaveCamp{}, play.BreakCamp{},
			func(s game.State) bool { return s.RestHeals && !s.RestStudies }},
		{"speaker over view", play.Exploring{Common: play.Common{Portrait: &play.Portrait{}}}, game.Search{}, play.Search{},
			func(s game.State) bool { return s.Window == game.WindowPicture }},
		{"continue over the view", play.Exploring{Common: play.Common{Ask: &play.Question{Continue: true}, Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Answer{}, play.Continue{},
			func(s game.State) bool { return s.Mode == game.ModeExplore && s.Pending.Kind == game.Continue }},
		{"open training", play.Exploring{}, game.OpenTraining{Member: 1}, play.OpenTraining{Member: 1}, nil},
		{"loot detect", play.Looting{Status: true, Common: allow(play.LootDetect{}, play.LootShare{})}, game.LootDetect{}, play.LootDetect{},
			func(s game.State) bool { return s.Loot.Status && s.Allowed["LootDetect"] && s.Allowed["LootShare"] }},
		{"loot share", play.Looting{}, game.LootShare{}, play.LootShare{}, nil},
		{"loot status", play.Looting{}, game.LootStatus{}, play.LootStatus{}, nil},
		{"power pick", play.Camping{Common: play.Common{PowerPick: &game.PowerPick{Name: "Bob"}}}, game.PickPower{ID: 2}, play.PickPower{ID: 2},
			func(s game.State) bool { return s.PowerPick != nil && s.PowerPick.Name == "Bob" }},
		{"use power", play.Camping{}, game.UsePower{Member: 1, Kind: 2}, play.UsePower{Member: 1, Kind: 2}, nil},
		{"policy", play.Camping{}, game.SetPolicy{ID: 1, Policy: 2}, play.SetPolicy{ID: 1, Policy: 2}, nil},
		{"mark", play.Exploring{}, game.Mark{At: image.Pt(1, 2), Text: "x"}, play.Mark{At: image.Pt(1, 2), Text: "x"}, nil},
		{"unmark", play.Exploring{}, game.Unmark{At: image.Pt(1, 2)}, play.Unmark{At: image.Pt(1, 2)}, nil},
		{"rename", play.Assembling{}, game.Rename{Name: "Al"}, play.Rename{Name: "Al"}, nil},
		{"quit", play.Over{Text: "The end"}, game.Quit{}, play.Quit{},
			func(s game.State) bool { return s.LastText == "The end" }},
		{"modify", play.Assembling{}, game.Modify{Member: 2}, play.Modify{Member: 2}, nil},
		{"quick start", play.Assembling{}, game.QuickStart{}, play.QuickStart{}, nil},
		{"roll", play.Assembling{Draft: &play.Draft{Step: "skills", Choices: []string{"a"}}}, game.Roll{Minimal: true}, play.Reroll{Minimal: true},
			func(s game.State) bool { return s.Creation.Step == "skills" && len(s.Creation.Choices) == 1 }},
		{"plan", play.Camping{}, game.Plan{Member: 1, Spell: 2, Count: 3}, play.Plan{Member: 1, Spell: 2, Count: 3}, nil},
		{"camp rest", play.Camping{}, game.CampRest{Kind: 1}, play.CampRest{Kind: 1}, nil},
		{"picker", play.Assembling{Common: play.Common{Picker: &game.Picker{Cur: 2}}}, game.Pick{N: 2}, play.Pick{N: 2},
			func(s game.State) bool { return s.Picker != nil && s.Picker.Cur == 2 }},
		{"open picker", play.Assembling{}, game.OpenPicker{Kind: 1}, play.OpenPicker{Kind: 1}, nil},
		{"close picker", play.Assembling{}, game.ClosePicker{}, play.ClosePicker{}, nil},
		{"icon editor", play.Exploring{Common: play.Common{IconEditor: &game.IconEditor{Large: true}}}, game.IconColour{Part: 1, Value: 3}, play.IconColour{Part: 1, Value: 3},
			func(s game.State) bool { return s.IconEditor.Large && s.Flags.IconDrop }},
		{"edit icon", play.Camping{}, game.EditIcon{Member: 1}, play.EditIcon{Member: 1}, nil},
		{"icon size", play.Camping{}, game.IconSize{}, play.IconSize{}, nil},
		{"icon remove", play.Camping{}, game.IconRemove{}, play.IconRemove{}, nil},
		{"icon png", play.Camping{}, game.IconPNG{Path: "a.png"}, play.IconPNG{Path: "a.png"}, nil},
		{"end icon", play.Camping{}, game.EndIcon{Keep: true}, play.EndIcon{Keep: true}, nil},
		{"title", play.Scene{Kind: play.SceneTitle, Intro: intro}, game.TitleDone{}, play.Continue{},
			func(s game.State) bool { return s.Mode == game.ModeTitle && s.Intro == intro && s.Pending == nil }},
		{"title step", play.Scene{Kind: play.SceneTitle, Intro: intro}, game.TitlePicture{N: 2}, nil, nil},
		{"sequence", play.Scene{Kind: play.SceneSequence, Sequence: "finale"}, game.SequenceDone{Name: "finale"}, play.Continue{},
			func(s game.State) bool { return s.Sequence == "finale" && s.Pending == nil }},
		{"results", play.Scene{Kind: play.SceneResults, Results: res}, game.Answer{}, play.Continue{},
			func(s game.State) bool { return s.Results == res && s.Pending.Kind == game.Continue }},
		{"placed text", play.Scene{Placed: []play.Placed{{At: image.Pt(3, 2), Text: "Hull", Tone: game.ToneBad}},
			Layers: []game.Layer{{At: image.Rect(1, 1, 5, 5), FrameMs: 100}}}, game.Answer{}, play.Continue{},
			func(s game.State) bool {
				v := s.Page.View
				return len(v) == 2 && v[1][0].X == 3 && v[1][0].Fg == tones[game.ToneBad] && s.Page.Layers[0].At.Max.X == 5 && s.Pacing.Pace == game.PaceAnim
			}},
		{"screen", play.Scene{Kind: play.SceneScreen, Screen: &game.Screen{Kind: game.ScreenStarmap}}, game.Answer{}, play.Continue{},
			func(s game.State) bool { return s.Screen.Kind == game.ScreenStarmap }},
		{"heading", play.Scene{Kind: play.SceneHeading, Text: []string{"En route"}}, game.Answer{N: 6}, play.Heading{Dir: 6},
			func(s game.State) bool {
				return s.Pending.Kind == game.Direction && s.Pending.Prompt == "En route" && s.Pending.Options[0].Do == game.Leave{}
			}},
		{"heading done", play.Scene{Kind: play.SceneHeading}, game.Leave{}, play.Leave{}, nil},
		{"ship fight", play.ShipFight{Fight: sf, Menu: []string{"Pilot", "Gunner"}}, game.Answer{N: 1}, play.Crew{N: 1},
			func(s game.State) bool {
				return s.Screen.Kind == game.ScreenSpaceFight && s.Screen.SpaceFight.Range == 4 && s.Pending.Options[1].Label == "Gunner"
			}},
		{"ship sheet", play.ShipFight{Sheet: &game.ShipSheet{Name: "Ark"}}, game.Answer{N: 0}, play.Crew{N: 0},
			func(s game.State) bool { return s.Screen.ShipSheet.Name == "Ark" }},
		{"maneuver named", play.Fighting{Common: allow(play.DoManeuver{ID: 3}), Field: play.Battlefield{Maneuvers: []game.TurnManeuver{man}}}, game.Guard{}, play.Guard{},
			func(s game.State) bool { return s.Combat.Turn.Maneuvers[0].Name == "Sweep" }},
		{"turn figures", play.Fighting{Field: play.Battlefield{Turn: 1, Moves: 5, Casting: 9, Weapon: &game.Weapon{},
			Units: []play.Unit{{ID: 1, Pose: 1}}}}, game.EndTurn{}, play.EndTurn{},
			func(s game.State) bool {
				t := s.Combat.Turn
				return t.Moves == 5 && t.Casting == 9 && t.Weapon != nil && s.Combat.Units[0].Pose == 1
			}},
		{"br intimidate", play.Fighting{Common: allow(play.Intimidate{}, play.Jetpack{})}, game.Answer{N: 'I'}, play.Intimidate{},
			func(s game.State) bool {
				o := s.Combat.Turn.Options
				return slices.ContainsFunc(o, func(x game.Option) bool { return x.Label == "Jetpack" && x.Do == game.Answer{N: 'J'} })
			}},
		{"br target", play.Fighting{Common: allow(play.Target{})}, game.Answer{N: 'T'}, play.Target{}, nil},
		{"br view", play.Fighting{Common: allow(play.ViewTarget{})}, game.Answer{N: 'V'}, play.ViewTarget{}, nil},
		{"br aid", play.Fighting{Common: allow(play.Aid{})}, game.Answer{N: 'A'}, play.Aid{}, nil},
		{"br sprint", play.Fighting{Common: allow(play.Sprint{})}, game.Answer{N: 'S'}, play.Sprint{}, nil},
		{"br dodge", play.Fighting{Common: allow(play.Dodge{})}, game.Answer{N: 'D'}, play.Dodge{}, nil},
		{"br jetpack", play.Fighting{Common: allow(play.Jetpack{})}, game.Answer{N: 'J'}, play.Jetpack{}, nil},
		{"member text", play.Camping{Common: play.Common{Party: []play.Member{{ClassText: "a/b", Memo: "0/3",
			Powers: []game.Entry{{Name: "Heal", Verb: "heal"}}, Sheet: &game.SheetPage{Name: "Al"}}}}}, game.OpenSheet{Member: 0}, nil,
			func(s game.State) bool {
				m := s.Party[0]
				return m.ClassText == "a/b" && m.Memo == "0/3" && s.Sheet.Powers[0].Name == "Heal" && s.SheetPage.Name == "Al"
			}},
		{"music", play.Exploring{Common: play.Common{Where: game.Where{Music: "town"}}, Map: play.ExploreMap{Automap: game.Automap{NoMap: true}}},
			game.Look{}, play.Look{}, func(s game.State) bool { return s.Where.Music == "town" && s.Map.NoMap }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fake{st: c.st}
			r := New(f, game.Data{})
			if res := r.Apply(c.d); res.Err != nil {
				t.Fatal(res.Err)
			}
			if c.sent == nil && len(f.sent) != 0 || c.sent != nil && (len(f.sent) != 1 || !same(f.sent[0], c.sent)) {
				t.Fatalf("sent %#v, want %#v", f.sent, c.sent)
			}
			if c.ok != nil && !c.ok(r.State()) {
				t.Fatalf("state %+v", r.State())
			}
		})
	}
}

func same(a, b play.Action) bool { return name(a) == name(b) && reflectEqual(a, b) }

// TestEvents: the new facts as the old events.
func TestEvents(t *testing.T) {
	es := events([]play.Event{play.SaveOffered{}, play.ShipShown{Fight: game.SpaceFight{Range: 2}}, play.MusicChange{Song: "x"}})
	if es[0].Kind != game.EvSaves || es[1].Screen.SpaceFight.Range != 2 || es[1].Ms == 0 || es[2].Kind != game.EvSong {
		t.Fatalf("%+v", es)
	}
}

// TestRestPaced: camp's rest goes in chunks by Step and stops on StopRest.
func TestRestPaced(t *testing.T) {
	f := &fake{st: play.Camping{Common: allow(play.AdvanceTimeBy{})}}
	r := New(f, game.Data{})
	r.Apply(game.Rest{Minutes: 30})
	if p := r.State().Pacing; p.Pace != game.PaceRest || !p.Resting {
		t.Fatalf("pacing %+v", p)
	}
	r.Step()
	r.Apply(game.StopRest{})
	r.Step()
	if len(f.sent) != 1 || f.sent[0] != (play.AdvanceTimeBy{D: restChunk}) || r.State().Pacing.Pace != game.PaceNone {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestWalkTo: the auto-walk goes round the wall a square a Step.
func TestWalkTo(t *testing.T) {
	f := &fake{st: play.Exploring{Map: play.ExploreMap{Walk: walker{}}}}
	r := New(f, game.Data{})
	if res := r.Apply(game.WalkTo{At: image.Pt(1, 0)}); res.Err != nil {
		t.Fatal(res.Err)
	}
	if r.State().Pacing.Pace != game.PaceWalk {
		t.Fatal("not walking")
	}
	r.Step() // (0,0) -> (0,1): south, the wall east; the party faces north, so it turns first
	if _, ok := f.sent[0].(play.Face); !ok || len(f.sent) != 4 {
		t.Fatalf("sent %v", f.sent)
	}
	if a, ok := f.sent[2].(play.Arrive); !ok || a.To != image.Pt(0, 1) {
		t.Fatalf("sent %v", f.sent)
	}
}

func reflectEqual(a, b play.Action) bool { return reflect.DeepEqual(a, b) }

// TestCampCard: each member's card menu comes from what Allowed lets that
// member do, so no game builds a menu.
func TestCampCard(t *testing.T) {
	c := allow(play.Memorize{Member: 1, Spell: 2}, play.DropMember{Member: 0}, play.DropMember{Member: 1})
	c.Party = []play.Member{{Name: "A"}, {Name: "B", Items: []game.Item{{Name: "Sword"}}}}
	s := New(&fake{st: play.Camping{Common: c}}, game.Data{}).State()
	got := fmt.Sprint(s.Party[0].Card, s.Party[1].Card)
	if want := "[{View view} {Drop drop}] [{View view} {Items items} {Memorize memorize} {Drop drop}]"; got != want {
		t.Fatalf("cards %s, want %s", got, want)
	}
}

// TestCampKeepsView: camp is a menu over the 3D view last explored.
func TestCampKeepsView(t *testing.T) {
	f := &fake{st: play.Exploring{Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{}}}}}}}
	r := New(f, game.Data{})
	r.State()
	f.st = play.Camping{}
	if r.State().Walls == nil {
		t.Fatal("camp lost the 3D view")
	}
}

// TestDraftAsks: a draft asks only what it has to ask: a typed name with
// no Step still asks, a bare Step (the sheet) does not.
func TestDraftAsks(t *testing.T) {
	r := New(&fake{st: play.Assembling{Draft: &play.Draft{Typed: true}}}, game.Data{})
	if a := r.Pending(); a == nil || a.Kind != game.Name {
		t.Fatalf("typed draft asks %+v", a)
	}
	if a := New(&fake{st: play.Assembling{Draft: &play.Draft{Step: "sheet"}}}, game.Data{}).Pending(); a != nil {
		t.Fatalf("the sheet asks %+v", a)
	}
}

// TestPictureIsLayer: Common.Picture reaches the front as the page's layer
// (the front draws layers only).
func TestPictureIsLayer(t *testing.T) {
	pic := game.Picture{File: "PIC1", Block: 57}
	s := New(&fake{st: play.Exploring{Common: play.Common{Picture: &pic}}}, game.Data{}).State()
	if len(s.Page.Layers) != 1 || s.Page.Layers[0].Block != 57 {
		t.Fatalf("layers %+v", s.Page.Layers)
	}
}

// TestCardItems: a member's card has Items when Allowed lets that member
// handle an item, though their items are not in the state.
func TestCardItems(t *testing.T) {
	c := allow(play.Ready{Member: 1}, play.Drop{Member: 2})
	c.Party = []play.Member{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	s := New(&fake{st: play.Camping{Common: c}}, game.Data{}).State()
	if got := fmt.Sprint(s.Party[0].Card, s.Party[1].Card, s.Party[2].Card); got != "[{View view}] [{View view} {Items items}] [{View view} {Items items}]" {
		t.Fatal(got)
	}
}

// TestCampKeepsMap: camp keeps the map last explored.
func TestCampKeepsMap(t *testing.T) {
	f := &fake{st: play.Exploring{Map: play.ExploreMap{Automap: game.Automap{Size: image.Pt(2, 2)}}}}
	r := New(f, game.Data{})
	r.State()
	f.st = play.Camping{}
	if r.State().Map.Size != image.Pt(2, 2) {
		t.Fatal("camp lost the map")
	}
}

// arriveSays is a fake whose step itself says something (time passes).
type arriveSays struct{ fake }

func (a *arriveSays) Do(x play.Action) (play.Outcome, error) {
	a.sent = append(a.sent, x)
	if _, ok := x.(play.Arrive); ok {
		return play.Outcome{Events: []play.Event{play.Sounded{Sound: "step"}}}, nil
	}
	return play.Outcome{}, nil
}

// TestWalkRunsEvent: a walk's step runs the square's event though the
// step itself said something, as a manual step does.
func TestWalkRunsEvent(t *testing.T) {
	f := &arriveSays{fake{st: play.Exploring{Map: play.ExploreMap{Walk: walker{}}}}}
	r := New(f, game.Data{})
	r.Apply(game.WalkTo{At: image.Pt(0, 2)})
	r.Step()
	if len(f.sent) != 4 || f.sent[3] != (play.RunEvent{}) { // a turn, the script ahead, the step, the square's event
		t.Fatalf("sent %v", f.sent)
	}
}

// TestArriveAsAllowed: a step goes as the rules allow it, with their
// facing (a step back that turns the party), never refused for gobox's.
func TestArriveAsAllowed(t *testing.T) {
	step := play.Arrive{From: image.Pt(0, 6), To: image.Pt(0, 7), Facing: 4}
	f := &fake{st: play.Exploring{Common: allow(step)}}
	New(f, game.Data{}).Apply(game.Arrive{From: step.From, To: step.To})
	if len(f.sent) != 1 || f.sent[0] != step {
		t.Fatalf("sent %v", f.sent)
	}
}

// -1 means the selected member, so OpenSheet{-1} (the View key) opens a sheet.
func TestWhoSelected(t *testing.T) {
	ps := play.Exploring{Common: play.Common{Selected: 2}}
	if who(-1, ps) != 2 || who(1, ps) != 1 {
		t.Fatal("who(-1) must be the selected member")
	}
}

// TestItemsTabSheet: Tab on the items (CloseItems, not All) shows the
// selected member's sheet; Esc (All) closes both.
func TestItemsTabSheet(t *testing.T) {
	c := play.Common{Party: []play.Member{{Name: "A", Levels: []int{3, 2}}}} // no page from the rules: gobox makes its own
	r := New(&fake{st: play.Exploring{Common: c}}, game.Data{})
	r.Apply(game.OpenItems{})
	r.Apply(game.CloseItems{})
	if st := r.State(); st.SheetPage == nil || st.Sheet == nil || st.SheetPage.Levels != "3/2" {
		t.Fatal("Tab closed the items without the sheet, or its levels")
	}
}

// TestShopBrowse: Browse opens the shop's goods pane.
func TestShopBrowse(t *testing.T) {
	r := New(&fake{st: play.Trading{Goods: []game.Item{{Name: "Sword"}}}}, game.Data{})
	r.Apply(game.Browse{Member: -1})
	if r.State().Items != game.ItemsShop {
		t.Fatal("Browse opened nothing")
	}
}

// TestReadyToggles: Ready on a readied item unreadies it.
func TestReadyToggles(t *testing.T) {
	c := play.Common{Party: []play.Member{{Items: []game.Item{{Name: "Gun", Readied: true}}}}}
	f := &fake{st: play.Exploring{Common: c}}
	New(f, game.Data{}).Apply(game.Ready{Member: 0, Item: 0})
	if len(f.sent) != 1 || f.sent[0] != (play.Unready{}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// beforeSays is a fake whose script ahead of a step says something.
type beforeSays struct{ fake }

func (b *beforeSays) Do(x play.Action) (play.Outcome, error) {
	b.sent = append(b.sent, x)
	if e, ok := x.(play.RunEvent); ok && e.Before {
		return play.Outcome{Events: []play.Event{play.Narrated{Text: "You approach the gate."}}}, nil
	}
	return play.Outcome{}, nil
}

// TestWalkFinishesHeldStep: the script ahead says something; the walk
// still makes the step it holds (Arrive), never leaving the game waiting.
func TestWalkFinishesHeldStep(t *testing.T) {
	f := &beforeSays{fake{st: play.Exploring{Map: play.ExploreMap{Walk: walker{}}}}}
	r := New(f, game.Data{})
	r.Apply(game.WalkTo{At: image.Pt(0, 1)})
	r.Step()
	if !slices.ContainsFunc(f.sent, func(a play.Action) bool { _, ok := a.(play.Arrive); return ok }) {
		t.Fatalf("held step never made: sent %v", f.sent)
	}
}
