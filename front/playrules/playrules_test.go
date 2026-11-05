package playrules

import (
	"errors"
	"image"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// fake is a play.Game whose state and outcome the test sets; it keeps what
// Do was sent.
type fake struct {
	st   play.State
	out  play.Outcome
	err  error
	sent []play.Action
	data []byte
}

func (f *fake) State() play.State { return f.st }
func (f *fake) Do(a play.Action) (play.Outcome, error) {
	f.sent = append(f.sent, a)
	return f.out, f.err
}
func (f *fake) Save() ([]byte, error) { return []byte("saved"), nil }
func (f *fake) Load(b []byte) error   { f.data = b; return nil }

func TestAdapter(t *testing.T) {
	wall := errors.New("a wall")
	talk := play.Scene{Kind: play.SceneTalk, Text: []string{"Who goes?"}, Replies: []string{"Friend", "Foe"}}
	fight := play.Fighting{Field: play.Battlefield{Turn: 1, Units: []play.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	hand := play.Scene{Kind: play.SceneHandout, Handout: "Journal 1", Lines: []string{"line"}}
	cases := []struct {
		name  string
		f     *fake
		d     game.Decision
		check func(t *testing.T, f *fake, r game.Rules, res game.Result)
	}{
		{"wall refused", &fake{st: play.Exploring{}, err: wall},
			game.Arrive{From: image.Pt(3, 3), To: image.Pt(3, 2)},
			func(t *testing.T, f *fake, r game.Rules, res game.Result) {
				if res.Err != wall || !slices.Equal(f.sent, []play.Action{play.Arrive{From: image.Pt(3, 3), To: image.Pt(3, 2)}}) {
					t.Fatalf("err %v sent %v", res.Err, f.sent)
				}
			}},
		{"scene reply", &fake{st: talk}, game.Answer{N: 1},
			func(t *testing.T, f *fake, r game.Rules, res game.Result) {
				if a := r.Pending(); a == nil || a.Kind != game.Choice || len(a.Options) != 2 || a.Options[1].Label != "Foe" {
					t.Fatalf("pending %+v", a)
				}
				if s := r.State(); s.Mode != game.ModeAsk || s.Page == nil || s.Pending == nil {
					t.Fatalf("state %+v", s)
				}
				if res.Err != nil || !slices.Equal(f.sent, []play.Action{play.Reply{N: 1}}) {
					t.Fatalf("err %v sent %v", res.Err, f.sent)
				}
			}},
		{"end turn events", &fake{st: fight, out: play.Outcome{Events: []play.Event{
			play.Walked{Unit: 2, Path: []game.Step{{At: image.Pt(1, 1)}}},
			play.Attacked{From: 2, To: 1, Hit: true, Damage: 4}, play.Fell{Unit: 1}}}},
			game.EndTurn{},
			func(t *testing.T, f *fake, r game.Rules, res game.Result) {
				var kinds []game.EventKind
				for _, e := range res.Events {
					kinds = append(kinds, e.Kind)
				}
				want := []game.EventKind{game.EvMove, game.EvAttack, game.EvHit, game.EvDamage, game.EvDown}
				if !slices.Equal(kinds, want) || res.Events[3].N != 4 || res.Events[1].Ms == 0 {
					t.Fatalf("events %+v", res.Events)
				}
				if s := r.State(); s.Mode != game.ModeCombat || s.Combat.Turn.Unit != 1 {
					t.Fatalf("state %+v", s.Combat)
				}
			}},
		{"handout read", &fake{st: hand}, game.Answer{N: 0},
			func(t *testing.T, f *fake, r game.Rules, res game.Result) {
				s, s2 := r.State(), r.State()
				if s.Read == nil || s.Read.Title != "Journal 1" || s.Read != s2.Read || s.Pending != nil {
					t.Fatalf("read %+v pending %+v", s.Read, s.Pending)
				}
				if !slices.Equal(f.sent, []play.Action{play.Continue{}}) {
					t.Fatalf("sent %v", f.sent)
				}
			}},
		{"step is empty", &fake{st: play.Exploring{}}, game.SetSpeed{},
			func(t *testing.T, f *fake, r game.Rules, res game.Result) {
				if st := r.Step(); st.Err != nil || len(st.Events) != 0 || len(f.sent) != 0 {
					t.Fatalf("step %+v sent %v", st, f.sent)
				}
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := New(c.f, game.Data{})
			c.check(t, c.f, r, r.Apply(c.d))
		})
	}
}

func TestAimAndSlots(t *testing.T) {
	cast := play.Cast{Spell: 7}
	f := &fake{st: play.Fighting{Common: play.Common{Allowed: play.Allowed{
		Aimed: []play.Aimed{{Action: cast, Squares: []image.Point{{2, 2}}, Targets: 1}}}}}}
	r := New(f, game.Data{}).(*rules)
	r.dir = t.TempDir()
	if res := r.Apply(game.Cast{Spell: 7}); res.Err != nil || len(f.sent) != 0 || !r.State().Combat.Turn.Aim {
		t.Fatalf("aim start %+v sent %v", res, f.sent)
	}
	if res := r.Apply(game.Aim{At: image.Pt(5, 5)}); res.Err == nil {
		t.Fatal("aim out of range taken")
	}
	r.Apply(game.Aim{At: image.Pt(2, 2)})
	if len(f.sent) != 1 || f.sent[0].(play.Cast).Ats[0] != image.Pt(2, 2) {
		t.Fatalf("sent %v", f.sent)
	}
	if res := r.Apply(game.Save{Name: "one"}); res.Err != nil {
		t.Fatal(res.Err)
	}
	if res := r.Apply(game.Load{Slot: 0}); res.Err != nil || string(f.data) != "saved" {
		t.Fatalf("load %+v %q", res, f.data)
	}
}

type artGame struct{ fake }

func (artGame) Art() map[string]game.Picture { return map[string]game.Picture{"w": {File: "live"}} }

func TestArtIsLive(t *testing.T) {
	r := New(&artGame{}, game.Data{}).(interface{ Art() map[string]game.Picture })
	if r.Art()["w"].File != "live" {
		t.Fatal("game art not read live")
	}
}

func TestYesNoDefault(t *testing.T) {
	r := New(&fake{}, game.Data{}).(*rules)
	if a := r.question(&play.Question{YesNo: true, Default: 1}, false); !a.Yes {
		t.Fatal("yes default lost")
	}
}

// Before gobox's first title step the page already has that step's
// picture: no black frame between the intro and the title.
func TestTitleFirstStepBeforeStepping(t *testing.T) {
	pic := game.Layer{Picture: game.Picture{File: "TITLE", Block: 3}} // the first step
	f := &fake{st: play.Scene{Kind: play.SceneTitle, Intro: &game.Intro{Pictures: []int{3, 1}},
		Steps: map[int]play.TitleStep{3: {Layers: []game.Layer{pic}}}}}
	s := New(f, game.Data{}).State()
	if s.Page == nil || len(s.Page.Layers) == 0 || s.Page.Layers[len(s.Page.Layers)-1].Picture.Block != 3 {
		t.Fatalf("no first step picture: %+v", s.Page)
	}
}

func TestShowMemberOpensSheet(t *testing.T) {
	r := New(&fake{st: play.Exploring{}}, game.Data{}).(*rules)
	if r.Apply(game.ShowMember{Member: 2}); r.sheet != 2 {
		t.Fatalf("sheet %d", r.sheet)
	}
}

// A Lowercase game's text keeps its case and uses the text glyphs
// (Upcase): PoD's lower-case letters showed as the charset's graphics.
func TestRunLowercaseKeepsTextGlyphs(t *testing.T) {
	r := New(&fake{}, game.Data{Lowercase: true}).(*rules)
	if got := r.run("Hello")[0]; got.Text != "Hello" || !got.Upcase {
		t.Fatalf("%+v", got)
	}
}
