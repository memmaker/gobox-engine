package rules1e

import "gobox/rules"

// Looks in flight, as the engine asks them (rules.FlightFact).

// spellFlights is the flight kind of the spells that have one; others are FlightDefault.
var spellFlights = map[rules.SpellID]rules.FlightKind{
	SpellStoneGaze:          rules.FlightGaze,
	SpellCharmGaze:          rules.FlightGaze,
	SpellSpitAcid:           rules.FlightSpit,
	SpellBreathWeapon:       rules.FlightBolt,
	SpellLightningBolt:      rules.FlightBolt,
	SpellJavelinOfLightning: rules.FlightBolt,
}

// itemFlights is the flight kind of the item types that have one; others are FlightDefault.
var itemFlights = map[rules.ItemID]rules.FlightKind{
	ItemIDDart:      rules.FlightPointed,
	ItemIDJavelin:   rules.FlightPointed,
	ItemIDQuarrel:   rules.FlightPointed,
	ItemIDSpear:     rules.FlightPointed,
	ItemIDArrow:     rules.FlightPointed,
	ItemIDHandAxe:   rules.FlightTumbling,
	ItemIDClub:      rules.FlightTumbling,
	ItemIDHammer:    rules.FlightTumbling,
	ItemIDHolyWater: rules.FlightFlask,
	ItemIDOil:       rules.FlightFlask,
	ItemIDSling:     rules.FlightSlingStone,
}

// flight is the flight kind of a spell, or of an item type when no spell.
func flight(spell rules.SpellID, item rules.ItemID) rules.FlightKind {
	if spell != "" {
		return spellFlights[spell]
	}
	return itemFlights[item]
}
