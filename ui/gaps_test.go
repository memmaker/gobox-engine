package ui

import (
	"fmt"
	"image"
	"image/color"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
)

func allow(a ...game.Action) game.Common { return game.Common{Allowed: game.Allowed{Plain: a}} }

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
		st   game.State
		d    any
		sent game.Action // nil: nothing reaches Do
		ok   func(s shown) bool
	}{
		{"door", game.Exploring{Door: door, Day: true, Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{}}}, Outside: true, Sun: true, Mountains: true}}},
			game.TryDoor{Way: 1}, game.TryDoor{Way: game.DoorBash},
			func(s shown) bool { return s.Door == door && s.Walls.Sun && s.Walls.Mountains && s.Flags.Day }},
		{"night sky", game.Exploring{Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{}}}, Outside: true}}}, game.Search{}, game.Search{},
			func(s shown) bool { return !s.Walls.Sun && !s.Flags.Day }},
		{"vault", game.Trading{Kind: game.TradeVault, Vault: &game.Vault{Money: map[string]int{"gold": 5}}}, game.Leave{}, game.Leave{},
			func(s shown) bool { return s.Mode == game.ModeVault && s.Vault.Money["gold"] == 5 }},
		{"shop identify", game.Trading{Identify: 20, Unknown: []game.ItemRef{{Member: 0, Item: 1}}}, game.Identify{Item: 1}, game.Identify{Item: 1},
			func(s shown) bool { return s.Shop.IdentifyPrice == 20 && len(s.Shop.Unknown) == 1 }},
		{"training learn", game.Trading{Kind: game.TradeTraining, Training: &game.Training{Learning: true}}, game.Learn{Spell: 4}, game.Learn{Spell: 4},
			func(s shown) bool { return s.Mode == game.ModeTraining && s.Training.Learning }},
		{"handout esc leaves", game.Scene{Kind: game.SceneHandout, Handout: "Journal", Layers: []game.Layer{{}}, Common: allow(game.Leave{})}, game.Cancel{}, game.Leave{},
			func(s shown) bool { return s.Read != nil && len(s.Page.Layers) == 1 && s.Holding == game.HoldNone }},
		{"handout read out", game.Scene{Kind: game.SceneHandout, Common: allow(game.Leave{})}, tAnswer{N: 2}, game.Leave{}, nil},
		{"story page", game.Scene{Text: []string{"hi"}, Common: allow(game.ContinueAction{})}, tAnswer{}, game.ContinueAction{},
			func(s shown) bool {
				r := s.Page.Pane[0][0]
				return s.Holding == game.HoldNone && r.Upcase && r.Fg == (color.RGBA{})
			}},
		{"forget entry", game.Assembling{Common: allow(game.OpenPicker{Kind: 5})}, game.OpenPicker{Kind: 5}, game.OpenPicker{Kind: 5},
			func(s shown) bool {
				return slices.ContainsFunc(s.Pending.Options, func(o game.Option) bool { return o.Label == "Forget Character" })
			}},
		{"rest heals", game.Camping{Heals: true, Common: allow(game.AdvanceTimeBy{})}, game.BreakCamp{}, game.BreakCamp{},
			func(s shown) bool { return s.RestHeals && !s.RestStudies }},
		{"speaker over view", game.Exploring{Common: game.Common{Portrait: &game.Portrait{}}}, game.Search{}, game.Search{},
			func(s shown) bool { return s.Window == game.WindowPicture }},
		{"continue over the view", game.Exploring{Common: game.Common{Ask: &game.Question{Continue: true}, Allowed: game.Allowed{Plain: []game.Action{game.ContinueAction{}}}}}, tAnswer{}, game.ContinueAction{},
			func(s shown) bool { return s.Mode == game.ModeExplore && s.Pending.Kind == game.Continue }},
		{"open training", game.Exploring{}, game.OpenTraining{Member: 1}, game.OpenTraining{Member: 1}, nil},
		{"loot detect", game.Looting{Status: true, Common: allow(game.LootDetect{}, game.LootShare{})}, game.LootDetect{}, game.LootDetect{},
			func(s shown) bool { return s.Loot.Status && s.Allowed["LootDetect"] && s.Allowed["LootShare"] }},
		{"loot esc leaves", game.Looting{Common: allow(game.Leave{})}, game.Cancel{}, game.Leave{}, nil},
		{"loot share", game.Looting{}, game.LootShare{}, game.LootShare{}, nil},
		{"loot status", game.Looting{}, game.LootStatus{}, game.LootStatus{}, nil},
		{"power pick", game.Camping{Common: game.Common{PowerPick: &game.PowerPick{Name: "Bob"}}}, game.PickPower{ID: 2}, game.PickPower{ID: 2},
			func(s shown) bool { return s.PowerPick != nil && s.PowerPick.Name == "Bob" }},
		{"use power", game.Camping{}, game.UsePower{Member: 1, Kind: game.PowerDisplay}, game.UsePower{Member: 1, Kind: game.PowerDisplay}, nil},
		{"policy", game.Camping{}, game.SetPolicy{ID: 1, Policy: 2}, game.SetPolicy{ID: 1, Policy: 2}, nil},
		{"mark", game.Exploring{}, game.Mark{At: image.Pt(1, 2), Text: "x"}, game.Mark{At: image.Pt(1, 2), Text: "x"}, nil},
		{"unmark", game.Exploring{}, game.Unmark{At: image.Pt(1, 2)}, game.Unmark{At: image.Pt(1, 2)}, nil},
		{"rename", game.Assembling{}, game.Rename{Name: "Al"}, game.Rename{Name: "Al"}, nil},
		{"quit", game.Over{Text: "The end"}, game.Quit{}, game.Quit{},
			func(s shown) bool { return s.LastText == "The end" }},
		{"modify", game.Assembling{}, game.Modify{Member: 2}, game.Modify{Member: 2}, nil},
		{"quick start", game.Assembling{}, game.QuickStart{}, game.QuickStart{}, nil},
		{"roll", game.Assembling{Draft: &game.Draft{Step: "skills", Ask: &game.Question{Replies: []string{"a"}}}}, tRoll{Minimal: true}, game.Reroll{Minimal: true},
			func(s shown) bool { return s.Creation.Step == "skills" && len(s.Creation.Choices) == 1 }},
		{"plan", game.Camping{}, game.Plan{Member: 1, Spell: 2, Count: 3}, game.Plan{Member: 1, Spell: 2, Count: 3}, nil},
		{"camp rest", game.Camping{}, game.CampRest{Kind: 1}, game.CampRest{Kind: 1}, nil},
		{"picker", game.Assembling{Common: game.Common{Picker: &game.Picker{Cur: 2}}}, game.Pick{N: 2}, game.Pick{N: 2},
			func(s shown) bool { return s.Picker != nil && s.Picker.Cur == 2 }},
		{"open picker", game.Assembling{}, game.OpenPicker{Kind: 1}, game.OpenPicker{Kind: 1}, nil},
		{"close picker", game.Assembling{}, game.ClosePicker{}, game.ClosePicker{}, nil},
		{"icon editor", game.Exploring{Common: game.Common{IconEditor: &game.IconEditor{Large: true}}}, game.IconColour{Part: 1, Value: 3}, game.IconColour{Part: 1, Value: 3},
			func(s shown) bool { return s.IconEditor.Large && s.Flags.IconDrop }},
		{"edit icon", game.Camping{}, game.EditIcon{Member: 1}, game.EditIcon{Member: 1}, nil},
		{"icon size", game.Camping{}, game.IconSize{}, game.IconSize{}, nil},
		{"icon remove", game.Camping{}, game.IconRemove{}, game.IconRemove{}, nil},
		{"icon png", game.Camping{}, game.IconPNG{Path: "a.png"}, game.IconPNG{Path: "a.png"}, nil},
		{"end icon", game.Camping{}, game.EndIcon{Keep: true}, game.EndIcon{Keep: true}, nil},
		{"title", game.Scene{Kind: game.SceneTitle, Intro: intro}, tTitleDone{}, game.ContinueAction{},
			func(s shown) bool { return s.Mode == game.ModeTitle && s.Intro == intro && s.Pending == nil }},
		{"title step", game.Scene{Kind: game.SceneTitle, Intro: intro}, tTitlePicture{N: 2}, nil, nil},
		{"sequence", game.Scene{Kind: game.SceneSequence, Sequence: "finale"}, tSequenceDone{Name: "finale"}, game.ContinueAction{},
			func(s shown) bool { return s.Sequence == "finale" && s.Pending == nil }},
		{"results", game.Scene{Kind: game.SceneResults, Results: res}, tAnswer{}, game.ContinueAction{},
			func(s shown) bool { return s.Results == res && s.Pending.Kind == game.Continue }},
		{"placed text", game.Scene{Placed: []game.Placed{{At: image.Pt(3, 2), Text: "Hull", Tone: game.ToneBad}},
			Layers: []game.Layer{{At: image.Rect(1, 1, 5, 5), FrameMs: 100}}}, tAnswer{}, game.ContinueAction{},
			func(s shown) bool {
				v := s.Page.View
				return len(v) == 2 && v[1][0].X == 3 && v[1][0].Fg == tones[game.ToneBad] && s.Page.Layers[0].At.Max.X == 5 && s.Pacing.Pace == game.PaceAnim
			}},
		{"screen", game.Scene{Kind: game.SceneScreen, Screen: &game.Screen{Kind: game.ScreenStarmap}}, tAnswer{}, game.ContinueAction{},
			func(s shown) bool { return s.Screen.Kind == game.ScreenStarmap }},
		{"heading", game.Scene{Kind: game.SceneHeading, Text: []string{"En route"}}, tAnswer{N: 6}, game.Heading{Dir: 6},
			func(s shown) bool {
				return s.Pending.Kind == game.Direction && s.Pending.Prompt == "En route" && s.Pending.Options[0].Do == game.Leave{}
			}},
		{"heading done", game.Scene{Kind: game.SceneHeading}, game.Leave{}, game.Leave{}, nil},
		{"ship fight", game.ShipFight{Fight: sf, Menu: &game.Question{Replies: []string{"Pilot", "Gunner"}}}, tAnswer{N: 1}, game.Crew{N: 1},
			func(s shown) bool {
				return s.Screen.Kind == game.ScreenSpaceFight && s.Screen.SpaceFight.Range == 4 && s.Pending.Options[1].Label == "Gunner"
			}},
		{"ship sheet", game.ShipFight{Sheet: &game.ShipSheet{Name: "Ark"}}, tAnswer{N: 0}, game.Crew{N: 0},
			func(s shown) bool { return s.Screen.ShipSheet.Name == "Ark" }},
		{"maneuver named", game.Fighting{Common: allow(game.DoManeuver{ID: 3}), Field: game.Battlefield{Maneuvers: []game.TurnManeuver{man}}}, game.Guard{}, game.Guard{},
			func(s shown) bool { return s.Combat.Turn.Maneuvers[0].Name == "Sweep" }},
		{"turn figures", game.Fighting{Field: game.Battlefield{Turn: 1, Moves: 5, Casting: 9, Weapon: &game.Weapon{},
			Units: []game.Unit{{ID: 1, Pose: 1}}}}, game.EndTurn{}, game.EndTurn{},
			func(s shown) bool {
				t := s.Combat.Turn
				return t.Moves == 5 && t.Casting == 9 && t.Weapon != nil && s.Combat.Units[0].Pose == 1
			}},
		{"br intimidate", game.Fighting{Common: allow(game.Intimidate{}, game.Jetpack{})}, tAnswer{N: 'I'}, game.Intimidate{},
			func(s shown) bool {
				o := s.Combat.Turn.Options
				return slices.ContainsFunc(o, func(x game.Option) bool { return x.Label == "Jetpack" && x.Do == game.Action(game.Jetpack{}) })
			}},
		{"br target", game.Fighting{Common: allow(game.Target{})}, tAnswer{N: 'T'}, game.Target{}, nil},
		{"br view", game.Fighting{Common: allow(game.ViewTarget{})}, tAnswer{N: 'V'}, game.ViewTarget{}, nil},
		{"br aid", game.Fighting{Common: allow(game.Aid{})}, tAnswer{N: 'A'}, game.Aid{}, nil},
		{"br sprint", game.Fighting{Common: allow(game.Sprint{})}, tAnswer{N: 'S'}, game.Sprint{}, nil},
		{"br dodge", game.Fighting{Common: allow(game.Dodge{})}, tAnswer{N: 'D'}, game.Dodge{}, nil},
		{"br jetpack", game.Fighting{Common: allow(game.Jetpack{})}, tAnswer{N: 'J'}, game.Jetpack{}, nil},
		{"member text", game.Camping{Common: game.Common{Party: []game.Member{{ClassText: "a/b", Memo: "0/3",
			SheetPowers: []game.Entry{{Name: "Heal", Verb: "heal"}}, Page: &game.SheetPage{Name: "Al"}}}}}, tOpenSheet{Member: 0}, nil,
			func(s shown) bool {
				m := s.Party[0]
				return m.ClassText == "a/b" && m.Memo == "0/3" && s.Sheet.Powers[0].Name == "Heal" && s.SheetPage.Name == "Al"
			}},
		{"music", game.Exploring{Common: game.Common{Where: game.Where{Music: "town"}}, Map: game.ExploreMap{Automap: game.Automap{NoMap: true}}},
			game.Look{}, game.Look{}, func(s shown) bool { return s.Where.Music == "town" && s.Map.NoMap }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fake{st: c.st}
			r := newRules(f, game.Data{})
			if res := r.send(c.d); res.Err != nil {
				t.Fatal(res.Err)
			}
			if c.sent == nil && len(f.sent) != 0 || c.sent != nil && (len(f.sent) != 1 || !same(f.sent[0], c.sent)) {
				t.Fatalf("sent %#v, want %#v", f.sent, c.sent)
			}
			if c.ok != nil && !c.ok(shownOf(r)) {
				t.Fatalf("state %+v", shownOf(r))
			}
		})
	}
}

