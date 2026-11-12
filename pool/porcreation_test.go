package pool

import (
	"errors"
	"reflect"
	"rules1e"
	"slices"
	"testing"

	"gobox/char"
	"gobox/core"
	flow "gobox/creation"
)

// TestPorCreation: the gobox flow over porCreation draws the dice
// CreateCharacter draws and rebuilds the very character the old path holds
// (charOf(draftOf(c)) == c), through rekinds, rolls and typed scores.
func TestPorCreation(t *testing.T) {
	e := testEngine(t)
	p := e.r1e().Creation()
	same := func(what string, want *rules1e.Character, d core.Draft) {
		t.Helper()
		got := p.CharOf(d)
		if !reflect.DeepEqual(char.SheetBytes(got.Sheet), char.SheetBytes(want.Sheet)) || !reflect.DeepEqual(got.Powers, want.Powers) {
			t.Fatalf("%s: rebuilt record differs\n got %+v\nwant %+v", what, got, want)
		}
	}
	seed := uint32(12345)
	e.state().RandSeed = seed
	old := e.state().Fresh(rules1e.Kind{Race: rules1e.RaceHuman, Class: rules1e.ClsFighter}, nil)
	e.state().RollCharacter(old)
	after := e.state().RandSeed

	r := &core.RNG{State: seed}
	d, err := flow.New(p, r)
	if err != nil {
		t.Fatal(err)
	}
	if r.State != after {
		t.Fatalf("New drew differently: seed %d, want %d", r.State, after)
	}
	same("new", old, d.D)
	if d.D.Head == "" || d.D.Body == "" || d.D.Icon.Head == "" || d.D.Icon.Body == "" {
		t.Fatalf("POR's own pictures have no PicID: %q %q %+v", d.D.Head, d.D.Body, d.D.Icon)
	}

	cr := &rules1e.SheetChar{C: old, K: rules1e.Kind{Race: rules1e.RaceHuman, Class: rules1e.ClsFighter}}
	e.state().RandSeed = r.State
	e.state().SetSheetField(cr, rules1e.FieldClass, slices.Index(rules1e.PickClasses(), rules1e.ClsMage))
	d.SetField(rules1e.FieldClass, slices.Index(rules1e.PickClasses(), rules1e.ClsMage))
	same("mage", cr.C, d.D)
	if r.State != e.state().RandSeed {
		t.Fatal("SetField drew differently")
	}
	if err := d.SetScore(core.Intelligence, "17"); err != nil {
		t.Fatal(err)
	}
	if d.D.Sheet.Scores[core.Intelligence] != 17 {
		t.Fatalf("int %d", d.D.Sheet.Scores[core.Intelligence])
	}
	d.Roll(core.RollFull)
	same("rolled", p.CharOf(d.D), p.DraftOf(p.CharOf(d.D), core.NoChar))
}

// TestSetScoreBaseline pins c70cbee's setStatTyped and rollMinimal: a typed
// score out of the race and class range is clamped, never refused; "18/5"
// is 18/05; text that is no score keeps the score; and the fighters-only
// strength message.
func TestSetScoreBaseline(t *testing.T) {
	e := testEngine(t)
	p := e.r1e().Creation()
	d, err := flow.New(p, &core.RNG{State: 7})
	if err != nil {
		t.Fatal(err)
	}
	c := p.CharOf(d.D)
	if err := d.SetScore(core.Constitution, "25"); err != nil || d.D.Sheet.Scores[core.Constitution] != int32(rules1e.LimitStat(c, rules1e.StatCon, 25)) {
		t.Errorf("25 not clamped: %v %d", err, d.D.Sheet.Scores[core.Constitution])
	}
	if err := d.SetScore(core.Strength, "18/5"); err != nil || d.D.Sheet.Scores[core.Strength] != 18 || d.D.Sheet.Scores[core.StrengthPercent] != 5 {
		t.Errorf("18/5: %v %v", err, d.D.Sheet.Scores)
	}
	if !errors.Is(d.SetScore(core.Strength, "x"), flow.ErrScore) || d.D.Sheet.Scores[core.StrengthPercent] != 5 {
		t.Error("no score changed it")
	}
	d.SetField(rules1e.FieldClass, slices.Index(rules1e.PickClasses(), rules1e.ClsMage))
	if err := d.SetScore(core.Strength, "18/50"); err != nil || d.D.Sheet.Scores[core.StrengthPercent] != 0 {
		t.Errorf("mage 18/50: %v %v", err, d.D.Sheet.Scores)
	}
	if got := rules1e.MinimumError(0, 1850, 1800, 0).Error(); got != "STR 18/xx is for fighters only, max 18" {
		t.Errorf("message %q", got)
	}
}
