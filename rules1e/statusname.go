package rules1e

import "gobox/rules"

// StatusName (extra) is c's status as the screens word it: "" past
// StatusGone, "Poisoned" for one dead of poison.
func (e *RuleState) StatusName(c *Character) string {
	if c.Status > StatusGone {
		return ""
	}
	if c.Status == StatusDead && e.FindAffect(c, AffPoisoned) != nil {
		return "Poisoned"
	}
	if c.Status == StatusOkay {
		if ail := e.Ailment(c); ail != "" { // extra: up, but ailing
			return ail
		}
	}
	return StatusNames[c.Status]
}

// Health (extra) is how fit c is: disabled when it is down or gone, or an
// affect takes its turns (the grpTurn ones: held, asleep, helpless ...);
// ailing with another status than Okay or an ailment.
func (e *RuleState) Health(c *Character) rules.Health {
	if c.Okay == 0 || c.Status > StatusAnimated {
		return rules.Disabled
	}
	if g := e.groups(); grpTurn < len(g) {
		for _, t := range g[grpTurn] {
			if e.FindAffect(c, t) != nil {
				return rules.Disabled
			}
		}
	}
	if c.Status != StatusOkay || e.Ailment(c) != "" {
		return rules.Ailing
	}
	return rules.Hale
}