func same(a, b game.Action) bool { return name(a) == name(b) && reflectEqual(a, b) }

// TestEvents: the new facts as the old events.
func TestEvents(t *testing.T) {
	es := events([]game.Event{game.SaveOffered{}, game.ShipShown{Fight: game.SpaceFight{Range: 2}}, game.MusicChange{Song: "x"}})
	if es[0].Kind != game.EvSaves || es[1].Screen.SpaceFight.Range != 2 || es[1].Ms == 0 || es[2].Kind != game.EvSong {
		t.Fatalf("%+v", es)
	}
}

// TestRestPaced: camp's rest goes in chunks by Step and stops on StopRest.
func TestRestPaced(t *testing.T) {
	f := &fake{st: game.Camping{Common: allow(game.AdvanceTimeBy{})}}
	r := newRules(f, game.Data{})
	r.send(tRest{Minutes: 30})
	if p := shownOf(r).Pacing; p.Pace != game.PaceRest || !p.Resting {
		t.Fatalf("pacing %+v", p)
	}
	r.Step()
	r.send(game.StopRest{})
	r.Step()
	if len(f.sent) != 1 || f.sent[0] != (game.AdvanceTimeBy{D: restChunk}) || shownOf(r).Pacing.Pace != game.PaceNone {
		t.Fatalf("sent %v", f.sent)
	}
}

