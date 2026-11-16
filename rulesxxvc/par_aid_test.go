package rulesxxvc

import (
	"image"
	"strings"
	"testing"

	"gobox/rules"
)

// parAid is a medic with the given First Aid level beside an ally in status st, and a monster far off; the round is begun.
func parAid(t *testing.T, aid int8, st Status, ally rules.BoardPos) (*Rules, rules.CharID, rules.CharID, rules.Fight, *rules.RNG) {
	t.Helper()
	r := New()
	m := member("medic", 10, 0, aid, 0)
	m.BaseMovement = 12
	w := member("hurt", 10, 0, 0, 0)
	w.Status, w.HP, w.MaxHP = st, 0, 10
	mid, wid := r.Add(m), r.Add(w)
	foe := r.Add(Character{Side: rules.SideMonsters, Enabled: true, Status: StOkay, HP: 5, MaxHP: 5})
	f := rules.Fight{Party: rules.Party{mid, wid}, Actor: mid, Board: rules.Board{Size: image.Pt(9, 9), Units: []rules.BoardUnit{
		{ID: mid, At: rules.BoardPos{X: 4, Y: 4}}, {ID: wid, At: ally}, {ID: foe, At: rules.BoardPos{X: 8, Y: 8}}}}}
	rng := rules.RNG{State: 1}
	r.BeginFight(f, &rng)
	r.BeginRound(f, &rng)
	return r, mid, wid, f, &rng
}

func offersAid(r *Rules, f rules.Fight) bool {
	for _, o := range r.allowedFight(f).Offers {
		if _, ok := o.Action.(rules.Aid); ok {
			return true
		}
	}
	return false
}

// TestAidOffered ports rules TestAid's offer (combat.pas CombatMenu 'Aid': viewchar.pas FindPatients): Aid is
// offered for a dying ally next to a member with First Aid; not without the skill, not for an ally who is not
// dying, not when the ally is not adjacent.
func TestAidOffered(t *testing.T) {
	next, far := rules.BoardPos{X: 5, Y: 4}, rules.BoardPos{X: 7, Y: 4}
	for _, c := range []struct {
		what  string
		skill int8
		st    Status
		at    rules.BoardPos
		want  bool
	}{{"medic and a dying ally beside", 10, StDying, next, true}, {"no skill", 0, StDying, next, false},
		{"ally not dying", 10, StUnconscious, next, false}, {"ally not adjacent", 10, StDying, far, false}} {
		r, _, _, f, _ := parAid(t, c.skill, c.st, c.at)
		if got := offersAid(r, f); got != c.want {
			t.Errorf("%s: offered %v, want %v", c.what, got, c.want)
		}
	}
}

// TestAidBandages ports rules TestAid's bandage (viewchar.pas TreatPatients): the dying ally is unconscious and the
// line is "is bandaged"; the aider's turn is over (CombatMain 'A').
func TestAidBandages(t *testing.T) {
	r, mid, wid, f, rng := parAid(t, 10, StDying, rules.BoardPos{X: 5, Y: 4})
	o, err := r.Resolve(f, rules.Aid{Target: wid}, rng)
	if err != nil {
		t.Fatal(err)
	}
	said := false
	for _, e := range o.Effects {
		if n, ok := e.(rules.Narrated); ok && strings.HasSuffix(n.Text, "hurt is bandaged") {
			said = true
		}
	}
	if r.chars[wid].Status != StUnconscious || !said || r.fight[mid].initiative != 0 || r.fight[mid].movement != 0 {
		t.Errorf("status %v said %v initiative %d movement %d", r.chars[wid].Status, said, r.fight[mid].initiative, r.fight[mid].movement)
	}
}

// TestAidByTheAI: a team member the computer plays treats the patients beside it before anything else
// (combatai.pas AiMoveAttack, not in a solo combat), the others' choice the Aid action.
func TestAidByTheAI(t *testing.T) {
	r, mid, wid, f, rng := parAid(t, 10, StDying, rules.BoardPos{X: 5, Y: 4})
	r.BeginTurn(f, mid, rng)
	if a, ok := r.Decide(f, mid, rng).(rules.Aid); !ok || a.Target != wid {
		t.Errorf("a team member beside a dying ally chose %T", r.Decide(f, mid, rng))
	}
	r, mid, _, f, rng = parAid(t, 10, StDying, rules.BoardPos{X: 5, Y: 4})
	f.Terms.Duel = true
	r.BeginTurn(f, mid, rng)
	if _, ok := r.Decide(f, mid, rng).(rules.Aid); ok {
		t.Error("a solo combat bandages no one")
	}
}

// TestShipAid ports rules TestSpaceAid (shipcbt.pas BuildPlayerMenu 'Aid', ViewChar.TreatPatients): in a ship fight the
// crew member (not the pilot) with First Aid bandages the dying one.
func TestShipAid(t *testing.T) {
	r, sf, pilot, medic := shipFightOf(t, 1) // the pilot has the menu of its own: Aid is the other crew's
	r.chars[medic].Skills.Set(firstAidSkill, 10)
	r.chars[pilot].Status = StDying
	r.space.crew = []rules.CharID{pilot, medic}
	if len(r.shipPatients(r.chars[medic])) != 1 || len(r.shipPatients(r.chars[pilot])) != 0 {
		t.Fatalf("patients of the medic %v, of the pilot %v", r.shipPatients(r.chars[medic]), r.shipPatients(r.chars[pilot]))
	}
	r.space.crewOn, r.space.turn, r.space.stage = true, 0, foeIdle
	if r.crewActor() != medic {
		t.Fatalf("the crew menu is %d's, not the medic's %d", r.crewActor(), medic)
	}
	has := false
	for _, o := range r.Allowed(sf).Offers {
		if _, ok := o.Action.(rules.Aid); ok {
			has = true
		}
	}
	if !has {
		t.Fatal("Aid not offered to the medic")
	}
	rng := rules.RNG{State: 1}
	if _, err := r.Resolve(sf, rules.Aid{}, &rng); err != nil || r.chars[pilot].Status != StUnconscious {
		t.Errorf("status %v, error %v", r.chars[pilot].Status, err)
	}
}
