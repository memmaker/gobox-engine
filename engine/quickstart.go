package engine

import (
	"fmt"
	"gobox/rules"

	"gobox/game"
)

// Quickstart (not in the original): the empty party menu offers a ready-made
// party of six. Each member goes through CreateCharacter with its kind
// from a template, rerolled until its prime abilities are high, saved to
// the roster and then added like the Add command does.

// quickPicks are template q's creation picks (race, gender, class and
// alignment as the ruleset knows them); a template naming what the ruleset
// lacks is a bug of the game's data.
func (e *Engine) quickPicks(q game.QuickPick) rules.Picked {
	o, err := e.ruleset().Resolve(nil, rules.PickKind{Race: q.Race, Gender: q.Gender, Class: q.Class, Align: q.Align, Weapon: q.Weapon, Skill: q.Skill}, nil)
	if err != nil {
		panic(fmt.Sprintf("quick party %s: %v", q.Name, err))
	}
	return o.Effects[0].(rules.Picked)
}

// QuickStart creates the Data.QuickParty members and puts them in the party
// (gobox/creation.QuickParty over porCreation: Make to the roster, Join).
func (e *Engine) QuickStart() error {
	e.do(rules.EmptyParty{})
	_, err := QuickParty(e.Creation(), nil) // nil: the Engine's own seed draws
	e.sel = 0
	return err
}
