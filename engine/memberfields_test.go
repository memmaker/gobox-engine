package engine

import (
	"testing"

	"gobox/game"
	"gobox/rules1e"
)

// podMember is pod's engine with a party begun and its selected member's record.
func podMember(t *testing.T) (*Engine, *rules1e.Character) {
	t.Helper()
	e := advStarted(t, advEnvPOD)
	for _, a := range []game.Action{game.QuickStart{}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	return e, e.member(e.selected())
}

// scriptRead is what a script's read of pod's field name, element i, gives.
func scriptRead(t *testing.T, e *Engine, name string, i int) int64 {
	t.Helper()
	return e.host.Get(e.field(name).Plus(i))
}

// TestPodMemberFields: pod's member fields read the selected member (SpecialReadZ1 of pod-re eclmenu.pas
// reads the record's bytes at the offset) and the ones SpecialWriteZ1 stores (hpCur) change it.
func TestPodMemberFields(t *testing.T) {
	e, c := podMember(t)
	c.HP, c.MaxHP, c.Sex, c.Class = 11, 40, 1, rules1e.ClsFighterThief
	c.Stats, c.Base = [6]rules1e.Score{15, 12, 11, 10, 9, 8}, [6]rules1e.Score{17, 13, 12, 11, 10, 9}
	c.Levels[rules1e.ClsMage], c.OldLevels[rules1e.ClsRanger] = 7, 5
	c.Thief = [8]rules1e.Percent{1, 2, 3, 4, 5, 6, 7, 8}
	reads := []struct {
		name string
		i    int
		want int64
	}{
		{"SelHpCur", 0, 11}, {"SelHpMax", 0, 40}, {"SelGender", 0, 1}, {"SelCharClass", 0, int64(rules1e.ClsFighterThief)},
		{"SelStats", 0, 17}, {"SelStats", 1, 15}, {"SelStats", 2, 13}, {"SelStats", 3, 12}, {"SelStats", 6, 11}, {"SelStats", 7, 10}, {"SelStats", 8, 10}, // (base, cur) by C_Stat
		{"SelCon", 0, 9}, {"SelCharisma", 0, 9}, {"SelCharisma", 1, 8},
		{"SelClassLevels", int(rules1e.ClsMage), 7}, {"SelClassLevels", int(rules1e.ClsRanger), 0},
		{"SelOldLevels", int(rules1e.ClsRanger), 5}, {"SelOldLevels", int(rules1e.ClsMage), 0},
		{"SelPickPockets", 0, 1}, {"SelOpenLocks", 0, 2}, {"SelFindTraps", 0, 3}, {"SelMoveSilently", 0, 4},
		{"SelHideShadows", 0, 5}, {"SelHearNoise", 0, 6}, {"SelThiefSkill6", 0, 7}, {"SelReadLanguages", 0, 8},
	}
	for _, r := range reads {
		if got := scriptRead(t, e, r.name, r.i); got != r.want {
			t.Errorf("%s_%d reads %d, want %d", r.name, r.i, got, r.want)
		}
	}
	e.host.Set(e.field("SelHpCur"), 23) // SpecialWriteZ1 $1FC: hpCur := Byte(value)
	if c.HP != 23 {
		t.Errorf("a write of SelHpCur left hit points %d", c.HP)
	}
}

// TestPodHealWrite: the heal of area 39 (and 75), SelHpCur = SelHpMax, raises the selected member to its maximum
// and no one else.
func TestPodHealWrite(t *testing.T) {
	e, c := podMember(t)
	c.MaxHP, c.HP = 40, 3
	var others []*rules1e.Character
	for _, id := range e.partyIDs() {
		if m := e.member(id); m != c {
			m.MaxHP, m.HP = 30, 2
			others = append(others, m)
		}
	}
	hp, hpMax := e.field("SelHpCur"), e.field("SelHpMax")
	e.host.Set(hp, e.host.Get(hpMax)) // what the VM does for the assignment: getVar, setVar
	if c.HP != 40 {
		t.Errorf("healed to %d, want 40", c.HP)
	}
	for _, m := range others {
		if m.HP != 2 {
			t.Errorf("another member healed to %d", m.HP)
		}
	}
}

// TestMemberFieldsOfOtherGames: the bindings the other Pool games' layouts name read the member as their
// SpecialReadZ1 does (cotab-re, sotsb-re eclmenu.pas); the fields are the engine's by name.
func TestMemberFieldsOfOtherGames(t *testing.T) {
	e, c := podMember(t)
	c.Stats = [6]rules1e.Score{0, 14, 0, 0, 0, 16}
	c.Race, c.Levels[rules1e.ClsMage], c.OldLevels[rules1e.ClsMage] = rules1e.RaceHuman, 3, 2
	c.Levels[rules1e.ClsRanger], c.OldLevels[rules1e.ClsRanger] = 2, 4
	c.ScriptFlag = 3
	c.MulticlassLevel = 0
	read := func(name string) int64 { v, _ := e.engineFields[name].get(0); return v }
	if got := read("SelInt"); got != 14 { // cotab $15
		t.Errorf("SelInt %d", got)
	}
	if got := read("SelCha"); got != 16 { // sotsb $19
		t.Errorf("SelCha %d", got)
	}
	if got := read("SelFlag1"); got != 3 { // the canonical byte; the cotab encoding keeps bit 0 (TestCodecCotabSelFlag1)
		t.Errorf("SelFlag1 %d", got)
	}
	// $C9/$D0: the level the game's rule counts, the class left behind's while the earlier class is active
	// (pod's: the higher of the two; effectiveLevel's own test has the others)
	if got := read("SelMageLevel"); got != 3 {
		t.Errorf("SelMageLevel %d, want 3", got)
	}
	if got := read("SelRangerLevel"); got != 4 {
		t.Errorf("SelRangerLevel %d, want 4", got)
	}
}

// TestStaleMemberCellsLoad: a save written while a member field was unbound carries its cell as a game field;
// it loads (a bound field no longer saves its cell, the stale one is carried unread) and the script's read is the member's.
func TestStaleMemberCellsLoad(t *testing.T) {
	e, c := podMember(t)
	c.HP = 9
	e.vars["SelHpCur"] = []int64{99}
	e.vars["SelGender"] = []int64{1}
	if err := e.saveTo("stale"); err != nil {
		t.Fatal(err)
	}
	if !e.loadFrom("stale") {
		t.Fatal("a save with cells of now-bound member fields did not load")
	}
	if got := scriptRead(t, e, "SelHpCur", 0); got != int64(e.member(e.selected()).HP) {
		t.Errorf("SelHpCur reads %d, the member has %d", got, e.member(e.selected()).HP)
	}
}
