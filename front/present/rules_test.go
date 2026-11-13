package present

import (
	"errors"
	"image"
	"reflect"
	"slices"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// fake is a play.Game whose state and outcome the test sets; it keeps what
// Do was sent.
type fake struct {
	st   play.State
	seen play.State // st at the last Version
	ver  uint64
	out  play.Outcome
	err  error
	sent []play.Action
	data []byte
}

func (f *fake) State() play.State { return f.st }

// Version changes iff the State does (tests assign st directly).
func (f *fake) Version() uint64 {
	if !reflect.DeepEqual(f.seen, f.st) {
		f.seen, f.ver = f.st, f.ver+1
	}
	return f.ver
}

// testVersions numbers the States of the fakes that build theirs on each
// call: every read is a new one.
var testVersions uint64

// fresh is the next version of a fake that builds a new State on each read.
func fresh() uint64 { testVersions++; return testVersions }
func (f *fake) Do(a play.Action) (play.Outcome, error) {
	f.sent = append(f.sent, a)
	return f.out, f.err
}
func (f *fake) Save() ([]byte, error) { return []byte("saved"), nil }
func (f *fake) Load(b []byte) error   { f.data = b; return nil }

func TestAdapter(t *testing.T) {
	wall := errors.New("a wall")
	talk := play.Scene{Kind: play.SceneTalk, Text: []string{"Who goes?"}, Common: play.Common{Ask: &play.Question{Replies: []string{"Friend", "Foe"}}}}
	fight := play.Fighting{Field: play.Battlefield{Turn: 1, Units: []game.Unit{{ID: 1, Member: 0}, {ID: 2, Member: -1, Foe: true}}}}
	hand := play.Scene{Kind: play.SceneHandout, Handout: "Journal 1", Lines: []string{"line"}}
	cases := []struct {
		name  string
		f     *fake
		d     any
		check func(t *testing.T, f *fake, r *frontState, res play.Result)
	}{
		{"wall refused", &fake{st: play.Exploring{}, err: wall},
			play.Arrive{From: image.Pt(3, 3), To: image.Pt(3, 2)},
			func(t *testing.T, f *fake, r *frontState, res play.Result) {
				if res.Err != wall || !slices.Equal(f.sent, []play.Action{play.Arrive{From: image.Pt(3, 3), To: image.Pt(3, 2)}}) {
					t.Fatalf("err %v sent %v", res.Err, f.sent)
				}
			}},
		{"scene reply", &fake{st: talk}, tAnswer{N: 1},
			func(t *testing.T, f *fake, r *frontState, res play.Result) {
				if a := r.pending(r.g.State()); a == nil || a.Kind != game.Choice || len(a.Options) != 2 || a.Options[1].Label != "Foe" {
					t.Fatalf("pending %+v", a)
				}
				if s := shownOf(r); s.Mode != game.ModeAsk || s.Page == nil || s.Pending == nil {
					t.Fatalf("state %+v", s)
				}
				if res.Err != nil || !slices.Equal(f.sent, []play.Action{play.Reply{N: 1}}) {
					t.Fatalf("err %v sent %v", res.Err, f.sent)
				}
			}},
		{"end turn events", &fake{st: fight, out: play.Outcome{Events: []play.Event{
			play.Walked{Unit: 2, Path: []game.Step{{At: image.Pt(1, 1)}}},
			play.Attacked{From: 2, To: 1, Hit: true, Damage: 4}, play.Fell{Unit: 1}}}},
			play.EndTurn{},
			func(t *testing.T, f *fake, r *frontState, res play.Result) {
				var kinds []game.EventKind
				for _, e := range res.Events {
					kinds = append(kinds, e.Kind)
				}
				want := []game.EventKind{game.EvMove, game.EvAttack, game.EvHit, game.EvDamage, game.EvDown}
				if !slices.Equal(kinds, want) || res.Events[3].N != 4 || res.Events[1].Ms == 0 {
					t.Fatalf("events %+v", res.Events)
				}
				if s := shownOf(r); s.Mode != game.ModeCombat || s.Combat.Turn.Unit != 1 {
					t.Fatalf("state %+v", s.Combat)
				}
			}},
		{"handout read", &fake{st: hand}, tAnswer{N: 0},
			func(t *testing.T, f *fake, r *frontState, res play.Result) {
				s, s2 := shownOf(r), shownOf(r)
				if s.Read == nil || s.Read.Title != "Journal 1" || s.Read != s2.Read || s.Pending != nil {
					t.Fatalf("read %+v pending %+v", s.Read, s.Pending)
				}
				if !slices.Equal(f.sent, []play.Action{play.Continue{}}) {
					t.Fatalf("sent %v", f.sent)
				}
			}},
		{"step is empty", &fake{st: play.Exploring{}}, tSetSpeed{},
			func(t *testing.T, f *fake, r *frontState, res play.Result) {
				if st := r.Step(); st.Err != nil || len(st.Events) != 0 || len(f.sent) != 0 {
					t.Fatalf("step %+v sent %v", st, f.sent)
				}
			}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := newRules(c.f, game.Data{})
			c.check(t, c.f, r, r.send(c.d))
		})
	}
}

