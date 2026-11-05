package playrules

import (
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
	r.Step() // (0,0) -> (0,1): south, the wall east
	if a, ok := f.sent[1].(play.Arrive); !ok || a.To != image.Pt(0, 1) || len(f.sent) != 3 {
		t.Fatalf("sent %v", f.sent)
	}
}

func reflectEqual(a, b play.Action) bool { return reflect.DeepEqual(a, b) }
