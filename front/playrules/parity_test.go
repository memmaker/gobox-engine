package playrules

import (
	"image"
	"slices"
	"testing"
	"time"

	"gobox/game"
	"gobox/game/play"
)

// slotted is fake with the game's own save slots (play.Slots).
type slotted struct {
	fake
	loaded int
}

func (s *slotted) Slots() []game.SaveSlot { return []game.SaveSlot{{Name: "A", Info: "old"}} }
func (s *slotted) LoadSlot(n int) error   { s.loaded = n + 1; return nil }

func labels(a *game.Ask) (l []string) {
	if a != nil {
		for _, o := range a.Options {
			l = append(l, o.Label)
		}
	}
	return l
}

// TestParity: what the coordinators' parity runs found missing, by fact.
func TestParity(t *testing.T) {
	pic := &game.Picture{File: "PIC", Block: 3}
	title := play.Scene{Kind: play.SceneTitle, Intro: &game.Intro{Pictures: []int{1, -1}},
		Steps: map[int]play.TitleStep{1: {Layers: []game.Layer{{Picture: game.Picture{File: "TITLE", Block: 1}}}, Song: "title"}}}
	cases := []struct {
		name string
		st   play.State
		data game.Data
		do   []game.Decision
		sent []play.Action
		ok   func(s game.State, res game.Result) bool
	}{
		{"title step layers and song", title, game.Data{}, []game.Decision{game.TitlePicture{N: 1}}, nil,
			func(s game.State, res game.Result) bool {
				return len(s.Page.Layers) == 1 && s.Page.Layers[0].Picture.File == "TITLE" && res.Events[0].Kind == game.EvSong
			}},
		{"title credits", title, game.Data{}, []game.Decision{game.TitleCredits{Page: 1}}, nil,
			func(s game.State, _ game.Result) bool {
				return s.Flags.Credits && s.Flags.CreditPage == 1 && len(s.Page.Layers) == 0
			}},
		{"title flash", title, game.Data{}, []game.Decision{game.TitleFlash{C: 9}}, nil, nil},
		{"holding door", play.Exploring{Door: &game.Door{}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Holding == game.HoldDoor }},
		{"holding fight", play.Fighting{}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Holding == game.HoldFight }},
		{"holding temple", play.Trading{Kind: play.TradeTemple}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Holding == game.HoldTemple }},
		{"window picture", play.Exploring{Common: play.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Window == game.WindowPicture && s.Picture.File == pic.File
			}},
		{"window map, no walls", play.Exploring{Map: play.ExploreMap{Overland: true}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Window == game.WindowMap && s.Walls == nil }},
		{"status line and pane", play.Fighting{Common: play.Common{Status: "Searching", Text: []string{"a", "b"}, Picture: pic}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Page.Prompt[0].Text == "Searching" && len(s.Page.Pane) == 2 && s.LastText == "b" && s.Picture.File == pic.File
			}},
		{"camp flags and automap", play.Camping{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Order{}}},
			Automap: game.Automap{Size: image.Pt(16, 16)}, Party: []play.Member{{CanCast: true}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Flags.Adventure && s.Flags.Ordering && s.Flags.CanCast && s.Map.Size.X == 16
			}},
		{"sweeping", play.Exploring{Common: play.Common{Where: game.Where{Searching: true}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Flags.Sweeping }},
		{"party menu", play.Assembling{Common: play.Common{Party: []play.Member{{}},
			Allowed: play.Allowed{Plain: []play.Action{play.NewCharacter{}, play.RemoveCharacter{}, play.QuickStart{}, play.Begin{}}}}},
			game.Data{Texts: map[string]string{"Remove Character": "Purge Character"}}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Mode == game.ModeParty && slices.Equal(labels(s.Pending),
					[]string{"Create Character", "Purge Character", "View Character", "Load Saved Game", "Save Current Game", "Quick Start", "Begin Adventuring"})
			}},
		{"creation step", play.Assembling{Draft: &play.Draft{Step: "career", Choices: []string{"Rogue", "Medic"}}}, game.Data{},
			[]game.Decision{game.Answer{N: 1}}, []play.Action{play.SetChoice{Field: "career", N: 1}},
			func(s game.State, _ game.Result) bool {
				return s.Pending.Kind == game.Choice && labels(s.Pending)[1] == "Medic"
			}},
		{"creation number", play.Assembling{Draft: &play.Draft{Step: "str", Number: true, Default: 9, Max: 18}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Pending.Kind == game.Number && s.Pending.Default == 9 && s.Pending.Max == 18
			}},
		{"question over exploring", play.Exploring{Common: play.Common{Ask: &play.Question{Text: "Bash it?", YesNo: true}},
			Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{}}}}}}, game.Data{},
			[]game.Decision{game.Answer{N: 1}}, []play.Action{play.Reply{N: 1}},
			func(s game.State, _ game.Result) bool {
				return s.Mode == game.ModeAsk && s.Walls != nil && s.Window == game.WindowScene && s.Holding == game.HoldNone && s.Flags.Adventure && s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Bash it?"
			}},
		{"question over fighting", play.Fighting{Common: play.Common{Ask: &play.Question{Replies: []string{"a"}}}}, game.Data{},
			[]game.Decision{game.Answer{N: 'A'}}, []play.Action{play.Reply{N: 'A'}},
			func(s game.State, _ game.Result) bool { return s.Combat != nil && s.Pending != nil }},
		{"series party menu", play.Assembling{Common: play.Common{Party: []play.Member{{}},
			Allowed: play.Allowed{Plain: []play.Action{play.RemoveCharacter{}, play.EditIcon{}, play.Begin{}}}}},
			game.Data{PartyMenu: []game.MenuKey{{Label: "Begin Adventuring", Key: 'B'}, {Label: "Remove Character", Key: 'P'}, {Label: "Icon Select", Key: 'I'}}}, nil, nil,
			func(s game.State, _ game.Result) bool {
				o := s.Pending.Options
				return slices.Equal(labels(s.Pending), []string{"Begin Adventuring", "Remove Character", "Icon Select"}) && o[1].Key == 'P'
			}},
		{"continue over fighting", play.Fighting{Common: play.Common{Status: "Press button or enter to continue.", Ask: &play.Question{Continue: true}}},
			game.Data{}, []game.Decision{game.Answer{}}, []play.Action{play.Continue{}},
			func(s game.State, _ game.Result) bool {
				return s.Combat != nil && s.Pending.Kind == game.Continue && s.Page.Prompt[0].Text != ""
			}},
		{"heading over exploring", play.Exploring{Common: play.Common{Ask: &play.Question{Heading: true}}},
			game.Data{}, []game.Decision{game.Answer{N: 2}}, []play.Action{play.Heading{Dir: 2}},
			func(s game.State, _ game.Result) bool {
				return s.Mode == game.ModeAsk && s.Pending.Kind == game.Direction && s.Pending.Options[0].Do == game.Leave{}
			}},
		{"draft sheet figures", play.Assembling{Draft: &play.Draft{Member: play.Member{Sheet: &game.SheetPage{Age: 23}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.SheetPage != nil && s.SheetPage.Age == 23 }},
		{"no save at party menu", play.Assembling{Common: play.Common{NoSave: true, NoLoad: true, Party: []play.Member{{Name: "A"}}}}, game.Data{},
			nil, nil,
			func(s game.State, res game.Result) bool {
				for _, o := range s.Pending.Options {
					if o.Label == "Save Current Game" || o.Label == "Load Saved Game" {
						return false
					}
				}
				return true
			}},
		{"free aim is gobox's", play.Fighting{}, game.Data{}, []game.Decision{game.Aim{At: image.Pt(3, 4)}}, nil,
			func(s game.State, res game.Result) bool { return res.Err == nil }},
		{"no clashing keys", play.Assembling{Draft: &play.Draft{Step: "race", Choices: []string{"Martian", "Mercurian"}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'M' && s.Pending.Options[1].Key == 0
			}},
		{"bars keep keys under NoHotkeys", play.ShipFight{Menu: []string{"Quit", "View"}}, game.Data{NoHotkeys: true}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'Q' && s.Pending.Options[1].Key == 'V'
			}},
		{"team menu keys under NoHotkeys", play.Assembling{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.NewCharacter{}}}}}, game.Data{NoHotkeys: true}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Pending.Options[0].Key == 'C' }},
		{"list Done keeps its key", play.Scene{Replies: []string{"Agility", "Next", "Done"}}, game.Data{NoHotkeys: true}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 'N' && s.Pending.Options[2].Key == 'D'
			}},
		{"roster pick prompt", play.Scene{Replies: []string{"Wilma", "Buck"}, Common: play.Common{Status: "Add a character: "}}, game.Data{NoHotkeys: true}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Pending.Prompt == "Add a character: " && s.Pending.Options[0].Key == 0
			}},
		{"turn keys unique", play.Fighting{Common: play.Common{Allowed: play.Allowed{Moves: []game.Reach{{}}, Plain: []play.Action{play.DoManeuver{}}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				o := s.Combat.Turn.Options
				return len(o) == 3 && o[0].Key == 'M' && o[1].Key == 0
			}},
		{"keeper portrait", play.Trading{Common: play.Common{Portrait: &play.Portrait{Head: game.Picture{File: "HEAD2", Block: 3}, Body: game.Picture{File: "BODY2", Block: 3}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Window == game.WindowPicture && len(s.Story) == 2 && s.Story[0].File == "HEAD2"
			}},
		{"adventure on town loot", play.Looting{Common: play.Common{Where: game.Where{Mode: "town"}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Flags.Adventure }},
		{"fight log from text", play.Fighting{Common: play.Common{Text: []string{"YOU GET INTO A BRAWL."}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return slices.Equal(s.Combat.Log, []string{"YOU GET INTO A BRAWL."}) }},
		{"aimed attack", play.Fighting{Common: play.Common{Allowed: play.Allowed{Aimed: []play.Aimed{{Action: play.Attack{}, Squares: []image.Point{{2, 3}}}}}}}, game.Data{},
			[]game.Decision{game.Attack{}, game.Aim{At: image.Pt(2, 3)}}, []play.Action{play.Attack{At: &image.Point{2, 3}}},
			func(s game.State, _ game.Result) bool {
				var att, aim bool
				for _, o := range s.Combat.Turn.Options {
					_, a := o.Do.(game.Attack)
					_, b := o.Do.(game.Aim)
					att, aim = att || a, aim || b
				}
				return att && aim
			}},
		{"camp layers and day", play.Camping{Common: play.Common{Day: true, Layers: []game.Layer{{FrameMs: 100}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				return s.Window == game.WindowPicture && len(s.Page.Layers) == 1 && s.Flags.Day
			}},
		{"camp picture", play.Camping{Common: play.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Window == game.WindowPicture }},
		{"reroll yes/no", play.Assembling{Draft: &play.Draft{Step: "roll", YesNo: true, Prompt: "Reroll stats?"}}, game.Data{},
			[]game.Decision{game.Answer{N: 1}}, []play.Action{play.SetChoice{Field: "roll", N: 1}},
			func(s game.State, _ game.Result) bool {
				return s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Reroll stats?"
			}},
		{"no hotkeys", play.Assembling{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Begin{}}}},
			Draft: &play.Draft{Step: "race", Choices: []string{"Terran"}}}, game.Data{NoHotkeys: true}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Pending.Options[0].Key == 0 }},
		{"continue only", play.Scene{Kind: play.SceneScreen, Screen: &game.Screen{Kind: game.ScreenShipSheet},
			Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"ship fight continue", play.ShipFight{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"move verb", play.Fighting{Common: play.Common{Allowed: play.Allowed{Moves: []game.Reach{{}}, Plain: []play.Action{play.Delay{}, play.EndTurn{}}}}},
			game.Data{}, nil, nil,
			func(s game.State, _ game.Result) bool {
				var l []string
				for _, o := range s.Combat.Turn.Options {
					l = append(l, o.Label)
				}
				return slices.Equal(l, []string{"Move", "Wait", "End", "Aim"})
			}},
		{"verb keys from data", play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Aid{}}}}},
			game.Data{VerbKeys: map[string]rune{"Aid": 'H'}}, []game.Decision{game.Answer{N: 'H'}}, []play.Action{play.Aid{}},
			func(s game.State, _ game.Result) bool { return s.Combat.Turn.Options[0].Key == 'H' }},
		{"rest capped", play.Camping{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{D: 5 * time.Minute}}}}},
			game.Data{}, []game.Decision{game.Rest{Minutes: 60}}, nil,
			func(s game.State, _ game.Result) bool { return s.Pacing.Pace == game.PaceRest }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fake{st: c.st}
			r := New(f, c.data)
			var res game.Result
			for _, d := range c.do {
				if res = r.Apply(d); res.Err != nil {
					t.Fatal(res.Err)
				}
			}
			if len(f.sent) != len(c.sent) || c.sent != nil && !same(f.sent[0], c.sent[0]) {
				t.Fatalf("sent %#v, want %#v", f.sent, c.sent)
			}
			if c.ok != nil && !c.ok(r.State(), res) {
				t.Fatalf("state %+v / %+v", r.State(), res)
			}
		})
	}
}

