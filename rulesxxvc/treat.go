package rulesxxvc

import (
	"slices"
	"strconv"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// postcbt.pas TreatInjuries (BR1 and BR2 alike): the medics treat poisoning (an antidote is offered when none
// succeeds), then every hurt member's wounds in four passes of falling urgency.
const (
	healPassFirstAid      = 0 // first aid: 1d6, up to 6 hit points, repeats over medics
	healPassCritical      = 1 // treat critical: 3d8+3, hard (x1/4) for a medic on itself
	healPassSerious       = 2 // treat serious: 2d8+1
	healPassLight         = 3 // treat light: 1d8
	healPasses            = 4
	firstAidHealCap       = 6                         // first aid never takes a member past 6 hit points
	unkillableMinHP       = 45                        // C_Effect.UNKILLABLE: a member holding it is put back on its feet with 45
	antidoteKind          = xxvc.ItemIDPoisonAntidote // postcbt.pas ANTIDOTE_TYPE 31
	treatPoisoningHeading = "  Treating Poisoning"
	treatWoundsHeading    = "  Treating Wounds"
)

// treatRun is a treatment under way: it stands at the poison pass's question until the antidote is answered.
type treatRun struct {
	rng     *rules.RNG
	party   []rules.CharID
	next    int          // the poison pass's patient (an index of party)
	asked   rules.CharID // the patient the antidote question is about, NoChar for none
	owner   rules.CharID // who carries the antidote offered
	item    int          // its place in owner's items
	out     []rules.Effect
	headed  bool // "Treating Poisoning" shown
	injured bool // some member was poisoned or hurt
	medics  bool // some medic could try
	treated bool // something was treated
}

func (t *treatRun) say(text string) { t.out = append(t.out, rules.Narrated{Text: text}) }

// skillMade is SelectSkillCharacter(n, h, False, c) >= 2 over one character.
func (r *Rules) skillMade(rng *rules.RNG, c *Character, n SkillNo, h rules.Hardness) bool {
	g, _ := SkillCheck([]int8{skillValue(c, n)}, h, rng)
	return g >= rules.Made
}

// treatInjuries begins TreatInjuries: nothing when none is pending (isTreatmentPending is cleared whatever
// follows), when the engine says to skip or when the ship fight ended in a boarding or a lost party.
func (r *Rules) treatInjuries(a rules.TreatInjuries, rng *rules.RNG) rules.Outcome {
	pending := r.treatPending
	r.treatPending = false
	if f := r.space; !pending || a.Skip || f != nil && (f.end == boarded || r.partyDefeated(f)) {
		return rules.Outcome{}
	}
	r.injuries = &treatRun{rng: r.stream(rng), party: slices.Clone(r.party), asked: rules.NoChar}
	return r.treatOn()
}

// antidote is the answer to the question: the antidote is used up and the patient cured, or slips into a coma.
func (r *Rules) antidote(a rules.Antidote) (rules.Outcome, error) {
	t := r.injuries
	if t == nil || t.asked == rules.NoChar {
		return rules.Outcome{}, errNotOffered
	}
	p := r.chars[t.asked]
	if a.Yes {
		owner := r.chars[t.owner]
		ci := owner.Items[t.item]
		if tpl, _, _ := r.itemType(ci); quantity(ci, tpl) <= 1 {
			owner.Items = slices.Delete(owner.Items, t.item, t.item+1)
			r.RecalcStats(owner)
		} else {
			owner.Items[t.item].Ref.Count--
		}
		t.say(p.Name + " is cured")
		t.cure(p)
	} else {
		t.say(p.Name + " slips into a coma.")
		p.Status, p.Enabled = StComatose, false
	}
	t.asked = rules.NoChar
	t.next++
	return r.treatOn(), nil
}

// cure is a poisoned member's cure: unconscious, with no hit points.
func (t *treatRun) cure(p *Character) {
	p.Status, p.Enabled, p.HP = StUnconscious, false, 0
	t.treated = true
}

// treatOn runs the poison pass from its patient, then the wound pass and the summary, until the antidote question.
func (r *Rules) treatOn() rules.Outcome {
	t := r.injuries
	r.treatPoison()
	o := rules.Outcome{}
	if t.asked != rules.NoChar {
		o.Confirm = "Use antidote on " + r.chars[t.asked].Name + "? "
	} else {
		r.treatWounds()
		r.injuries = nil
	}
	o.Effects, t.out = t.out, nil
	return o
}

// treatPoison is the poison pass from patient t.next: each medic in turn tries Treat Poisoning; when none cures,
// the first antidote any member carries is offered.
func (r *Rules) treatPoison() {
	t := r.injuries
	for ; t.next < len(t.party); t.next++ {
		p := r.chars[t.party[t.next]]
		if p == nil || p.Status != StPoisoned {
			continue
		}
		t.injured = true
		if !t.headed {
			t.headed = true
			t.say(treatPoisoningHeading)
		}
		cured := false
		for _, id := range t.party {
			m := r.chars[id]
			if cured || m == nil || m.Career != xxvc.Medic || m.Status != StOkay {
				continue
			}
			t.medics = true
			if r.skillMade(t.rng, m, TreatPoisoning, rules.Normal) {
				cured = true
				t.say(m.Name + " " + p.Name + " cured")
				t.cure(p)
			}
		}
		if cured {
			continue
		}
		t.say(p.Name + " is dying of poison!")
		if owner, item, ok := r.firstAntidote(t.party); ok {
			t.asked, t.owner, t.item = t.party[t.next], owner, item
			return
		}
		t.say(p.Name + " slips into a coma.")
		p.Status, p.Enabled = StComatose, false
	}
}

// firstAntidote is UseAntidote's search: the first member in order carrying an antidote, and its place.
func (r *Rules) firstAntidote(party []rules.CharID) (rules.CharID, int, bool) {
	for _, id := range party {
		c := r.chars[id]
		if c == nil {
			continue
		}
		if i := slices.IndexFunc(c.Items, func(ci Carried) bool { return r.carriesKind(ci, antidoteKind) }); i >= 0 {
			return id, i, true
		}
	}
	return rules.NoChar, 0, false
}

// treatWounds is the wound pass: each hurt member of the team goes through four passes, each medic in turn
// trying that pass's skill; then the summary. UNKILLABLE puts a member holding it back on its feet.
func (r *Rules) treatWounds() {
	t := r.injuries
	headed := false
	for _, pid := range t.party {
		p := r.chars[pid]
		if p == nil {
			continue
		}
		done := true
		if (p.Status == StOkay || p.Status == StDying || p.Status == StUnconscious) && p.team() && p.HP < p.MaxHP {
			done = false
			t.injured = true
		}
		for pass := range healPasses {
			for _, mid := range t.party {
				m := r.chars[mid]
				if done || m == nil || m.Status != StOkay || !m.team() || m.HP <= 0 {
					continue
				}
				skill, hardness, amount := FirstAid, rules.Normal, 0
				switch pass {
				case healPassFirstAid:
					amount = rules.RNGDice(t.rng, 1, 6)
					if p.HP+amount > firstAidHealCap {
						amount = firstAidHealCap - p.HP
					}
				case healPassCritical:
					skill, amount = TreatCritical, rules.RNGDice(t.rng, 3, 8)+3
					if m == p {
						hardness = rules.VeryHard
					}
				case healPassSerious:
					skill, amount = TreatSerious, rules.RNGDice(t.rng, 2, 8)+1
				case healPassLight:
					skill, amount = TreatLight, rules.RNGDice(t.rng, 1, 8)
				}
				if pass != healPassFirstAid {
					amount *= 2
				}
				amount = min(amount, p.MaxHP-p.HP)
				if m.Skills.Level(skill) > 0 {
					t.medics = true
				}
				if amount <= 0 || !r.skillMade(t.rng, m, skill, hardness) {
					continue
				}
				if !headed {
					headed = true
					t.say(treatWoundsHeading)
				}
				t.say(m.Name + " " + p.Name + " " + strconv.Itoa(amount))
				p.HP += amount
				p.Status, p.Enabled = StOkay, true
				if pass != healPassFirstAid {
					done = true
				}
				t.treated = true
			}
		}
		if p.team() && p.HP < unkillableMinHP && slices.ContainsFunc(p.Effects, func(e Effect) bool { return e.ID == EffUnkillable }) {
			p.Status, p.Enabled, p.HP = StOkay, true, unkillableMinHP
		}
	}
	switch {
	case !t.injured:
		t.say("  No injuries were sustained.")
	case !t.medics:
		t.say("  No medics are available.")
	case !t.treated:
		t.say("  No injuries were successfully treated.")
	default:
		t.say("  Healing completed.")
	}
}
