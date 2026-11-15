package rules1e

import (
	"gobox/rules"
	"slices"
)

// Camp: resting, healing, studying (core/camp.go).

// checkCamp answers the Checks of the camp topic.
func (r *Rules) checkCamp(k rules.Check) (rules.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case rules.HealMinutes:
		return rules.Answer{Value: int32(s.HealMinutes())}, true
	case rules.PartyStudy:
		return rules.Answer{Value: int32(s.PartyStudy())}, true
	case rules.StudyNeeded:
		return rules.Answer{Value: int32(s.StudyNeeded())}, true
	case rules.HealClerics:
		return rules.Answer{Value: int32(len(s.HealClerics()))}, true
	case rules.SpellsAllowed:
		return rules.Answer{Yes: !k.NoMagic}, true
	case rules.RestUnit:
		return rules.Answer{Value: int32(s.Camp.Rest[k.Unit])}, true
	case rules.SkyDay:
		return rules.Answer{Yes: s.SkyDay(k)}, true
	case rules.RestDone:
		return rules.Answer{Yes: s.RestDone(k.UntilHealed)}, true
	default:
		return rules.Answer{}, false
	}
}

// resolveCamp answers the Actions of the camp topic.
func (r *Rules) resolveCamp(a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.CampOpen:
		s.Camp.Heal = 0
		s.Camp.Rest = [7]int{}
		s.Camp.RestEvery, s.Camp.RestChance = a.RestEvery, a.RestChance
	case rules.ClearStudy:
		s.ClearStudy()
	case rules.CampClose:
		s.Camp.RestEvery, s.Camp.RestChance = 0, 0
		for _, c := range s.Party {
			s.PlanSpells(c)
		}
		s.SpellCaster = nil
	case rules.PlanParty:
		for _, c := range s.Party {
			s.PlanSpells(c)
		}
	case rules.ForgetCures:
		for _, c := range s.HealClerics() {
			s.ForgetCures(c)
			s.PlanSpells(c)
		}
	case rules.MemorizeCures:
		for _, c := range s.HealClerics() {
			s.MemorizeCures(c)
		}
	case rules.CastCures:
		s.CastCures(a.NoMagic)
	case rules.SetRest:
		s.Camp.Rest = [7]int{}
		s.Camp.Rest[1] = int(a.Minutes)
		s.NormalizeRest()
	case rules.BeginRest:
		s.BeginRest()
	case rules.StopRest:
		s.Camp.StopRest = true
	case rules.Plan:
		s.PlanSpell(r.chars[a.Who], max(slices.Index(SpellIDs[:], a.Spell), 0), int(a.Count))
	case rules.Scribe:
		s.Scribe(r.chars[a.Who], max(slices.Index(SpellIDs[:], a.Spell), 0))
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}