// TestRestChunkCap: a chunk is no longer than the allowed rest.
func TestRestChunkCap(t *testing.T) {
	f := &fake{st: play.Camping{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{D: 5 * time.Minute}}}}}}
	r := New(f)
	r.Apply(game.Rest{Minutes: 60})
	r.Step()
	if f.sent[0] != (play.AdvanceTimeBy{D: 5 * time.Minute}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestCombatLogNote: the fight's text builds the log; a refusal is the toast.
func TestCombatLogNote(t *testing.T) {
	f := &fake{st: play.Fighting{}, out: play.Outcome{Events: []play.Event{play.Narrated{Text: "Hits!"}, play.Refused{Why: "no"}}}}
	r := New(f)
	r.Apply(game.EndTurn{})
	if s := r.State(); !slices.Equal(s.Combat.Log, []string{"Hits!"}) || s.Note.Text != "no" {
		t.Fatalf("log %v note %v", s.Combat.Log, s.Note)
	}
}

// TestGameSlots: the game's own slots list after gobox's and load by index.
func TestGameSlots(t *testing.T) {
	g := &slotted{fake: fake{st: play.Assembling{}}}
	r := New(g).(*rules)
	r.dir = t.TempDir()
	s := r.State()
	if len(s.Saves) != 1 || s.Saves[0].Info != "old" {
		t.Fatalf("saves %v", s.Saves)
	}
	if res := r.Apply(game.Load{Slot: 0}); res.Err != nil || g.loaded != 1 {
		t.Fatalf("load %+v %d", res, g.loaded)
	}
}

func TestStudyRestAndSaveDir(t *testing.T) {
	dir := t.TempDir()
	g := &fake{st: play.Camping{Study: 3 * time.Hour, Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{}}}}}}
	r := New(g, game.Data{SaveDir: dir}).(*rules)
	r.Apply(game.Rest{})
	if r.rest != 3*time.Hour || r.dir != dir {
		t.Fatalf("rest %v dir %q", r.rest, r.dir)
	}
	g.st = play.Camping{Common: play.Common{NoSave: true, NoLoad: true}}
	if f := r.State().Flags; f == nil || !f.NoSave || !f.NoLoad {
		t.Fatal("NoSave/NoLoad not published")
	}
	if r.Apply(game.Save{Name: "x"}).Err == nil {
		t.Fatal("saved where the rules refuse it")
	}
}
