package engine

import (
	"strconv"
	"strings"
	"testing"

	"gobox/rules"
)

// br1Run runs body as a block of the test's own on BR1 begun with its one member.
func br1Run(t *testing.T, block int, body string) *Engine {
	t.Helper()
	e := br1Begun(t)
	br1OwnBlock(t, e, block, body)
	return e
}

// TestBR1WallDistance: WallDistance(dst, dir) writes the squares to a wall that way, the dialect's five at most
// (eclmenu.pas MaxMonsterDist), and the Pool's two when the dialect says so (FarSight).
func TestBR1WallDistance(t *testing.T) {
	e := br1Run(t, 201, "    AreaKind(0, 255)\n    WallDistance(Word4C31, 0)\n    WallDistance(Word4C75, 1)")
	for _, c := range []struct {
		cell string
		dir  int
	}{{"Word4C31", 0}, {"Word4C75", 1}} {
		want := 0
		x, y := e.posX, e.posY
		for want < freeStepsFar && e.WallAt(x, y, c.dir*2) == 0 {
			want++
			x, y = x+[]int{0, 1, 0, -1}[c.dir], y+[]int{-1, 0, 1, 0}[c.dir]
		}
		if got := br1Cell(t, e, c.cell); got != int64(want) {
			t.Errorf("%s: dir %d gave %d, want %d", c.cell, c.dir, got, want)
		}
	}
	if !e.eclVersion.Behaviour.FarSight {
		t.Error("BR counts five squares")
	}
}

// TestBR1FindItem: FindItem(type) sets the compare flags the next if reads: equal when someone carries an item of
// the type (eclvm.pas EclFindItem; the Pool's FindItem), not equal otherwise.
func TestBR1FindItem(t *testing.T) {
	e := br1Begun(t)
	have, lack := 0, 0
	for n := 1; n <= 38 && (have == 0 || lack == 0); n++ {
		id, err := e.itemID(n)
		if err != nil {
			t.Fatal(err)
		}
		if e.ruleset().Check(rules.HasItem{Whom: e.partyIDs(), Item: id}, nil).Yes {
			have = n
		} else {
			lack = n
		}
	}
	if have == 0 || lack == 0 {
		t.Fatalf("the member carries types %d, lacks %d", have, lack)
	}
	body := "    FindItem(" + strconv.Itoa(have) + ")\n    if (==) { Word4C31 = 1 }\n    FindItem(" + strconv.Itoa(lack) + ")\n    if (==) { Word4C75 = 1 }"
	br1OwnBlock(t, e, 201, body)
	if br1Cell(t, e, "Word4C31") != 1 || br1Cell(t, e, "Word4C75") != 0 {
		t.Errorf("have %d: %d, lack %d: %d", have, br1Cell(t, e, "Word4C31"), lack, br1Cell(t, e, "Word4C75"))
	}
}

// TestBR1Skill: Skill(skill, whole party, scale) writes who rolled and the grade into its second and third
// operands; the skill's level scaled by the difficulty decides (very easy x4 of 60 always makes it well, none
// always fails badly). SkillShow says the ruleset's report.
func TestBR1Skill(t *testing.T) {
	const firstAid = 13
	e := br1Begun(t)
	id := e.partyIDs()[0]
	set := func(level rules.Number) {
		if err := e.ruleset().SetFact(id, rules.SkillFact{Skill: "first-aid"}, level); err != nil {
			t.Fatal(err)
		}
	}
	set(60)
	br1OwnBlock(t, e, 201, "    Word4C31 = 0\n    Word4C75 = 0\n    Skill("+strconv.Itoa(firstAid)+", Word4C31, Word4C75)")
	if br1Cell(t, e, "Word4C31") != 0 || br1Cell(t, e, "Word4C75") != int64(rules.MadeWell) {
		t.Errorf("trained: who %d grade %d", br1Cell(t, e, "Word4C31"), br1Cell(t, e, "Word4C75"))
	}
	set(0)
	br1OwnBlock(t, e, 202, "    Word4C31 = 1\n    Word4C75 = 0\n    Skill("+strconv.Itoa(firstAid)+", Word4C31, Word4C75)")
	if br1Cell(t, e, "Word4C31") != 0 || br1Cell(t, e, "Word4C75") != int64(rules.FailedBadly) {
		t.Errorf("untrained, whole party: who %d grade %d", br1Cell(t, e, "Word4C31"), br1Cell(t, e, "Word4C75"))
	}
}

// TestBR1SkillShow: SkillShow says "<name> succeeds at <skill>" or "party fails at <skill>" (EclSkillCheck's message).
func TestBR1SkillShow(t *testing.T) {
	const firstAid = 13
	e := br1Begun(t)
	id := e.partyIDs()[0]
	name := e.ruleset().Sheet(id).Name
	if err := e.ruleset().SetFact(id, rules.SkillFact{Skill: "first-aid"}, rules.Number(60)); err != nil {
		t.Fatal(err)
	}
	br1OwnBlock(t, e, 201, "    Word4C31 = 0\n    Word4C75 = 0\n    SkillShow("+strconv.Itoa(firstAid)+", Word4C31, Word4C75)")
	if got := paneText(commonOf(e.build())); !strings.Contains(got, name+" succeeds at First Aid") {
		t.Errorf("text %q", got)
	}
	if err := e.ruleset().SetFact(id, rules.SkillFact{Skill: "first-aid"}, rules.Number(0)); err != nil {
		t.Fatal(err)
	}
	br1OwnBlock(t, e, 202, "    Word4C31 = 1\n    Word4C75 = 0\n    SkillShow("+strconv.Itoa(firstAid)+", Word4C31, Word4C75)")
	if got := paneText(commonOf(e.build())); !strings.Contains(got, "party fails at First Aid") {
		t.Errorf("text %q", got)
	}
}

