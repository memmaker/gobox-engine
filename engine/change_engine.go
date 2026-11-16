package engine

import (
	"slices"

	"gobox/rules"
)

// The Change menu (explore.pas CampAlterMenu), from the rest menu and the orbit-map menu: Order and Purge (not on the orbit
// map), Level; Messages and Anim are the front's own settings (message speed and animation are not the game's), so the menu has
// no such rows. The menus ask as the rest menu does: each answer completes into the next question, Exit goes back to the
// menu Change was asked from.

// changeMenuText is GDConsts.changeMenuHead 'Order Purge ' and changeMenuTail 'Messages Anim Level Exit' less the two
// settings; on the orbit map the head is left out (CampAlterMenu: ORBIT_MAP).
const (
	changeMenuText      = "~Order ~Purge ~Level ~Exit"
	changeMenuOrbitText = "~Level ~Exit"
)

// The Change menu's places, by word.
type changePlace int

const (
	changeOrder changePlace = iota
	changePurge
	changeLevel
	changeExit
)

// changePlaces are the places of the menu asked, in its order.
func (e *Engine) changePlaces() []changePlace {
	if e.run.changeOnOrbit {
		return []changePlace{changeLevel, changeExit}
	}
	return []changePlace{changeOrder, changePurge, changeLevel, changeExit}
}

// openChange puts the Change menu up; orbit: asked from the orbit-map menu.
func (e *Engine) openChange(orbit bool) {
	e.run.changeOnOrbit = orbit
	e.askChange()
}

func (e *Engine) askChange() {
	text := changeMenuText
	if e.run.changeOnOrbit {
		text = changeMenuOrbitText
	}
	e.askChoice("", text, false, false, askDone{kind: dChange})
}

// changeBack goes back to the menu Change was asked from.
func (e *Engine) changeBack() {
	if e.run.changeOnOrbit {
		e.askRocketMenu()
		return
	}
	e.askRest()
}

// changeAnswered takes the Change menu's answer n (a place in its text).
func (e *Engine) changeAnswered(n int) {
	places := e.changePlaces()
	if n < 0 || n >= len(places) {
		e.askChange()
		return
	}
	switch places[n] {
	case changeOrder:
		e.askMember(orderPrompt, rules.NoChar, askDone{kind: dOrderWho})
	case changePurge:
		e.askMember(purgePrompt, rules.NoChar, askDone{kind: dPurgeWho})
	case changeLevel:
		e.askChoice("", levelMenuText, false, false, askDone{kind: dLevel})
	case changeExit:
		e.changeBack()
	}
}

// The party order menu (explore.pas PartyOrderMenu: Select a member, then Place it: OrderMoveUp, OrderMoveDown).
const (
	orderPrompt   = "Party Order:"
	orderMenuText = "~Up ~Down ~Exit"
)

// The place of orderMenuText.
const (
	orderUp = iota
	orderDown
)

// orderChosen asks where the member picked goes: up or down one place.
func (e *Engine) orderChosen(who rules.CharID) {
	e.askChoice("", orderMenuText, false, false, askDone{kind: dOrderMove, who: who})
}

// orderMoved moves member who up or down one place (OrderMoveUp, OrderMoveDown: the first goes last and the last first), then asks
// again; Exit is back at the Change menu.
func (e *Engine) orderMoved(who rules.CharID, n int) {
	ids := e.partyIDs()
	at := slices.Index(ids, who)
	if at < 0 || n > orderDown {
		e.askChange()
		return
	}
	to := (at + len(ids) - 1) % len(ids)
	if n == orderDown {
		to = (at + 1) % len(ids)
	}
	if err := e.moveMember(at, to); err != nil {
		e.say(err.Error())
	}
	e.askChoice("", orderMenuText, false, false, askDone{kind: dOrderMove, who: who})
}

// The purge (chargen.pas DropCharacterMenu): two questions, default No, then the member's files go and it leaves the party.
const (
	purgePrompt = "Purge whom?"
	purgeFinal  = "Are you sure? "
)

func purgeFirst(name string) string { return "Drop " + name + " forever? " }

// purgeNo is DropCharacterMenu's words for a member kept.
func purgeKept(name string) string { return name + " gives a sigh of relief." }

func (e *Engine) purgeAsk(text string, d askDone) {
	q := yesNo(text, false)
	q.Default = 0
	e.askYesNo(q, false, d)
}

// purgeChosen asks the first question about member who.
func (e *Engine) purgeChosen(who rules.CharID) {
	e.purgeAsk(purgeFirst(e.ruleset().Sheet(who).Name), askDone{kind: dPurgeSure, who: who})
}

// purgeSure asks the second.
func (e *Engine) purgeSure(who rules.CharID, yes bool) {
	if !yes {
		e.say(purgeKept(e.ruleset().Sheet(who).Name))
		e.purgeBack()
		return
	}
	e.purgeAsk(purgeFinal, askDone{kind: dPurgeDone, who: who})
}

// purgeDone ends the purge (purge: the ruleset's words, the files deleted, the member out of the party). The Pascal does nothing
// more when no member is left (DropCharacterMenu only sets isPartyDefeated): the menu it was asked from stays up, and the game loop
// ends when control reaches it (EclVm GameLoop: the map loop runs until isPartyDefeated, then the party menu).
func (e *Engine) purgeDone(who rules.CharID, yes bool) {
	name := e.ruleset().Sheet(who).Name
	if !yes {
		e.say(purgeKept(name))
		e.purgeBack()
		return
	}
	e.purge(who)
	if e.partySize() == 0 && !e.menu.on {
		e.run.quit = true
	}
	e.purgeBack()
}

// purgeBack is where a purge returns: the party menu (it waits under the question), else the Change menu.
func (e *Engine) purgeBack() {
	if !e.menu.on {
		e.askChange()
	}
}

// The difficulty menu (GDConsts.difficultyMenu 'Novice Cadet Trooper Veteran Ace Exit'): a level sets zone0's difficulty.
const levelMenuText = "~Novice ~Cadet ~Trooper ~Veteran ~Ace ~Exit"

// levels are the menu's levels in its order (c_diff.pas DIFFICULTY_NOVICE .. DIFFICULTY_ACE).
var levels = [...]rules.Difficulty{rules.DifficultyEasiest, rules.DifficultyEasier, rules.DifficultyStandard, rules.DifficultyHarder, rules.DifficultyHardest}

// levelChosen sets the level n names (Exit: none); the Change menu is asked again.
func (e *Engine) levelChosen(n int) {
	if n >= 0 && n < len(levels) {
		e.must(rules.SetDifficulty{Level: levels[n]})
	}
	e.askChange()
}

// purge is DropCharacterMenu's end: the ruleset's line for it, its saved files deleted when the ruleset purges (rules.Purge),
// and it leaves the party.
func (e *Engine) purge(who rules.CharID) {
	name := e.ruleset().Sheet(who).Name
	if a := e.ruleset().Check(rules.Purge{Who: who}, nil); a.Yes {
		e.say(a.Text)
		e.purgeSaved(name)
	}
	e.selectChar(who)
	e.RemoveSelected(false, true)
}
