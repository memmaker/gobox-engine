package rulesxxvc

import "gobox/rules"

// MinutesPerDay is the clock's day (gdconsts.pas clockUnitScale: 10 minutes a
// turn, 6 turns an hour, 24 hours a day).
const MinutesPerDay = 10 * 6 * 24

// RestAndHeal is gametime.pas RestAndHeal(days) for c: the unconscious wake,
// the okay are enabled and gain one HP a day up to their maximum. It answers
// the HP gained.
func RestAndHeal(c *Character, days int) int {
	if c.Status == StUnconscious {
		c.Status = StOkay
	}
	if c.Status != StOkay {
		return 0
	}
	c.Enabled = true
	was := c.HP
	c.HP = min(c.HP+days, c.MaxHP)
	return c.HP - was
}

// rest is the camp's Rest (explore.pas RestMenu: each Rest is RestAndHeal(1),
// a day): whole days of Minutes, every member in party order.
func (r *Rules) rest(p rules.Party, minutes int32) rules.Outcome {
	var o rules.Outcome
	days := int(minutes / MinutesPerDay)
	if days <= 0 {
		return o
	}
	for _, id := range p {
		if c := r.chars[id]; c != nil {
			if hp := RestAndHeal(c, days); hp > 0 {
				o.Effects = append(o.Effects, rules.Healed{Unit: id, HP: int32(hp)})
			}
		}
	}
	return o
}