func reflectEqual(a, b game.Action) bool { return reflect.DeepEqual(a, b) }

// TestCampCard: each member's card menu comes from what Allowed lets that
// member do, so no game builds a menu.
func TestCampCard(t *testing.T) {
	c := allow(game.Memorize{Member: 1, Spell: 2}, game.DropMember{Member: 0}, game.DropMember{Member: 1})
	c.Party = []game.Member{{Name: "A"}, {Name: "B", Items: []game.Item{{Name: "Sword"}}}}
	cards := cardsOf(game.Camping{Common: c})
	got := fmt.Sprint(cards[0].Menu, cards[1].Menu)
	if want := "[{View view} {Drop drop}] [{View view} {Items items} {Memorize memorize} {Drop drop}]"; got != want {
		t.Fatalf("cards %s, want %s", got, want)
	}
}

// TestCampKeepsView: camp is a menu over the 3D view last explored.
func TestCampKeepsView(t *testing.T) {
	f := &fake{st: game.Exploring{Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{}}}}}}}
	r := newRules(f, game.Data{})
	shownOf(r)
	f.st = game.Camping{}
	if shownOf(r).Walls == nil {
		t.Fatal("camp lost the 3D view")
	}
}

// TestDraftAsks: a draft asks only what it has to ask: a typed name with
// no Step still asks, a bare Step (the sheet) does not.
func TestDraftAsks(t *testing.T) {
	r := newRules(&fake{st: game.Assembling{Draft: &game.Draft{Ask: &game.Question{Typed: true}}}}, game.Data{})
	if a := r.pending(r.g.State()); a == nil || a.Kind != game.Name {
		t.Fatalf("typed draft asks %+v", a)
	}
	if a := func() *game.Ask {
		r := newRules(&fake{st: game.Assembling{Draft: &game.Draft{Step: "sheet"}}}, game.Data{})
		return r.pending(r.g.State())
	}(); a != nil {
		t.Fatalf("the sheet asks %+v", a)
	}
}

