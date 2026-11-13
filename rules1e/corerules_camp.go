package rules1e

import (
	"gobox/core"
	"slices"
)

// Camp: resting, healing, studying (core/camp.go).

// checkCamp answers the Checks of the camp topic.
func (r *Rules) checkCamp(k core.Check) (core.Answer, bool) {
	s := r.S
	switch k := k.(type) {
	case core.HealMinutes:
		return core.Answer{Value: int32(s.HealMinutes())}, true
	case core.PartyStudy:
		return core.Answer{Value: int32(s.PartyStudy())}, true
	case core.HealClerics:
		return core.Answer{Value: int32(len(s.HealClerics()))}, true
	case core.SpellsAllowed:
		return core.Answer{Yes: s.SpellsAllowed()}, true
	case core.RestUnit:
		return core.Answer{Value: int32(s.Camp.Rest[k.Unit])}, true
	case core.SkyDay:
		return core.Answer{Yes: s.SkyDay(k.Inside)}, true
	case core.RestDone:
		return core.Answer{Yes: s.RestDone(k.UntilHealed)}, true
	default:
		return core.Answer{}, false
	}
}

// resolveCamp answers the Actions of the camp topic.
func (r *Rules) resolveCamp(a core.Action) (core.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case core.CampOpen:
		s.Camp.Heal = 0
		s.Camp.Rest = [7]int{}
	case core.ClearStudy:
		s.ClearStudy()
	case core.CampClose:
		for _, c := range s.Party {
			s.PlanSpells(c)
		}
		s.SpellCaster = nil
	case core.PlanParty:
		for _, c := range s.Party {
			s.PlanSpells(c)
		}
	case core.ForgetCures:
		for _, c := range s.HealClerics() {
			s.ForgetCures(c)
			s.PlanSpells(c)
		}
	case core.MemorizeCures:
		for _, c := range s.HealClerics() {
			s.MemorizeCures(c)
		}
	case core.CastCures:
		s.CastCures()
	case core.SetRest:
		s.Camp.Rest = [7]int{}
		s.Camp.Rest[1] = int(a.Minutes)
		s.NormalizeRest()
	case core.BeginRest:
		s.BeginRest()
	case core.StopRest:
		s.Camp.StopRest = true
	case core.Plan:
		s.PlanSpell(r.chars[a.Who], max(slices.Index(SpellIDs[:], a.Spell), 0), int(a.Count))
	case core.Scribe:
		s.Scribe(r.chars[a.Who], max(slices.Index(SpellIDs[:], a.Spell), 0))
	default:
		return core.Outcome{}, false, nil
	}
	return core.Outcome{}, true, nil
}
