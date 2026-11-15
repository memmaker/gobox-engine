package engine

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"gobox/game"
	"gobox/rules"
	"gobox/rules1e"
)

// setScriptWord writes a world word the way a script does: through its layout field name.
func setScriptWord(t *testing.T, e *Engine, name string, v int) {
	t.Helper()
	e.loadScript(14)
	r, ok := e.layoutField(name)
	if !ok {
		t.Fatalf("no layout field %s", name)
	}
	e.host.Set(r, int64(v))
}

// terrainGrid is the overland's published terrain, as one string of hex bytes.
func terrainGrid(e *Engine) string {
	var b strings.Builder
	for _, v := range e.build().(game.Exploring).Map.Automap.Terrain {
		fmt.Fprintf(&b, "%02x", v)
	}
	return b.String()
}

// diffGrid lists the squares where b differs from a, "x,y:a>b".
func diffGrid(a, b string) string {
	var out []string
	for i := 0; i+2 <= len(a); i += 2 {
		if a[i:i+2] != b[i:i+2] {
			n := i / 2
			out = append(out, fmt.Sprintf("%d,%d:%s>%s", n%44, n/44, a[i:i+2], b[i:i+2]))
		}
	}
	return strings.Join(out, " ")
}

// TestPlaceWordsMarks: the overland's terrain shows the area map's marks by the mark words
// (expectations recorded before the words moved to the engine).
func TestPlaceWordsMarks(t *testing.T) {
	cases := []struct {
		words map[string]int
		want  string
	}{
		{map[string]int{"Mark3": 0xFF, "Boss": 0xFE}, marksA},
		{map[string]int{"MarkFlags57": 3, "Mark4Off": 1}, marksB},
	}
	for _, c := range cases {
		e := testEngine(t)
		e.Explore = true
		e.DevAction = e.DevScene("overland")
		e.Init()
		e.Start(e.Play)
		if !playOn(e, 50, func(s game.State) bool { x, ok := s.(game.Exploring); return ok && x.Ask == nil && x.Map.Walk != nil }) {
			t.Fatal("not exploring the overland")
		}
		base := terrainGrid(e)
		for name, v := range c.words {
			setScriptWord(t, e, name, v)
		}
		if got := diffGrid(base, terrainGrid(e)); got != c.want {
			t.Errorf("%v: changed squares %q, want %q", c.words, got, c.want)
		}
	}
}

// TestPlaceWordsDay: the published Day by the sky words, the 3D flag and the hour, in camp and out.
func TestPlaceWordsDay(t *testing.T) {
	e := quickStarted(t)
	var got []string
	for _, in3D := range []int{0, 1} {
		for _, sky := range []int{0, 8, 9} {
			for _, inside := range []int{0, 8, 9} {
				for _, hour := range []uint16{12, 23} {
					setScriptWord(t, e, "In3D", in3D)
					setScriptWord(t, e, "Sky", sky)
					setScriptWord(t, e, "SkyInside", inside)
					e.state().Area.Clock[3], e.state().Area.Clock[2], e.state().Area.Clock[1] = hour, 0, 0
					got = append(got, fmt.Sprintf("%d%d%d%d:%v%v", in3D, sky, inside, hour, e.flagDay(false), e.flagDay(true)))
				}
			}
		}
	}
	e.CampMenu()
	got = append(got, fmt.Sprint("camp:", e.flagDay(false), e.flagDay(true)))
	if s := strings.Join(got, " "); s != days {
		t.Errorf("days %s\nwant %s", s, days)
	}
}

// TestPlaceWordsNoMagic: Cast in camp is refused where the area forbids magic, offered where it does not.
func TestPlaceWordsNoMagic(t *testing.T) {
	e := quickStarted(t)
	c := e.member(e.selected())
	c.Memorized[0] = rules1e.SpCureLightWounds
	e.CampMenu()
	cast := func() error {
		return e.runCall(pendingCall{kind: callPower, member: -1, power: game.PowerCast, spell: 0, target: -1})
	}
	if r, ok := cast().(*game.Refusal); !ok || r.Ask == nil {
		t.Fatal("Cast without NoMagic: no spell pick")
	}
	setScriptWord(t, e, "NoMagic", 1)
	if err := cast(); !errors.Is(err, rules.ErrNotNow) {
		t.Errorf("Cast with NoMagic: %v, want not now", err)
	}
	setScriptWord(t, e, "NoMagic", 0)
	if r, ok := cast().(*game.Refusal); !ok || r.Ask == nil {
		t.Error("Cast after NoMagic is cleared: no spell pick")
	}
}

// TestPlaceWordsHall: the hall trains a member whose class its word names (the dev scene's
// word is all classes), none when it names none, anyone in debug; training is free in debug.
func TestPlaceWordsHall(t *testing.T) {
	card := func(e *Engine) rules.TrainCard {
		return e.fact(e.selected(), rules.TrainFact{Hall: e.hallTerms()}).(rules.TrainCard)
	}
	e := testEngine(t)
	e.Explore = true
	e.DevAction = e.DevScene("hall")
	e.Init()
	e.Start(e.Play)
	e.Advance()
	if c := card(e); !c.Rises || c.Why != "" {
		t.Errorf("the dev hall: %+v", c)
	}
	setScriptWord(t, e, "PartyExp", 0)
	if c := card(e); c.Rises || c.Why != "We don't train that class here" {
		t.Errorf("a hall of no class: %+v", c)
	}
	e.state().Party[0].Exp, e.state().Purse.Gold = 0, 0
	if c := card(e); c.Rises {
		t.Errorf("no experience, no gold: %+v", c)
	}
	e.FreeTraining = true
	e.partyMenu(true, true)
	if c := card(e); !c.Rises || c.Why != "" {
		t.Errorf("debug: %+v", c)
	}
	gold := e.state().Purse.Gold
	if err := e.openTraining(0); err != nil {
		t.Fatal(err)
	}
	if err := e.train(); err != nil || e.state().Purse.Gold != gold {
		t.Errorf("free training: %v, gold %d (was %d)", err, e.state().Purse.Gold, gold)
	}
}

// Recorded on the code before the words moved to the engine.
const (
	marksA = "19,16:e7>f6 25,27:fc>ff 12,31:01>f8"
	marksB = "37,8:92>fe 25,11:00>fa 32,15:a4>fb"
	days   = "00012:truetrue 00023:falsefalse 00812:truetrue 00823:falsefalse 00912:truetrue 00923:falsefalse 08012:truetrue 08023:falsefalse 08812:truetrue 08823:falsefalse 08912:truetrue 08923:falsefalse 09012:truetrue 09023:falsefalse 09812:truetrue 09823:falsefalse 09912:truetrue 09923:falsefalse 10012:truetrue 10023:falsefalse 10812:truefalse 10823:falsefalse 10912:truetrue 10923:falsetrue 18012:falsetrue 18023:falsefalse 18812:falsefalse 18823:falsefalse 18912:falsetrue 18923:falsetrue 19012:truetrue 19023:truefalse 19812:truefalse 19823:truefalse 19912:truetrue 19923:truetrue camp:false false"
)