// TestPictureIsLayer: Common.Picture reaches the front as the page's layer
// (the front draws layers only).
func TestPictureIsLayer(t *testing.T) {
	pic := game.Picture{File: "PIC1", Block: 57}
	s := shownOf(newRules(&fake{st: game.Exploring{Common: game.Common{Picture: &pic}}}, game.Data{}))
	if len(s.Page.Layers) != 1 || s.Page.Layers[0].Block != 57 {
		t.Fatalf("layers %+v", s.Page.Layers)
	}
}

// TestCardItems: a member's card has Items when Allowed lets that member
// handle an item, though their items are not in the state.
func TestCardItems(t *testing.T) {
	c := allow(game.Ready{Member: 1}, game.Drop{Member: 2})
	c.Party = []game.Member{{Name: "A"}, {Name: "B"}, {Name: "C"}}
	cards := cardsOf(game.Camping{Common: c})
	if got := fmt.Sprint(cards[0].Menu, cards[1].Menu, cards[2].Menu); got != "[{View view}] [{View view} {Items items}] [{View view} {Items items}]" {
		t.Fatal(got)
	}
}

// TestCampKeepsMap: camp keeps the map last explored.
func TestCampKeepsMap(t *testing.T) {
	f := &fake{st: game.Exploring{Map: game.ExploreMap{Automap: game.Automap{Size: image.Pt(2, 2), Known: []image.Point{{0, 0}}}}}}
	r := newRules(f, game.Data{})
	shownOf(r)
	f.st = game.Camping{}
	if shownOf(r).Map.Size != image.Pt(2, 2) {
		t.Fatal("camp lost the map")
	}
}

