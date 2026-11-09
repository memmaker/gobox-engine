package classic

import (
	"image"
	"slices"
	"testing"
	"time"

	"gobox/game"
	"gobox/game/play"
)

// slottedFake is fake with the game's own save slots (play.Slots).
type slottedFake struct {
	fake
	loaded int
}

func (s *slottedFake) Slots() []game.SaveSlot { return []game.SaveSlot{{Name: "A", Info: "old"}} }
func (s *slottedFake) LoadSlot(n int) error   { s.loaded = n + 1; return nil }

func optionLabels(a *play.Ask) (l []string) {
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
		do   []any
		sent []play.Action
		ok   func(s shown, res play.Result) bool
	}{
		{"title step layers and song", title, game.Data{}, []any{tTitlePicture{N: 1}}, nil,
			func(s shown, res play.Result) bool {
				return len(s.Page.Layers) == 1 && s.Page.Layers[0].Picture.File == "TITLE" && res.Events[0].Kind == game.EvSong
			}},
		{"title credits", title, game.Data{}, []any{tTitleCredits{Page: 1}}, nil,
			func(s shown, _ play.Result) bool {
				return s.Flags.Credits && s.Flags.CreditPage == 1 && len(s.Page.Layers) == 0
			}},
		{"title flash", title, game.Data{}, []any{tTitleFlash{C: 9}}, nil, nil},
		{"holding door", play.Exploring{Door: &game.Door{}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Holding == game.HoldDoor }},
		{"holding fight", play.Fighting{}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Holding == game.HoldFight }},
		{"holding temple", play.Trading{Kind: play.TradeTemple}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Holding == game.HoldTemple }},
		{"window picture", play.Exploring{Common: play.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Window == game.WindowPicture && s.Picture.File == pic.File
			}},
		{"window map, no walls", play.Exploring{Map: play.ExploreMap{Overland: true}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Window == game.WindowMap && s.Walls == nil }},
		{"status line and pane", play.Fighting{Common: play.Common{Status: "Searching", Text: []string{"a", "b"}, Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Page.Prompt[0].Text == "Searching" && len(s.Page.Pane) == 2 && s.LastText == "b" && s.Picture.File == pic.File
			}},
		{"camp flags and automap", play.Camping{Common: play.Common{Adventure: true, Allowed: play.Allowed{Plain: []play.Action{play.Order{}}},
			Automap: game.Automap{Size: image.Pt(16, 16)}, Party: []game.Member{{CanCast: true}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Flags.Adventure && s.Flags.Ordering && s.Flags.CanCast && s.Map.Size.X == 16
			}},
		{"sweeping", play.Exploring{Common: play.Common{Where: game.Where{Searching: true}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Flags.Sweeping }},
		{"party menu", play.Assembling{Common: play.Common{Party: []game.Member{{}},
			Allowed: play.Allowed{Plain: []play.Action{play.NewCharacter{}, play.RemoveCharacter{}, play.QuickStart{}, play.Begin{}}}}},
			game.Data{Texts: map[string]string{"Remove Character": "Purge Character"}}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Mode == game.ModeParty && slices.Equal(optionLabels(s.Pending),
					[]string{"Create Character", "Purge Character", "View Character", "Load Saved Game", "Save Current Game", "Quick Start", "Begin Adventuring"})
			}},
		{"creation step", play.Assembling{Draft: &play.Draft{Step: "career", Ask: &play.Question{Replies: []string{"Rogue", "Medic"}}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []play.Action{play.SetChoice{Field: "career", N: 1}},
			func(s shown, _ play.Result) bool {
				return s.Pending.Kind == game.Choice && optionLabels(s.Pending)[1] == "Medic"
			}},
		{"creation number", play.Assembling{Draft: &play.Draft{Step: "str", Ask: &play.Question{Number: true, Default: 9, Max: 18}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Kind == game.Number && s.Pending.Default == 9 && s.Pending.Max == 18
			}},
		{"question over exploring", play.Exploring{Common: play.Common{Adventure: true, Ask: &play.Question{Text: "Bash it?", YesNo: true}},
			Map: play.ExploreMap{Walls: play.Walls{Side: [][][4]int{{{}}}}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []play.Action{play.Reply{N: 1}},
			func(s shown, _ play.Result) bool {
				return s.Mode == game.ModeAsk && s.Walls != nil && s.Window == game.WindowScene && s.Holding == game.HoldNone && s.Flags.Adventure && s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Bash it?"
			}},
		{"question over fighting", play.Fighting{Common: play.Common{Ask: &play.Question{Replies: []string{"a"}}}}, game.Data{},
			[]any{tAnswer{N: 'A'}}, []play.Action{play.Reply{N: 'A'}},
			func(s shown, _ play.Result) bool { return s.Combat != nil && s.Pending != nil }},
		{"series party menu", play.Assembling{Common: play.Common{Party: []game.Member{{}},
			Allowed: play.Allowed{Plain: []play.Action{play.RemoveCharacter{}, play.EditIcon{}, play.Begin{}}}}},
			game.Data{PartyMenu: []game.MenuKey{{Label: "Begin Adventuring", Key: 'B'}, {Label: "Remove Character", Key: 'P'}, {Label: "Icon Select", Key: 'I'}}}, nil, nil,
			func(s shown, _ play.Result) bool {
				o := s.Pending.Options
				return slices.Equal(optionLabels(s.Pending), []string{"Begin Adventuring", "Remove Character", "Icon Select"}) && o[1].Key == 'P'
			}},
		{"continue over fighting", play.Fighting{Common: play.Common{Status: "Press button or enter to continue.", Ask: &play.Question{Continue: true}}},
			game.Data{}, []any{tAnswer{}}, []play.Action{play.Continue{}},
			func(s shown, _ play.Result) bool {
				return s.Combat != nil && s.Pending.Kind == game.Continue && s.Page.Prompt[0].Text != ""
			}},
		{"heading over exploring", play.Exploring{Common: play.Common{Ask: &play.Question{Heading: true}}},
			game.Data{}, []any{tAnswer{N: 2}}, []play.Action{play.Heading{Dir: 2}},
			func(s shown, _ play.Result) bool {
				return s.Mode == game.ModeAsk && s.Pending.Kind == game.Direction && s.Pending.Options[0].Do == play.Leave{}
			}},
		{"draft sheet figures", play.Assembling{Draft: &play.Draft{Member: game.Member{Page: &game.SheetPage{Age: 23}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.SheetPage != nil && s.SheetPage.Age == 23 }},
		{"no save at party menu", play.Assembling{Common: play.Common{NoSave: true, NoLoad: true, Party: []game.Member{{Name: "A"}}}}, game.Data{},
			nil, nil,
			func(s shown, res play.Result) bool {
				for _, o := range s.Pending.Options {
					if o.Label == "Save Current Game" || o.Label == "Load Saved Game" {
						return false
					}
				}
				return true
			}},
		{"free aim is gobox's", play.Fighting{}, game.Data{}, []any{play.Aim{At: image.Pt(3, 4)}}, nil,
			func(s shown, res play.Result) bool { return res.Err == nil }},
		{"a creation list has no keys", play.Assembling{Draft: &play.Draft{Step: "race", Ask: &play.Question{Replies: []string{"Martian", "Mercurian"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 0
			}},
		{"bars keep keys in a list", play.ShipFight{Menu: &play.Question{Replies: []string{"Quit", "View"}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 'Q' && s.Pending.Options[1].Key == 'V'
			}},
		{"ship bar keeps the game's keys", play.ShipFight{Menu: &play.Question{Replies: []string{"Close In", "Character"}, Keys: []rune{0, 'C'}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 'C' && s.Pending.Options[1].Key == 'C' // the game's own (BR2 space: C Character)
			}},
		{"team menu keys in a list", play.Assembling{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.NewCharacter{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending.Options[0].Key == 'C' }},
		{"list Done keeps its key", play.Scene{Kind: play.SceneScreen, Common: play.Common{Ask: &play.Question{List: true, Replies: []string{"Agility", "Next", "Done"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 'N' && s.Pending.Options[2].Key == 'D'
			}},
		{"menu over the view keeps keys", play.Scene{Kind: play.SceneTalk, Common: play.Common{Ask: &play.Question{Replies: []string{"Move", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 'M' && s.Pending.Options[1].Key == 'S'
			}},
		{"question pick is a list in a list", play.Exploring{Common: play.Common{Ask: &play.Question{List: true, Replies: []string{"* BAR", "CLINIC"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 0
			}},
		{"question menu keeps keys", play.Exploring{Common: play.Common{Ask: &play.Question{Replies: []string{"Rest", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Options[0].Key == 'R' && s.Pending.Options[1].Key == 'S'
			}},
		{"lone aim asks for the square", play.Fighting{Common: play.Common{Status: "Aim:", Allowed: play.Allowed{Aimed: []play.Aimed{{Action: play.Shoot{}, Targets: 1}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending != nil && s.Pending.Kind == game.Square && s.Pending.Prompt == "Aim:" && s.Combat.Turn.Aim
			}},
		{"no Aim on a turn without a menu", play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				for _, o := range s.Combat.Turn.Options {
					if o.Label == "Aim" {
						return false
					}
				}
				return true
			}},
		{"scene menu is an ask", play.Scene{Kind: play.SceneTalk, Common: play.Common{Ask: &play.Question{Replies: []string{"Move", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Mode == game.ModeAsk && s.Pending != nil }},
		{"draft question is an ask", play.Assembling{Common: play.Common{Ask: &play.Question{Replies: []string{"A", "B"}}}, Draft: &play.Draft{Step: "sheet"}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Mode == game.ModeAsk }},
		{"scene list keeps its default", play.Scene{Kind: play.SceneScreen, Common: play.Common{Ask: &play.Question{List: true, Replies: []string{"a", "b", "c"}, Default: 2}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending != nil && s.Pending.Default == 2 }},
		{"marks are no key", play.Exploring{Common: play.Common{Ask: &play.Question{Replies: []string{"* BAR"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending.Options[0].Key == 'B' }},
		{"roster pick prompt", play.Scene{Common: play.Common{Status: "Add a character: ", Ask: &play.Question{Text: "Add a character: ", List: true, Replies: []string{"Wilma", "Buck"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Pending.Prompt == "Add a character: " && s.Pending.Options[0].Key == 0
			}},
		{"turn keys unique", play.Fighting{Common: play.Common{Allowed: play.Allowed{Moves: []game.Reach{{}}, Plain: []play.Action{play.DoManeuver{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				o := s.Combat.Turn.Options
				return len(o) == 2 && o[0].Key == 'M' && o[1].Key == 0
			}},
		{"keeper portrait", play.Trading{Common: play.Common{Portrait: &play.Portrait{Head: game.Picture{File: "HEAD2", Block: 3}, Body: game.Picture{File: "BODY2", Block: 3}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Window == game.WindowPicture && len(s.Story) == 2 && s.Story[0].File == "HEAD2"
			}},
		{"adventure on town loot", play.Looting{Common: play.Common{Adventure: true}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Flags.Adventure }},
		{"fight log from text", play.Fighting{Common: play.Common{Text: []string{"YOU GET INTO A BRAWL."}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return slices.Equal(s.Combat.Log, []string{"YOU GET INTO A BRAWL."})
			}},
		{"aimed attack", play.Fighting{Common: play.Common{Allowed: play.Allowed{Aimed: []play.Aimed{{Action: play.Attack{}, Squares: []image.Point{{2, 3}}}}}}}, game.Data{VerbOrder: []string{"Attack", "Aim"}}, // the series has Aim: gobox's free cursor
			[]any{play.Attack{}, play.Aim{At: image.Pt(2, 3)}}, []play.Action{play.Attack{At: &image.Point{2, 3}}},
			func(s shown, _ play.Result) bool {
				var att, aim bool
				for _, o := range s.Combat.Turn.Options {
					_, a := o.Do.(play.Attack)
					_, b := o.Do.(play.Aim)
					att, aim = att || a, aim || b
				}
				return att && aim
			}},
		{"door over exploring", play.Exploring{Door: &game.Door{}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Mode == game.ModeAsk && s.Holding == game.HoldDoor }},
		{"no adventure in shop", play.Trading{Common: play.Common{Where: game.Where{Mode: "town"}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return !s.Flags.Adventure }},
		{"temple keeper portrait", play.Trading{Kind: play.TradeTemple, Common: play.Common{Portrait: &play.Portrait{Head: game.Picture{File: "HEAD2", Block: 3}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Window == game.WindowPicture && len(s.Story) == 2 }},
		{"no free aim", play.Fighting{}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return len(s.Combat.Turn.Options) == 0 }},
		{"camp layers and day", play.Camping{Common: play.Common{Day: true, Layers: []game.Layer{{FrameMs: 100}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				return s.Window == game.WindowPicture && len(s.Page.Layers) == 1 && s.Flags.Day
			}},
		{"camp picture", play.Camping{Common: play.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Window == game.WindowPicture }},
		{"reroll yes/no", play.Assembling{Draft: &play.Draft{Step: "roll", Ask: &play.Question{YesNo: true, Text: "Reroll stats?"}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []play.Action{play.SetChoice{Field: "roll", N: 1}},
			func(s shown, _ play.Result) bool {
				return s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Reroll stats?"
			}},
		{"no hotkeys", play.Assembling{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Begin{}}}},
			Draft: &play.Draft{Step: "race", Ask: &play.Question{Replies: []string{"Terran"}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending.Options[0].Key == 0 }},
		{"continue only", play.Scene{Kind: play.SceneScreen, Screen: &game.Screen{Kind: game.ScreenShipSheet},
			Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"ship fight continue", play.ShipFight{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Continue{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"move verb", play.Fighting{Common: play.Common{Allowed: play.Allowed{Moves: []game.Reach{{}}, Plain: []play.Action{play.Delay{}, play.EndTurn{}}}}},
			game.Data{}, nil, nil,
			func(s shown, _ play.Result) bool {
				var l []string
				for _, o := range s.Combat.Turn.Options {
					l = append(l, o.Label)
				}
				return slices.Equal(l, []string{"Move", "Wait", "End"}) // no Aim the series does not place
			}},
		{"verb keys from data", play.Fighting{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.Aid{}}}}},
			game.Data{VerbKeys: map[string]rune{"Aid": 'H'}}, []any{tAnswer{N: 'H'}}, []play.Action{play.Aid{}},
			func(s shown, _ play.Result) bool { return s.Combat.Turn.Options[0].Key == 'H' }},
		{"rest capped", play.Camping{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{D: 5 * time.Minute}}}}},
			game.Data{}, []any{tRest{Minutes: 60}}, nil,
			func(s shown, _ play.Result) bool { return s.Pacing.Pace == game.PaceRest }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fake{st: c.st}
			r := newRules(f, c.data)
			var res play.Result
			for _, d := range c.do {
				if res = r.send(d); res.Err != nil {
					t.Fatal(res.Err)
				}
			}
			if len(f.sent) != len(c.sent) || c.sent != nil && !same(f.sent[0], c.sent[0]) {
				t.Fatalf("sent %#v, want %#v", f.sent, c.sent)
			}
			if c.ok != nil && !c.ok(shownOf(r), res) {
				t.Fatalf("state %+v / %+v", shownOf(r), res)
			}
		})
	}
}

// TestRestChunkCap: a chunk is no longer than the allowed rest.
func TestRestChunkCap(t *testing.T) {
	f := &fake{st: play.Camping{Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{D: 5 * time.Minute}}}}}}
	r := newRules(f, game.Data{})
	r.send(tRest{Minutes: 60})
	r.Step()
	if f.sent[0] != (play.AdvanceTimeBy{D: 5 * time.Minute}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestCombatLogNote: the fight's text builds the log; a refusal is the toast.
func TestCombatLogNote(t *testing.T) {
	f := &fake{st: play.Fighting{}, out: play.Outcome{Events: []play.Event{play.Narrated{Text: "Hits!"}, play.Refused{Why: "no"}}}}
	r := newRules(f, game.Data{})
	r.send(play.EndTurn{})
	if s := shownOf(r); !slices.Equal(s.Combat.Log, []string{"Hits!"}) || s.Note.Text != "no" {
		t.Fatalf("log %v note %v", s.Combat.Log, s.Note)
	}
}

// TestGameSlots: the game's own slots list after gobox's and load by index.
func TestGameSlots(t *testing.T) {
	g := &slottedFake{fake: fake{st: play.Assembling{}}}
	r := newRules(g, game.Data{})
	r.dir = t.TempDir()
	s := shownOf(r)
	if len(s.Saves) != 1 || s.Saves[0].Info != "old" {
		t.Fatalf("saves %v", s.Saves)
	}
	if res := r.send(tLoad{Slot: 0}); res.Err != nil || g.loaded != 1 {
		t.Fatalf("load %+v %d", res, g.loaded)
	}
}

func TestStudyRestAndSaveDir(t *testing.T) {
	dir := t.TempDir()
	g := &fake{st: play.Camping{Study: 3 * time.Hour, Common: play.Common{Allowed: play.Allowed{Plain: []play.Action{play.AdvanceTimeBy{}}}}}}
	r := newRules(g, game.Data{SaveDir: dir})
	r.send(tRest{})
	if r.rest != 3*time.Hour || r.dir != dir {
		t.Fatalf("rest %v dir %q", r.rest, r.dir)
	}
	g.st = play.Camping{Common: play.Common{NoSave: true, NoLoad: true}}
	if f := shownOf(r).Flags; f == nil || !f.NoSave || !f.NoLoad {
		t.Fatal("NoSave/NoLoad not published")
	}
	if r.send(tSave{Name: "x"}).Err == nil {
		t.Fatal("saved where the rules refuse it")
	}
}
