package pool

import (
	"fmt"
	"gobox/core"
	"image"
	"reflect"
	"rules1e"
	"slices"
	"strings"
	"testing"

	"gobox/game"
	"gobox/game/play"
)

// TestNoSilentDo (BUG-CLASSES 31): in each state, every action the state
// allows (Allowed's Plain and Aimed, the ask's replies), done on the state
// as it stands, either changes the state (reflect.DeepEqual), tells a fact
// (an event), or is refused. Two State calls in a row are DeepEqual first,
// so a difference is the action's; nothing is offered twice (class 34).
func TestNoSilentDo(t *testing.T) {
	quick := quickStarted
	scene := func(name string) func(t *testing.T) *Engine {
		return func(t *testing.T) *Engine {
			e := testEngine(t)
			e.Explore = true
			sc := e.DevScene(strings.TrimSuffix(name, " caster"))
			e.DevAction = sc
			if strings.HasSuffix(name, " caster") { // Magic Missile memorized: the turn offers Cast
				e.DevAction = func() {
					sc()
					for _, c := range e.state().Party {
						c.Memorized[0] = rules1e.SpMagicMissile
					}
				}
				name = strings.TrimSuffix(name, " caster")
			}
			seed := uint32(1)
			e.FixedSeed = &seed
			e.Init()
			e.Start(e.Play)
			for range 400 { // up to the first party turn or a still view
				if e.cs.turn != nil && e.cs.actor != core.NoChar || name != "combat" && e.exploring() {
					break
				}
				if e.waiting(PendFight) {
					e.Step()
				} else {
					press(e)
				}
			}
			return e
		}
	}
	states := map[string]func(t *testing.T) *Engine{
		"party menu": quick,
		"creation": func(t *testing.T) *Engine {
			e := quick(t)
			if _, err := e.Do(play.NewCharacter{}); err != nil {
				t.Fatal(err)
			}
			return e
		},
		"combat":        scene("combat"),
		"combat caster": scene("combat caster"),
		"overland":      scene("overland"),
		"camp": func(t *testing.T) *Engine {
			e := scene("overland")(t)
			if _, err := e.Do(play.MakeCamp{}); err != nil {
				t.Fatal(err)
			}
			return e
		},
	}
	noSilentDo(t, states)
}

// TestNoSilentDoCOTAB: the same on cotab's imported folder (it skips without
// one), in the states reached without POR's dev scenes: the party menu,
// creation, and exploring after Quick Start and Begin.
func TestNoSilentDoCOTAB(t *testing.T) {
	quick := func(t *testing.T) *Engine { return quickStartedOn(t, advEngine(t, advEnvCOTAB)) }
	noSilentDo(t, map[string]func(t *testing.T) *Engine{
		"party menu": quick,
		"creation": func(t *testing.T) *Engine {
			e := quick(t)
			if _, err := e.Do(play.NewCharacter{}); err != nil {
				t.Fatal(err)
			}
			return e
		},
		"explore": func(t *testing.T) *Engine {
			e := advStarted(t, advEnvCOTAB)
			for _, a := range []play.Action{play.QuickStart{}, play.Begin{}} {
				if _, err := e.Do(a); err != nil {
					t.Fatalf("%T: %v", a, err)
				}
			}
			return e
		},
	})
}

// noSilentDo is TestNoSilentDo's check over the states each opener opens.
func noSilentDo(t *testing.T, states map[string]func(t *testing.T) *Engine) {
	for name, open := range states {
		e := open(t)
		s := e.State()
		for range 20 { // BUG-CLASSES 33: a field built from a map differs on some call
			if !reflect.DeepEqual(s, e.State()) {
				t.Fatalf("%s: two State calls differ", name)
			}
		}
		c := commonOf(s)
		seen := map[string]bool{}
		var acts []play.Action
		for _, a := range c.Allowed.Plain {
			k := fmt.Sprintf("%T%+v", a, a)
			if seen[k] {
				t.Errorf("%s: %s offered twice", name, k)
			}
			seen[k] = true
			acts = append(acts, a)
		}
		for _, a := range c.Allowed.Aimed { // at the first square offered
			if len(a.Squares) == 0 {
				continue
			}
			at := []image.Point{a.Squares[0]}
			switch x := a.Action.(type) {
			case play.Cast:
				x.Ats = at
				acts = append(acts, x)
			case play.Shoot:
				x.Ats = at
				acts = append(acts, x)
			case play.UseItem:
				x.Ats = at
				acts = append(acts, x)
			}
		}
		if c.Ask != nil {
			for n := range max(len(c.Ask.Replies), 1) {
				acts = append(acts, play.Reply{N: n})
			}
		}
		for _, a := range acts {
			switch x := a.(type) {
			case play.Quit: // exits
				continue
			case play.SelectMember: // the selected one again
				if x.Member == commonOf(s).Selected {
					continue
				}
			case play.SetChoice: // the field's own pick again: the list's no-op
				if as, ok := s.(play.Assembling); ok && as.Draft != nil && slices.ContainsFunc(as.Draft.Fields, func(f game.Field) bool { return f.Name == x.Field && f.Pick == x.N }) {
					continue
				}
			}
			before := e.State()
			out, err, panicked := doCatch(e, a)
			switch {
			case panicked != nil:
				t.Errorf("%s: %T%+v panics: %v", name, a, a, panicked)
			case err == nil && len(out.Events) == 0 && reflect.DeepEqual(before, e.State()):
				t.Errorf("%s: %T%+v taken with no effect", name, a, a)
			}
			if panicked != nil || !reflect.DeepEqual(before, e.State()) {
				e = open(t) // fresh for the next one
			}
		}
	}
}

func doCatch(e *Engine, a play.Action) (out play.Outcome, err error, panicked any) {
	defer func() { panicked = recover() }()
	out, err = e.Do(a)
	return out, err, nil
}

// TestGuardsRefuse (BUG-CLASSES 85): an action a guard turns away is an
// error, not a nil: Modify without a party, KeepCharacter without a
// creation, a second NewCharacter; an empty note is refused.
func TestGuardsRefuse(t *testing.T) { guardsRefuse(t, testEngine) }

// TestGuardsRefuseCOTAB: the same on cotab's imported folder.
func TestGuardsRefuseCOTAB(t *testing.T) {
	guardsRefuse(t, func(t *testing.T) *Engine { return advEngine(t, advEnvCOTAB) })
}

// guardsRefuse is TestGuardsRefuse on the engines base makes.
func guardsRefuse(t *testing.T, base func(*testing.T) *Engine) {
	e := quickStartedOn(t, base(t))
	for _, a := range []play.Action{play.KeepCharacter{}, play.AddNote{}} {
		if _, err := e.Do(a); err == nil {
			t.Errorf("%T taken", a)
		}
	}
	if _, err := e.Do(play.NewCharacter{}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Do(play.NewCharacter{}); err == nil {
		t.Error("second NewCharacter taken")
	}
	e = base(t)
	if _, err := e.Do(play.Modify{}); err == nil {
		t.Error("Modify with no party taken")
	}
}
