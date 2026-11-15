package rules1e

import (
	"gobox/rules"
	"testing"
)

// TestFlightKinds pins what the engine's missile art once switched on.
func TestFlightKinds(t *testing.T) {
	for _, c := range []struct {
		spell rules.SpellID
		item  rules.ItemID
		want  rules.FlightKind
	}{
		{SpellStoneGaze, 0, rules.FlightGaze}, {SpellCharmGaze, 0, rules.FlightGaze},
		{SpellSpitAcid, 0, rules.FlightSpit},
		{SpellBreathWeapon, 0, rules.FlightBolt}, {SpellLightningBolt, 0, rules.FlightBolt}, {SpellJavelinOfLightning, 0, rules.FlightBolt},
		{SpellFireball, 0, rules.FlightDefault},
		{"", ItemIDDart, rules.FlightPointed}, {"", ItemIDJavelin, rules.FlightPointed}, {"", ItemIDQuarrel, rules.FlightPointed},
		{"", ItemIDSpear, rules.FlightPointed}, {"", ItemIDArrow, rules.FlightPointed},
		{"", ItemIDHandAxe, rules.FlightTumbling}, {"", ItemIDClub, rules.FlightTumbling}, {"", ItemIDHammer, rules.FlightTumbling},
		{"", ItemIDHolyWater, rules.FlightFlask}, {"", ItemIDOil, rules.FlightFlask},
		{"", ItemIDSling, rules.FlightSlingStone},
		{"", ItemIDMace, rules.FlightDefault},
	} {
		if got := flight(c.spell, c.item); got != c.want {
			t.Errorf("flight(%q, %d) = %d, want %d", c.spell, c.item, got, c.want)
		}
	}
}