// TestCampBareMap: a camp map with nothing known (the party's square
// only) is the map last explored, not a black minimap.
func TestCampBareMap(t *testing.T) {
	known := []image.Point{{0, 0}, {1, 0}}
	f := &fake{st: game.Exploring{Map: game.ExploreMap{Automap: game.Automap{Size: image.Pt(16, 16), Known: known}}}}
	r := newRules(f, game.Data{})
	shownOf(r)
	f.st = game.Camping{Common: game.Common{Automap: game.Automap{Size: image.Pt(16, 16)}}}
	if got := shownOf(r).Map.Known; len(got) != 2 {
		t.Fatalf("camp map knows %v", got)
	}
}

// TestArriveAsAllowed: a step goes as the rules allow it, with their
// facing (a step back that turns the party), never refused for gobox's.
func TestArriveAsAllowed(t *testing.T) {
	step := game.Arrive{From: image.Pt(0, 6), To: image.Pt(0, 7), Facing: 4}
	f := &fake{st: game.Exploring{Common: allow(step)}}
	newRules(f, game.Data{}).send(game.Arrive{From: step.From, To: step.To})
	if len(f.sent) != 1 || f.sent[0] != step {
		t.Fatalf("sent %v", f.sent)
	}
}

// -1 means the selected member, so OpenSheet{-1} (the View key) opens a sheet.
func TestWhoSelected(t *testing.T) {
	ps := game.Exploring{Common: game.Common{Selected: 2}}
	if who(-1, ps) != 2 || who(1, ps) != 1 {
		t.Fatal("who(-1) must be the selected member")
	}
}

// TestItemsTabSheet: Tab on the items (CloseItems, not All) shows the
// selected member's sheet; Esc (All) closes both.
func TestItemsTabSheet(t *testing.T) {
	c := game.Common{Party: []game.Member{{Name: "A", Levels: []int{3, 2}}}} // no page from the rules: gobox makes its own
	r := newRules(&fake{st: game.Exploring{Common: c}}, game.Data{})
	r.send(game.OpenItems{})
	r.send(tCloseItems{})
	if st := shownOf(r); st.SheetPage == nil || st.Sheet == nil || st.SheetPage.Levels != "3/2" {
		t.Fatal("Tab closed the items without the sheet, or its levels")
	}
}

// TestShopBrowse: Browse opens the shop's goods pane.
func TestShopBrowse(t *testing.T) {
	r := newRules(&fake{st: game.Trading{Goods: []game.Item{{Name: "Sword"}}}}, game.Data{})
	r.send(tBrowse{Member: -1})
	if shownOf(r).Items != game.ItemsShop {
		t.Fatal("Browse opened nothing")
	}
}

// TestReadyToggles: Ready on a readied item unreadies it.
func TestReadyToggles(t *testing.T) {
	c := game.Common{Party: []game.Member{{Items: []game.Item{{Handle: 1, Name: "Gun", Readied: true}}}}}
	f := &fake{st: game.Exploring{Common: c}}
	newRules(f, game.Data{}).send(game.Ready{Member: 0, Item: 1})
	if len(f.sent) != 1 || f.sent[0] != (game.Unready{Item: 1}) {
		t.Fatalf("sent %v", f.sent)
	}
}
