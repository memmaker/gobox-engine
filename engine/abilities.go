package engine

import (
	"gobox/game"
	"gobox/rules"
)

// The abilities a sheet lists are the ruleset's: Catalog.Abilities in its order, worded by Catalog.AbilityNames.

// abilityList is the ruleset's abilities in its order.
func (e *Engine) abilityList() []rules.Ability { return e.ruleset().Catalog().Abilities }

// abilityKey is the state's key for ability a (game.Member.Stats): its abbreviation in lower case.
func (e *Engine) abilityKey(a rules.Ability) string {
	return e.ruleset().Catalog().AbilityNames[a].Key()
}

// strengthPercentKey is the state's key for the exceptional strength (game.StatKeys, Member.Str00): not one of the
// Catalog's abilities, the extra its Strength carries.
const strengthPercentKey = "str00"

// abilityByKey is the ability the state's key names.
func (e *Engine) abilityByKey(key string) (rules.Ability, bool) {
	if key == strengthPercentKey {
		return rules.StrengthPercent, true
	}
	cat := e.ruleset().Catalog()
	for _, a := range cat.Abilities {
		if cat.AbilityNames[a].Key() == key {
			return a, true
		}
	}
	return "", false
}

// statDefs are the abilities of id's sheet in the ruleset's order, with their scores and, for modifiers, mod
// (nil: none shown). Strength carries the exceptional strength the ruleset words as its extra.
func (e *Engine) statDefs(id rules.CharID, mod func(rules.Ability) int) []game.StatDef {
	cat := e.ruleset().Catalog()
	sh := e.ruleset().Sheet(id)
	var out []game.StatDef
	for _, a := range cat.Abilities {
		n, ok := cat.AbilityNames[a]
		if !ok {
			continue
		}
		d := game.StatDef{Key: n.Key(), Label: n.Short, Name: n.Long, Value: int(sh.Scores[a])}
		if a == rules.Strength {
			d.Extra = e.word(id, rules.FactExceptional)
		}
		if mod != nil {
			d.Mod = mod(a)
		}
		out = append(out, d)
	}
	return out
}
