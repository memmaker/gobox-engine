package engine

import (
	"testing"

	"gobox/adventure"
	"gobox/game"
	"gobox/rules1e"
	"slices"
)

// TestCampDayIsClock: in camp the published Day is the clock's, the one Rest
// Until rests by, even when the sky's colour says day at night.
func TestCampDayIsClock(t *testing.T) {
	e := testEngine(t)
	e.Explore = true
	e.DevAction = func() {
		e.state().Area.Clock[3], e.state().Area.Clock[2], e.state().Area.Clock[1] = 23, 0, 0
		e.world.Sky, e.world.SkyInside = 9, 9 // a day colour
		e.CampMenu()
	}
	e.Init()
	e.Start(e.Play)
	if !e.campUp() {
		t.Fatal("camp not up")
	}
	if day := commonOf(e.build()).Day; day || e.state().IsDay() {
		t.Errorf("23:00 in camp: Day %v, isDay %v", day, e.state().IsDay())
	}
}

// TestCampSaysWhatRestDoes: the camping state says whether a rest would
// heal someone (Heals) and how long the planned spells take (Study), which
// is when gobox offers Fix; none with the party whole and nothing planned.
func TestCampSaysWhatRestDoes(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	c, ok := e.build().(game.Camping)
	if !ok || c.Heals {
		t.Fatalf("a whole party in camp: %T heals %v", e.build(), c.Heals)
	}
	e.state().Party[0].HP--
	if c := e.build().(game.Camping); !c.Heals {
		t.Fatal("a wounded member: Heals is not set, Fix is not offered")
	}
}

// TestFixAtOnce: camp's Fix is done in one go and unseen: the wounded are
// healed and nothing is still resting when it returns, and all it says is
// Refreshed: no sound, no line of the casts and rests it took.
func TestFixAtOnce(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	c := e.state().Party[0]
	c.HP -= 2
	out, err := e.doNow(game.CampRest{Kind: adventure.RestFix})
	if err != nil {
		t.Fatal(err)
	}
	if c.HP != c.MaxHP || e.resting != nil {
		t.Fatalf("after Fix: HP %d of %d, still resting %v", c.HP, c.MaxHP, e.resting != nil)
	}
	said := 0
	for _, ev := range out.Events {
		switch ev := ev.(type) {
		case game.Refreshed:
			said++
			if !ev.Healed {
				t.Errorf("Refreshed %+v: the healing is not named", ev)
			}
		case game.Sounded, game.Narrated, game.Said:
			t.Errorf("Fix shows %T %+v", ev, ev)
		}
	}
	if said != 1 {
		t.Fatalf("Refreshed said %d times in %+v", said, out.Events)
	}
	if _, ok := e.build().(game.Camping); !ok {
		t.Fatalf("after Fix: %T, want camp", e.build())
	}
}

// TestSaveLevel: of a save as Save gives it the game tells the highest
// level of a member of its party (gobox's slot list shows it).
func TestSaveLevel(t *testing.T) {
	e := quickStarted(t)
	e.ext.inGame = true // quickStarted skips the game start; canSave needs a game in play
	want := 0
	for _, c := range e.state().Party {
		for _, l := range c.Levels {
			want = max(want, int(int8(l)))
		}
	}
	b, err := e.Save()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := e.SaveLevel(b); err != nil || got != want || want == 0 {
		t.Fatalf("SaveLevel %d (%v), the party's highest level %d", got, err, want)
	}
	if got, err := e.SaveLevel([]byte("no save")); err == nil || got != 0 {
		t.Fatalf("SaveLevel of no save: %d, %v", got, err)
	}
	var _ game.SaveLevels = e
}

// TestFixMemorizesPlannedSpells: camp's Fix learns the spells the casters
// planned: after it they are memorized and no study is left. The state is
// built on every step of the rest: asking it for the study still to do
// must not set the casters' study going anew (it did: nothing was learned).
func TestFixMemorizesPlannedSpells(t *testing.T) {
	e := quickStarted(t)
	e.CampMenu()
	s := e.state()
	planned := map[int]rules1e.SpellID{}
	for n, c := range s.Party {
		if book, _ := s.SpellLists(c); len(book) > 0 {
			if _, err := e.doNow(game.Plan{Member: n, Spell: book[0].ID, Count: 1}); err != nil {
				t.Fatalf("%s plans a spell: %v", c.Name(), err)
			}
			planned[n] = rules1e.SpellID(book[0].ID)
		}
	}
	if len(planned) == 0 {
		t.Skip("no caster in the quick party")
	}
	if c := e.build().(game.Camping); c.Study == 0 {
		t.Fatal("spells planned, but the state says no study is needed: Fix is not offered")
	}
	out, err := e.doNow(game.CampRest{Kind: adventure.RestFix})
	if err != nil {
		t.Fatal(err)
	}
	for n, sp := range planned {
		c := s.Party[n]
		if !slices.ContainsFunc(c.Memorized[:], func(m rules1e.SpellID) bool { return m.Spell() == sp && !m.Pending() }) {
			t.Errorf("%s has not memorized spell %d: %v", c.Name(), sp, c.Memorized[:6])
		}
	}
	if c := e.build().(game.Camping); c.Study != 0 {
		t.Errorf("after Fix %v of study is still needed", c.Study)
	}
	if !slices.ContainsFunc(out.Events, func(ev game.Event) bool { r, ok := ev.(game.Refreshed); return ok && r.Studied }) {
		t.Errorf("Fix did not say the spells were memorized: %+v", out.Events)
	}
}
