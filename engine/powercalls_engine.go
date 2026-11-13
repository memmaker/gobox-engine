package engine

import (
	"slices"

	"gobox/game"
	"gobox/rules"
)

// UsePower uses member n's (< 0: the selected one) power k out of combat
// (Power.Kind; Cast also as play.Cast out of a fight): spell and
// target (a member, -1: not picked yet) are the picks it asked for, sure
// its yes/no answered Yes. Magic is a list (EffectsList).
func (e *Engine) UsePower(n int, k game.PowerKind, spell, target int, sure bool) error {
	if e.worldWord(rules.WordBeginFlag) == 5 || e.partySize() == 0 {
		return rules.ErrNotNow
	}
	who := e.actor(n)
	if k != game.PowerCast && !slices.ContainsFunc(e.powers(who), func(p Power) bool { return p.Kind == localPower[k] }) {
		return rules.ErrNotNow
	}
	p := pendingCall{kind: callPower, member: n, power: k, spell: spell, target: target}
	party := e.partyIDs()
	target0 := rules.NoChar
	if target >= 0 && target < len(party) {
		target0 = party[target]
	}
	switch k {
	case game.PowerHeal: // CotAB PaladinHeal: lay on hands for 2 hp per level
		if target0 == rules.NoChar {
			return e.askTarget("Heal whom?", p)
		}
		name := e.ruleset().Sheet(target0).Name
		out, err := e.ruleset().Resolve(nil, rules.PaladinHeal{Who: who, Target: target0}, nil)
		if err != nil {
			return err
		}
		if len(out.Effects) > 0 {
			e.DisplayMessageWait(game.InkLightGreen, name+" feels better")
		} else {
			e.DisplayMessageWait(game.InkLightGreen, name+" is unaffected")
		}
	case game.PowerCure: // CotAB PaladinCureDisease: the temple's diseases go
		if target0 == rules.NoChar {
			return e.askTarget("Cure whom?", p)
		}
		name := e.ruleset().Sheet(target0).Name
		if !e.ruleset().Check(rules.Diseased{Who: target0}, nil).Yes && !sure {
			return e.askSure(name+" is not diseased. Cure anyway?", p)
		}
		if _, err := e.ruleset().Resolve(nil, rules.PaladinCure{Who: who, Target: target0}, nil); err != nil {
			return err
		}
		e.DisplayMessageWait(game.InkLightGreen, name+" is cured")
	case game.PowerCast: // CastMenu = ovr16:04A1, one cast (extra: the pick list, as the combat's Cast)
		if _, err := e.ruleset().Resolve(nil, rules.OpenCast{Who: who}, nil); err != nil {
			return rules.ErrNotNow
		}
		spells := e.memorizedSpells()
		if len(spells) == 0 {
			e.CharMessage(who, "has no spells memorized", true)
			return rules.ErrNotNow
		}
		if !slices.ContainsFunc(spells, func(s Spell) bool { return s.ID == spell }) {
			return e.askSpell(spells, p)
		}
		sure, on, err := e.castAsk(spell, false, target, sure, p)
		if err != nil {
			return err
		}
		e.do(rules.CampCast{Who: who, Spell: e.spellID(spell), Target: on, Sure: sure})
		e.DrawMainScreen()
	default:
		return rules.ErrNotNow
	}
	return nil
}

// memorizedSpells are the selected member's memorized spells, once each.
func (e *Engine) memorizedSpells() []Spell {
	if !e.buildSpellList(0) {
		return nil
	}
	var spells []Spell
	for _, id := range e.spellIDs {
		if !slices.ContainsFunc(spells, func(s Spell) bool { return s.ID == int(id) }) { // memorized twice: once in the list
			spells = append(spells, Spell{ID: int(id), Name: e.castCard(int(id)).Name})
		}
	}
	return spells
}

// castAsk is castAsks out of combat as the pending question of call p: the
// yes/no of a combat-only spell or item and the member a spell goes on
// (target -1 not picked yet), asked before the cast runs: the answers
// (sure: yes to the yes/no; on: the member, NoChar none).
func (e *Engine) castAsk(spell int, usingItem bool, target int, sure bool, p pendingCall) (bool, rules.CharID, error) {
	card := e.castCard(spell)
	switch {
	case card.Aim == rules.CastAimCombatOnly:
		if !sure {
			prompt := "Lose it?"
			if !usingItem {
				e.text(1, 19, game.InkLightGreen, card.Name)
				e.text(1, 20, game.InkLightGreen, "is a combat-only spell...")
			} else {
				e.text(1, 19, game.InkLightGreen, "That Item")
				e.text(1, 20, game.InkLightGreen, "is a combat-only item...")
				prompt = "Use it?"
			}
			return false, rules.NoChar, e.askSure(prompt, p)
		}
		return true, rules.NoChar, nil
	case !e.inFight() && card.Aim == rules.CastAimMember:
		party := e.partyIDs()
		if target < 0 || target >= len(party) {
			e.DrawMainScreen()
			return false, rules.NoChar, e.askTarget("Cast Spell on whom", p)
		}
		return false, party[target], nil
	}
	return false, rules.NoChar, nil
}

// castCard is spell number n as a cast out of a fight sees it.
func (e *Engine) castCard(n int) rules.CastCard {
	return e.fact(rules.NoChar, rules.CastCardFact{Number: n}).(rules.CastCard)
}

// localPower is a game.PowerKind as the member's Power says it.
var localPower = map[game.PowerKind]PowerKind{game.PowerCast: PowerSpells, game.PowerDisplay: PowerMagic, game.PowerHeal: PowerHeal, game.PowerCure: PowerCure}

// EffectsList is member n's spell effects (the sheet's Magic,
// displayAffects = ovr16:0FBC) as the nodes of a list.
func (e *Engine) EffectsList(n int) []*ListNode {
	e.setWorldWord(rules.WordWho, e.sel) // as before: the list names the actor
	id := e.selected()
	if party := e.partyIDs(); n >= 0 && n < len(party) {
		id = party[n]
	}
	sh := e.ruleset().Sheet(id)
	list := []*ListNode{{Text: " "}, {Text: sh.Name, Flag: 1}}
	for _, name := range sh.Effects {
		list = append(list, &ListNode{Text: rules.Cut(" "+name, textCols)})
	}
	if len(sh.Effects) == 0 {
		list = append(list, &ListNode{Text: " <No Spell Effects>"})
	}
	return list
}