func TestAimAndSlots(t *testing.T) {
	cast := play.Cast{Spell: 7}
	f := &fake{st: play.Fighting{Common: play.Common{Allowed: play.Allowed{
		Aimed: []play.Aimed{{Action: cast, Squares: []image.Point{{2, 2}}, Targets: 1}}}}}}
	r := newRules(f, game.Data{})
	r.dir = t.TempDir()
	if res := r.send(play.Cast{Spell: 7}); res.Err != nil || len(f.sent) != 0 || !shownOf(r).Combat.Turn.Aim {
		t.Fatalf("aim start %+v sent %v", res, f.sent)
	}
	if res := r.send(play.Aim{At: image.Pt(5, 5)}); res.Err != nil || len(f.sent) != 0 || !shownOf(r).Combat.Turn.Aim { // not offered: no word, the aim stays (the original's Center)
		t.Fatalf("aim out of range: %+v sent %v", res, f.sent)
	}
	r.send(play.Aim{At: image.Pt(2, 2)})
	if len(f.sent) != 1 || f.sent[0].(play.Cast).Ats[0] != image.Pt(2, 2) {
		t.Fatalf("sent %v", f.sent)
	}
	if res := r.send(tSave{Name: "one"}); res.Err != nil {
		t.Fatal(res.Err)
	}
	if res := r.send(tLoad{Slot: 0}); res.Err != nil || string(f.data) != "saved" {
		t.Fatalf("load %+v %q", res, f.data)
	}
}

type artGame struct{ fake }

func (artGame) Art() map[string]game.Picture { return map[string]game.Picture{"w": {File: "live"}} }

func TestArtIsLive(t *testing.T) {
	r := newRules(&artGame{}, game.Data{})
	if r.Art()["w"].File != "live" {
		t.Fatal("game art not read live")
	}
}

func TestYesNoDefault(t *testing.T) {
	r := newRules(&fake{}, game.Data{})
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
	s := shownOf(newRules(f, game.Data{}))
	if s.Page == nil || len(s.Page.Layers) == 0 || s.Page.Layers[len(s.Page.Layers)-1].Picture.Block != 3 {
		t.Fatalf("no first step picture: %+v", s.Page)
	}
}

func TestShowMemberOpensSheet(t *testing.T) {
	r := newRules(&fake{st: play.Exploring{}}, game.Data{})
	if r.send(tShowMember{Member: 2}); r.sheet != 2 {
		t.Fatalf("sheet %d", r.sheet)
	}
}

// playGame is the presenter's game over a fake play.Game in st: a test
// sees the play actions the presenter's calls become in its sent.
func playGame(st play.State) (*classicGame, *fake) {
	f := &fake{st: st}
	return &classicGame{Game: NewGame(f, game.Data{}), p: &Presenter{}}, f
}

// bought is the item of the first Buy sent (-1: none).
func bought(f *fake) int {
	for _, a := range f.sent {
		if b, ok := a.(play.Buy); ok {
			return int(b.Item)
		}
	}
	return -1
}
