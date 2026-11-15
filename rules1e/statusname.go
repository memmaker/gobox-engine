package rules1e

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
