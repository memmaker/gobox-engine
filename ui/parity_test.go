package ui

import (
	"gobox/adventure"
	"slices"
	"testing"
	"time"

	"gobox/game"
	"gobox/rules"
)

// slottedFake is fake with the game's own save slots (game.Slots).
type slottedFake struct {
	fake
	loaded int
}

func (s *slottedFake) Slots() []game.SaveSlot { return []game.SaveSlot{{Name: "A", Info: "old"}} }
func (s *slottedFake) LoadSlot(n int) error   { s.loaded = n + 1; return nil }

func optionLabels(a *game.Ask) (l []string) {
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
	withScripts(t, &adventure.Sequence{Steps: []adventure.SequenceStep{{Kind: adventure.StepPicture, Song: "title",
		Layers: []adventure.Layer{{Art: adventure.Art{File: "TITLE", Block: 1}}}}}}, nil)
	title := game.Scene{Kind: game.SceneTitle, Intro: &game.Intro{Pictures: []int{1, -1}}}
	cases := []struct {
		name string
		st   game.State
		data game.Data
		do   []any
		sent []game.Action
		ok   func(s shown, res game.Result) bool
	}{
		{"title step layers and song", title, game.Data{}, []any{tTitlePicture{N: 1}}, nil,
			func(s shown, res game.Result) bool {
				return len(s.Page.Layers) == 1 && s.Page.Layers[0].File == "TITLE" && res.Events[0].Kind == game.EvSong
			}},
		{"title credits", title, game.Data{}, []any{tTitleCredits{Page: 1}}, nil,
			func(s shown, _ game.Result) bool {
				return s.Flags.Credits && s.Flags.CreditPage == 1 && len(s.Page.Layers) == 0
			}},
		{"holding door", game.Exploring{Door: &game.Door{}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Holding == game.HoldDoor }},
		{"holding fight", game.Fighting{}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Holding == game.HoldFight }},
		{"holding temple", game.Trading{Kind: game.TradeTemple}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Holding == game.HoldTemple }},
		{"window picture", game.Exploring{Common: game.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Window == game.WindowPicture && s.Picture.File == pic.File
			}},
		{"window map, no walls", game.Exploring{Map: game.ExploreMap{Overland: true}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Window == game.WindowMap && s.Walls == nil }},
		{"status line and pane", game.Fighting{Common: game.Common{Status: "Searching", Text: []string{"a", "b"}, Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Page.Prompt[0].Text == "Searching" && len(s.Page.Pane) == 2 && s.LastText == "b" && s.Picture.File == pic.File
			}},
		{"camp flags and automap", game.Camping{Common: game.Common{Adventure: true, Allowed: game.Allowed{Plain: []game.Action{game.Order{}}},
			Automap: game.Automap{Size: rules.Size{W: 16, H: 16}}, Party: []game.Member{{CanCast: true}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Flags.Adventure && s.Flags.Ordering && s.Flags.CanCast && s.Map.Size.W == 16
			}},
		{"sweeping", game.Exploring{Common: game.Common{Where: game.Where{Searching: true}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Flags.Sweeping }},
		{"party menu", game.Assembling{Common: game.Common{Party: []game.Member{{}},
			Allowed: game.Allowed{Plain: []game.Action{game.NewCharacter{}, game.RemoveCharacter{}, game.QuickStart{}, game.Begin{}}}}},
			game.Data{Texts: map[string]string{"Remove Character": "Purge Character"}}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Mode == game.ModeParty && slices.Equal(optionLabels(s.Pending),
					[]string{"Create Character", "Purge Character", "View Character", "Load Saved Game", "Save Current Game", "Quick Start", "Begin Adventuring"})
			}},
		{"creation step", game.Assembling{Draft: &game.Draft{Step: "career", Ask: &game.Question{Replies: []string{"Rogue", "Medic"}}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []game.Action{game.SetChoice{Field: "career", N: 1}},
			func(s shown, _ game.Result) bool {
				return s.Pending.Kind == game.Choice && optionLabels(s.Pending)[1] == "Medic"
			}},
		{"creation number", game.Assembling{Draft: &game.Draft{Step: "str", Ask: &game.Question{Number: true, Default: 9, Max: 18}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Kind == game.Number && s.Pending.Default == 9 && s.Pending.Max == 18
			}},
		{"question over exploring", game.Exploring{Common: game.Common{Adventure: true, Ask: &game.Question{Text: "Bash it?", YesNo: true}},
			Map: game.ExploreMap{Walls: game.MapWalls{Side: [][][4]int{{{}}}}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []game.Action{game.Reply{N: 1}},
			func(s shown, _ game.Result) bool {
				return s.Mode == game.ModeAsk && s.Walls != nil && s.Window == game.WindowScene && s.Holding == game.HoldNone && s.Flags.Adventure && s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Bash it?"
			}},
		{"question over fighting", game.Fighting{Common: game.Common{Ask: &game.Question{Replies: []string{"a"}}}}, game.Data{},
			[]any{tAnswer{N: 'A'}}, []game.Action{game.Reply{N: 'A'}},
			func(s shown, _ game.Result) bool { return s.Combat != nil && s.Pending != nil }},
		{"series party menu", game.Assembling{Common: game.Common{Party: []game.Member{{}},
			Allowed: game.Allowed{Plain: []game.Action{game.RemoveCharacter{}, game.EditIcon{}, game.Begin{}}}}},
			game.Data{PartyMenu: []game.MenuKey{{Label: "Begin Adventuring", Key: 'B'}, {Label: "Remove Character", Key: 'P'}, {Label: "Icon Select", Key: 'I'}}}, nil, nil,
			func(s shown, _ game.Result) bool {
				o := s.Pending.Options
				return slices.Equal(optionLabels(s.Pending), []string{"Begin Adventuring", "Remove Character", "Icon Select"}) && o[1].Key == 'P'
			}},
		{"continue over fighting", game.Fighting{Common: game.Common{Status: "Press button or enter to continue.", Ask: &game.Question{Continue: true}}},
			game.Data{}, []any{tAnswer{}}, []game.Action{game.ContinueAction{}},
			func(s shown, _ game.Result) bool {
				return s.Combat != nil && s.Pending.Kind == game.Continue && s.Page.Prompt[0].Text != ""
			}},
		{"heading over exploring", game.Exploring{Common: game.Common{Ask: &game.Question{Heading: true}}},
			game.Data{}, []any{tAnswer{N: 2}}, []game.Action{game.Heading{Dir: 2}},
			func(s shown, _ game.Result) bool {
				return s.Mode == game.ModeAsk && s.Pending.Kind == game.Direction && s.Pending.Options[0].Do == game.Leave{}
			}},
		{"draft sheet figures", game.Assembling{Draft: &game.Draft{Member: game.Member{Page: &game.SheetPage{Age: 23}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.SheetPage != nil && s.SheetPage.Age == 23 }},
		{"no save at party menu", game.Assembling{Common: game.Common{NoSave: true, NoLoad: true, Party: []game.Member{{Name: "A"}}}}, game.Data{},
			nil, nil,
			func(s shown, res game.Result) bool {
				for _, o := range s.Pending.Options {
					if o.Label == "Save Current Game" || o.Label == "Load Saved Game" {
						return false
					}
				}
				return true
			}},
		{"free aim is gobox's", game.Fighting{}, game.Data{}, []any{game.Aim{At: rules.BoardPos{X: 3, Y: 4}}}, nil,
			func(s shown, res game.Result) bool { return res.Err == nil }},
		{"a creation list has no keys", game.Assembling{Draft: &game.Draft{Step: "race", Ask: &game.Question{Replies: []string{"Martian", "Mercurian"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 0
			}},
		{"bars keep keys in a list", game.ShipFight{Menu: &game.Question{Replies: []string{"Quit", "View"}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'Q' && s.Pending.Options[1].Key == 'V'
			}},
		{"ship bar keeps the game's keys", game.ShipFight{Menu: &game.Question{Replies: []string{"Close In", "Character"}, Keys: []rune{0, 'C'}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'C' && s.Pending.Options[1].Key == 'C' // the game's own (BR2 space: C Character)
			}},
		{"team menu keys in a list", game.Assembling{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.NewCharacter{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending.Options[0].Key == 'C' }},
		{"list Done keeps its key", game.Scene{Kind: game.SceneScreen, Common: game.Common{Ask: &game.Question{List: true, Replies: []string{"Agility", "Next", "Done"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 'N' && s.Pending.Options[2].Key == 'D'
			}},
		{"menu over the view keeps keys", game.Scene{Kind: game.SceneTalk, Common: game.Common{Ask: &game.Question{Replies: []string{"Move", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'M' && s.Pending.Options[1].Key == 'S'
			}},
		{"question pick is a list in a list", game.Exploring{Common: game.Common{Ask: &game.Question{List: true, Replies: []string{"* BAR", "CLINIC"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 0 && s.Pending.Options[1].Key == 0
			}},
		{"question menu keeps keys", game.Exploring{Common: game.Common{Ask: &game.Question{Replies: []string{"Rest", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Options[0].Key == 'R' && s.Pending.Options[1].Key == 'S'
			}},
		{"lone aim asks for the square", game.Fighting{Common: game.Common{Status: "Aim:", Allowed: game.Allowed{Aimed: []game.Aimed{{Action: game.Shoot{}, Targets: 1}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending != nil && s.Pending.Kind == game.Square && s.Pending.Prompt == "Aim:" && s.Combat.Turn.Aim
			}},
		{"no Aim on a turn without a menu", game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.ContinueAction{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				for _, o := range s.Combat.Turn.Options {
					if o.Label == "Aim" {
						return false
					}
				}
				return true
			}},
		{"scene menu is an ask", game.Scene{Kind: game.SceneTalk, Common: game.Common{Ask: &game.Question{Replies: []string{"Move", "Save"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Mode == game.ModeAsk && s.Pending != nil }},
		{"draft question is an ask", game.Assembling{Common: game.Common{Ask: &game.Question{Replies: []string{"A", "B"}}}, Draft: &game.Draft{Step: "sheet"}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Mode == game.ModeAsk }},
		{"scene list keeps its default", game.Scene{Kind: game.SceneScreen, Common: game.Common{Ask: &game.Question{List: true, Replies: []string{"a", "b", "c"}, Default: 2}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending != nil && s.Pending.Default == 2 }},
		{"marks are no key", game.Exploring{Common: game.Common{Ask: &game.Question{Replies: []string{"* BAR"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending.Options[0].Key == 'B' }},
		{"roster pick prompt", game.Scene{Common: game.Common{Status: "Add a character: ", Ask: &game.Question{Text: "Add a character: ", List: true, Replies: []string{"Wilma", "Buck"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Pending.Prompt == "Add a character: " && s.Pending.Options[0].Key == 0
			}},
		{"turn keys unique", game.Fighting{Common: game.Common{Allowed: game.Allowed{Moves: []game.Reach{{}}, Plain: []game.Action{game.DoManeuver{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				o := s.Combat.Turn.Options
				return len(o) == 2 && o[0].Key == 'M' && o[1].Key == 0
			}},
		{"keeper portrait", game.Trading{Common: game.Common{Portrait: &game.Portrait{Head: game.Picture{File: "HEAD2", Block: 3}, Body: game.Picture{File: "BODY2", Block: 3}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Window == game.WindowPicture && len(s.Story) == 2 && s.Story[0].File == "HEAD2"
			}},
		{"adventure on town loot", game.Looting{Common: game.Common{Adventure: true}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Flags.Adventure }},
		{"fight log from text", game.Fighting{Common: game.Common{Text: []string{"YOU GET INTO A BRAWL."}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return slices.Equal(s.Combat.Log, []string{"YOU GET INTO A BRAWL."})
			}},
		{"aimed attack", game.Fighting{Common: game.Common{Allowed: game.Allowed{Aimed: []game.Aimed{{Action: game.Attack{}, Squares: []rules.BoardPos{{X: 2, Y: 3}}}}}}}, game.Data{VerbOrder: []string{"Attack", "Aim"}}, // the series names Aim: no extra verb for it
			[]any{game.Attack{}, game.Aim{At: rules.BoardPos{X: 2, Y: 3}}}, []game.Action{game.Attack{At: &rules.BoardPos{X: 2, Y: 3}}},
			func(s shown, _ game.Result) bool {
				var att, aim bool
				for _, o := range s.Combat.Turn.Options {
					_, a := o.Do.(game.Attack)
					_, b := o.Do.(game.Aim)
					att, aim = att || a, aim || b
				}
				return att && !aim // no free-cursor Aim verb
			}},
		{"door over exploring", game.Exploring{Door: &game.Door{}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Mode == game.ModeAsk && s.Holding == game.HoldDoor }},
		{"no adventure in shop", game.Trading{Common: game.Common{Where: game.Where{Mode: "town"}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return !s.Flags.Adventure }},
		{"temple keeper portrait", game.Trading{Kind: game.TradeTemple, Common: game.Common{Portrait: &game.Portrait{Head: game.Picture{File: "HEAD2", Block: 3}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Window == game.WindowPicture && len(s.Story) == 2 }},
		{"no free aim", game.Fighting{}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return len(s.Combat.Turn.Options) == 0 }},
		{"camp picture and day", game.Camping{Common: game.Common{Day: true, Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				return s.Window == game.WindowPicture && len(s.Page.Layers) == 1 && s.Flags.Day
			}},
		{"camp picture", game.Camping{Common: game.Common{Picture: pic}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Window == game.WindowPicture }},
		{"reroll yes/no", game.Assembling{Draft: &game.Draft{Step: "roll", Ask: &game.Question{YesNo: true, Text: "Reroll stats?"}}}, game.Data{},
			[]any{tAnswer{N: 1}}, []game.Action{game.SetChoice{Field: "roll", N: 1}},
			func(s shown, _ game.Result) bool {
				return s.Pending.Kind == game.YesNo && s.Pending.Prompt == "Reroll stats?"
			}},
		{"no hotkeys", game.Assembling{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Begin{}}}},
			Draft: &game.Draft{Step: "race", Ask: &game.Question{Replies: []string{"Terran"}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending.Options[0].Key == 0 }},
		{"continue only", game.Scene{Kind: game.SceneScreen, Screen: &game.Screen{Kind: game.ScreenShipSheet},
			Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.ContinueAction{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"ship fight continue", game.ShipFight{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.ContinueAction{}}}}}, game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool { return s.Pending != nil && s.Pending.Kind == game.Continue }},
		{"move verb", game.Fighting{Common: game.Common{Allowed: game.Allowed{Moves: []game.Reach{{}}, Plain: []game.Action{game.Delay{}, game.EndTurn{}}}}},
			game.Data{}, nil, nil,
			func(s shown, _ game.Result) bool {
				var l []string
				for _, o := range s.Combat.Turn.Options {
					l = append(l, o.Label)
				}
				return slices.Equal(l, []string{"Move", "Wait", "End"}) // no Aim the series does not place
			}},
		{"verb keys from data", game.Fighting{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.Aid{}}}}},
			game.Data{VerbKeys: map[string]rune{"Aid": 'H'}}, []any{tAnswer{N: 'H'}}, []game.Action{game.Aid{}},
			func(s shown, _ game.Result) bool { return s.Combat.Turn.Options[0].Key == 'H' }},
		{"rest capped", game.Camping{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.AdvanceTimeBy{D: 5 * time.Minute}}}}},
			game.Data{}, []any{tRest{Minutes: 60}}, nil,
			func(s shown, _ game.Result) bool { return s.Pacing.Pace == game.PaceRest }},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &fake{st: c.st}
			r := newRules(f, c.data)
			var res game.Result
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
	f := &fake{st: game.Camping{Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.AdvanceTimeBy{D: 5 * time.Minute}}}}}}
	r := newRules(f, game.Data{})
	r.send(tRest{Minutes: 60})
	r.Step()
	if f.sent[0] != (game.AdvanceTimeBy{D: 5 * time.Minute}) {
		t.Fatalf("sent %v", f.sent)
	}
}

// TestCombatLogNote: the fight's text builds the log; a refusal is the toast.
func TestCombatLogNote(t *testing.T) {
	f := &fake{st: game.Fighting{}, out: game.Outcome{Events: []game.Event{game.Narrated{Text: "Hits!"}, game.Refused{Why: "no"}}}}
	r := newRules(f, game.Data{})
	r.send(game.EndTurn{})
	if s := shownOf(r); !slices.Equal(s.Combat.Log, []string{"Hits!"}) || s.Note.Text != "no" {
		t.Fatalf("log %v note %v", s.Combat.Log, s.Note)
	}
}

// TestGameSlots: the game's own slots list after gobox's and load by index.
func TestGameSlots(t *testing.T) {
	g := &slottedFake{fake: fake{st: game.Assembling{}}}
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
	g := &fake{st: game.Camping{Study: 3 * time.Hour, Common: game.Common{Allowed: game.Allowed{Plain: []game.Action{game.AdvanceTimeBy{}}}}}}
	r := newRules(g, game.Data{SaveDir: dir})
	r.send(tRest{})
	if r.rest != 3*time.Hour || r.dir != dir {
		t.Fatalf("rest %v dir %q", r.rest, r.dir)
	}
	g.st = game.Camping{Common: game.Common{NoSave: true, NoLoad: true}}
	if f := shownOf(r).Flags; f == nil || !f.NoSave || !f.NoLoad {
		t.Fatal("NoSave/NoLoad not published")
	}
	if r.send(tSave{Name: "x"}).Err == nil {
		t.Fatal("saved where the rules refuse it")
	}
}
