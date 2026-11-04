package parity

import (
	"gobox/game"
	"os"
	"path/filepath"
	"por/rules"
	"reflect"
	"strings"
	"testing"

	"gobox/front"
	"gobox/front/classic"
)

// TestJournalPageAdapter: por's JournalPage and the adapter's give DeepEqual
// pages (and scroll limits) over every page of the journal text each scene
// reads (the clue book where it reads none), at two page sizes and scrolls.
func TestJournalPageAdapter(t *testing.T) {
	for _, scene := range []string{"journal", "talljournal", "parlay"} {
		t.Run(scene, func(t *testing.T) {
			a, b := engine(t, true, scene), engine(t, true, scene)
			g := classic.NewGame(b.Rules(), b.Data())
			defer a.Apply(front.Action{Kind: front.ActResize, X: 40, Y: 25})
			reads := []front.Read{}
			if r := a.Rules().State().Read; r != nil {
				reads = append(reads, front.Read{Title: r.Title, Lines: r.Lines, Page: r.Page})
			}
			for i := range g.ClueChapters() {
				r := a.ClueRead(i)
				reads = append(reads, *r)
			}
			if len(reads) == 0 {
				t.Fatal("nothing to read")
			}
			scrolled := false
			for _, size := range []int{40, 50} {
				act := front.Action{Kind: front.ActResize, X: size, Y: size * 5 / 8}
				a.Apply(act)
				g.Answer(act)
				for _, r := range reads {
					for p := 0; ; p++ {
						for _, off := range []int{0, 60} {
							want, wmax := a.JournalPage(r.Title, r.Lines, p, off)
							got, gmax := g.JournalPage(r.Title, r.Lines, p, off)
							scrolled = scrolled || wmax > 0
							if !reflect.DeepEqual(got, want) || gmax != wmax {
								t.Fatalf("%s page %d off %d at %d: adapter %+v max %d, por %+v max %d", r.Title, p, off, size, got, gmax, want, wmax)
							}
						}
						if want, _ := a.JournalPage(r.Title, r.Lines, p, 0); want.Page != p || p == want.Pages-1 {
							break
						}
					}
				}
			}
			if scene == "talljournal" && !scrolled {
				t.Error("no scrolling picture compared")
			}
		})
	}
}

// speakSpy records the Speak decisions the adapter takes.
type speakSpy struct {
	game.Rules
	keys []string
}

func (s *speakSpy) Apply(d game.Decision) game.Result {
	if k, ok := d.(game.Speak); ok {
		s.keys = append(s.keys, k.Key)
	}
	return s.Rules.Apply(d)
}

// TestSpeakPageAdapter: SpeakPage asks the game to voice the page's title
// and then its voiced paragraphs (por's own playback is audio and not
// observable here, so the keys are checked against the lines).
func TestSpeakPageAdapter(t *testing.T) {
	a, b := engine(t, true, "journal"), engine(t, true, "journal")
	spy := &speakSpy{Rules: b.Rules()}
	g := classic.NewGame(spy, b.Data())
	r := a.Rules().State().Read
	if r == nil {
		t.Fatal("no journal read")
	}
	a.SpeakPage(r.Lines, 0)
	g.SpeakPage(r.Lines, 0)
	if want := strings.ReplaceAll(strings.TrimPrefix(r.Lines[0], "\x00"), " ", "_"); len(spy.keys) == 0 || spy.keys[0] != want {
		t.Errorf("keys %q, want %q first", spy.keys, want)
	}
}

// TestSlotsAdapter: the adapter's SlotsView equals por's, with their times
// and levels, from the party menu's Load and in a game, over the same saves.
func TestSlotsAdapter(t *testing.T) {
	for _, scene := range []string{"", "combat", "shop"} {
		t.Run("scene "+scene, func(t *testing.T) {
			a, b := engine(t, true, scene), engine(t, true, scene)
			b.Cfg.SaveDir = a.Cfg.SaveDir // one set of slots for both
			if scene != "combat" {
				a.Save("one")
				a.Save("two")
			}
			g := classic.NewGame(b.Rules(), b.Data())
			for _, loadOnly := range []bool{false, true} {
				want, got := a.SlotsView(loadOnly), g.SlotsView(loadOnly)
				if !reflect.DeepEqual(got, want) {
					t.Errorf("loadOnly %v: adapter %+v, por %+v", loadOnly, got, want)
				}
				if scene == "" && (want == nil || len(want.Names) == 0 || want.Levels[0] == 0 || want.Times[0] == "") {
					t.Errorf("no slots with times and levels: %+v", want)
				}
			}
		})
	}
}

// TestTitleAdapter: the title calls (picture, lightning, credits, done)
// through the adapter leave the same State as por's own (and por's, at).
func TestTitleAdapter(t *testing.T) {
	dir := filepath.Join(porDir(t), "data")
	if _, err := os.Stat(filepath.Join(dir, "dos", "TITLE.DAX")); err != nil {
		t.Skip("no game files")
	}
	var es [2]*rules.Engine
	for i := range es {
		e, err := rules.New(dir)
		if err != nil {
			t.Fatal(err)
		}
		e.Cfg.SaveDir = t.TempDir()
		e.Start(e.Run)
		es[i] = e
	}
	a, b := es[0], es[1]
	g := classic.NewGame(b.Rules(), b.Data())
	for i, step := range []struct {
		name string
		por  func()
		ad   func()
	}{
		{"start", func() {}, func() {}},
		{"picture", func() { a.TitlePicture(1) }, func() { g.TitlePicture(1) }},
		{"flash", func() { a.TitleFlash(14) }, func() { g.TitleFlash(14) }},
		{"credits", a.TitleCredits, g.TitleCredits},
		{"done", a.TitleDone, g.TitleDone},
	} {
		step.por()
		step.ad()
		a.Apply()
		g.Settle()
		va, vb := normal(a), normal(b)
		at(t, step.name, b)
		if i == 0 && va.Intro == nil {
			t.Fatal("no title up")
		}
		if i == 4 && va.Intro != nil {
			t.Error("title not done")
		}
		for _, n := range diffFields(va, vb) {
			t.Errorf("%s: %s differs:\n  engine  %+v\n  adapter %+v", step.name, n, reflect.ValueOf(va).FieldByName(n), reflect.ValueOf(vb).FieldByName(n))
		}
	}
}