// TestBR1NPCJoin: AddNPC(monster, morale), BR's NPCJoin: the monster joins the party, is the selected member,
// and its control byte is the morale halved with the non-player flag (eclvm.pas EclAddNpc).
func TestBR1NPCJoin(t *testing.T) {
	e := br1Begun(t)
	before := len(e.partyIDs())
	br1OwnBlock(t, e, 201, "    AddNPC(106, 100)")
	ids := e.partyIDs()
	if len(ids) != before+1 || e.selected() != ids[before] {
		t.Fatalf("party %d -> %d, selected %v", before, len(ids), e.selected())
	}
	if c, err := e.ruleset().Fact(ids[before], rules.FactControl); err != nil || c != rules.Number(100/2|0x80) {
		t.Errorf("control %v %v", c, err)
	}
	if e.ruleset().Sheet(ids[before]).Name == "" {
		t.Error("no monster loaded")
	}
}

// br1CarriedType is an item type (1..38) the party carries, and one it does not.
func br1CarriedType(t *testing.T, e *Engine) (have, lack int) {
	t.Helper()
	for n := 1; n <= 38 && (have == 0 || lack == 0); n++ {
		id, err := e.itemID(n)
		if err != nil {
			t.Fatal(err)
		}
		if e.ruleset().Check(rules.HasItem{Whom: e.partyIDs(), Item: id}, nil).Yes {
			have = n
		} else {
			lack = n
		}
	}
	if have == 0 || lack == 0 {
		t.Fatalf("the party carries type %d, lacks %d", have, lack)
	}
	return have, lack
}

// TestBR1TakeItem: TakeItem(whole party, type) takes the items of the type from the selected member (everyone
// when the first operand is not 0); the +128 form takes one (eclvm.pas EclDestroyItems).
func TestBR1TakeItem(t *testing.T) {
	e := br1Begun(t)
	have, _ := br1CarriedType(t, e)
	id, _ := e.itemID(have)
	carries := func() bool { return e.ruleset().Check(rules.HasItem{Whom: e.partyIDs(), Item: id}, nil).Yes }
	n := int(e.ruleset().Check(rules.CarryCount{Who: e.partyIDs()[0]}, nil).Value)
	br1OwnBlock(t, e, 201, "    TakeItem(1, "+strconv.Itoa(have)+")")
	if carries() {
		t.Error("the items of the type are still carried")
	}
	if m := int(e.ruleset().Check(rules.CarryCount{Who: e.partyIDs()[0]}, nil).Value); m >= n {
		t.Errorf("carried %d, was %d", m, n)
	}
}

// TestBR1Stash: Stash(0) puts every member's items on its second list, Stash(1) brings them back
// (eclvm.pas EclTransferCharacterLists).
func TestBR1Stash(t *testing.T) {
	e := br1Begun(t)
	who := e.partyIDs()[0]
	count := func() int { return int(e.ruleset().Check(rules.CarryCount{Who: who}, nil).Value) }
	n := count()
	if n == 0 {
		t.Fatal("the member carries nothing")
	}
	br1OwnBlock(t, e, 201, "    Stash(0)")
	if count() != 0 {
		t.Errorf("stashed: %d carried", count())
	}
	br1OwnBlock(t, e, 202, "    Stash(1)")
	if count() != n {
		t.Errorf("back: %d carried, was %d", count(), n)
	}
}

// TestBR1SkillDamage: SkillDamage(skill, whole party, difficulty, dice, sides, bonus): a member who fails the
// roll takes the damage and the message says so (EclInflictDamage, the number unpadded); one who makes it does not.
func TestBR1SkillDamage(t *testing.T) {
	const firstAid = 13
	e := br1Begun(t)
	id := e.partyIDs()[0]
	hp := e.ruleset().Sheet(id).HP
	if hp <= 5 {
		t.Fatalf("hit points %d", hp)
	}
	set := func(level rules.Number) {
		if err := e.ruleset().SetFact(id, rules.SkillFact{Skill: "first-aid"}, level); err != nil {
			t.Fatal(err)
		}
	}
	set(60)
	br1OwnBlock(t, e, 201, "    SkillDamage("+strconv.Itoa(firstAid)+", 1, 0, 1, 1, 4)") // very easy x4: made
	if got := e.ruleset().Sheet(id).HP; got != hp {
		t.Errorf("made the roll, hit points %d -> %d", hp, got)
	}
	set(0)
	br1OwnBlock(t, e, 202, "    SkillDamage("+strconv.Itoa(firstAid)+", 1, 0, 1, 1, 4)\n    PressKey()")
	if got := e.ruleset().Sheet(id).HP; got != hp-5 {
		t.Errorf("failed the roll, hit points %d -> %d, want %d", hp, got, hp-5)
	}
	if got := paneText(commonOf(e.build())); !strings.Contains(got, " is hit FOR 5 points of Damage.") {
		t.Errorf("text %q", got)
	}
}
