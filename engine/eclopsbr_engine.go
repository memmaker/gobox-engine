package engine

import (
	"fmt"
	"slices"

	"gobox/ecl"
	"gobox/rules"
)

func init() {
	eclOps[ecl.OpWallDist] = (*Engine).opWallDistance
	eclOps[ecl.OpSkill] = (*Engine).opSkill
	eclOps[ecl.OpSkillShow] = (*Engine).opSkill
	eclOps[ecl.OpTakeItem] = (*Engine).opTakeItem
	eclOps[ecl.OpStash] = (*Engine).opStash
	eclOps[ecl.OpSkillDamage] = (*Engine).opSkillDamage
	eclOps[ecl.OpDuel] = (*Engine).opDuel
	eclOps[ecl.OpShop] = (*Engine).opShop
}

// opWallDistance = BR eclvm.pas EclDistanceAhead (result, direction 0..3): the squares from the party's
// position to a wall that way (MaxMonsterDist, at most 5) into the result.
func (e *Engine) opWallDistance() {
	e.vm.ReadArgs(2)
	dst := e.vm.Ref(1)
	facing := int(byte(e.vm.Arg(2, e.host)) * 2) // a Byte product
	e.host.Set(dst, int64(e.FreeSteps(e.posX, e.posY, facing)))
}

// skillReportPauses is how many message delays SkillShow waits after its message (EclSkillCheck's PauseOnly calls).
const skillReportPauses = 4

// opSkill = BR eclvm.pas EclSkillCheck (skill, whole party, difficulty), the roll the ruleset's SkillRoll
// makes by the selected member or the party's best: the one who rolled goes to the second operand
// (its place in the party), the grade (0 failed badly .. 3 made well) to the third. SkillShow also
// shows the ruleset's report of it.
func (e *Engine) opSkill() {
	e.vm.ReadArgs(3)
	roll := rules.SkillRoll{
		Script:   rules.ScriptSkill(byte(e.vm.Arg(1, e.host))),
		All:      byte(e.vm.Arg(2, e.host)) != 0,
		Hardness: rules.Hardness(byte(e.vm.Arg(3, e.host))),
	}
	who, outcome := e.vm.Ref(2), e.vm.Ref(3)
	party := e.partyIDs()
	switch {
	case roll.All:
		roll.Whom = party
	case len(party) > 0:
		roll.Whom = []rules.CharID{e.selected()}
	}
	a := e.ruleset().Check(roll, nil)
	place := slices.Index(party, a.Who)
	if place < 0 {
		place = len(party) // the original counts the party's members up to the one who rolled: all of them when none did
	}
	e.host.Set(who, int64(place))
	e.host.Set(outcome, int64(a.Grade))
	if e.cmdIs(ecl.OpSkillShow) {
		e.scriptText(true, a.Text)
		e.askPause(skillReportPauses*int(e.msgDelay)*messageDelayUnit, askDone{})
	}
}

// destroy item operand: +128 takes one, the low seven bits the item type (EclDestroyItems' highBit).
const (
	takeOne      = 128
	takeTypeMask = 127
)

// opTakeItem = BR eclvm.pas EclDestroyItems (whole party, type): the items of the type go from the selected
// member or from everyone; with +128 only one of each member's (a stack loses one). BR's own: the
// Pool's DestroyItems takes the whole party's, none by member.
func (e *Engine) opTakeItem() {
	e.vm.ReadArgs(2)
	whole := byte(e.vm.Arg(1, e.host)) != 0
	v := byte(e.vm.Arg(2, e.host))
	id, err := e.itemID(int(v & takeTypeMask))
	if err != nil { // a type without an id: no one carries it
		return
	}
	e.do(rules.TakeItems{Whom: e.whom(whole), Item: id, OneEach: v&takeOne != 0})
}

// whom is the party, or the selected member alone (none in an empty party).
func (e *Engine) whom(party bool) rules.Party {
	switch {
	case party:
		return e.partyIDs()
	case e.partySize() > 0:
		return rules.Party{e.selected()}
	}
	return nil
}

// opStash = BR eclvm.pas EclTransferCharacterLists (0 away, else back): every member's items go onto its
// second list (put away) or come back from it.
func (e *Engine) opStash() {
	e.vm.ReadArgs(1)
	e.do(rules.StashItems{Whom: e.partyIDs(), Away: e.vm.Arg(1, e.host) == 0})
}

// opSkillDamage = BR eclvm.pas EclSkillDamage (skill, whole party, difficulty, dice, sides, bonus): each of the
// selected member or everyone who fails the skill roll takes the damage. The original ends the script after
// the one member of a single roll (its isEclStopped := wholeParty = 0).
func (e *Engine) opSkillDamage() {
	e.vm.ReadArgs(6)
	whole := byte(e.vm.Arg(2, e.host)) != 0
	o, err := e.ruleset().Resolve(rules.Scripted{Party: e.partyIDs()}, rules.SkillDamage{
		Whom: e.whom(whole), Skill: rules.ScriptSkill(byte(e.vm.Arg(1, e.host))), Hardness: rules.Hardness(byte(e.vm.Arg(3, e.host))),
		Dice: int32(byte(e.vm.Arg(4, e.host))), Sides: int32(byte(e.vm.Arg(5, e.host))), Bonus: int32(byte(e.vm.Arg(6, e.host))),
	}, nil)
	if err != nil {
		panic(fmt.Sprintf("script: SkillDamage: %v", err))
	}
	for _, f := range o.Effects {
		if h, ok := f.(rules.Hurt); ok {
			e.HitMessage(h)
		}
	}
	e.vm.Stop = !whole
}

// soloCombat is run.duelMode (DS:84D4) of a fight only the selected member takes part in.
const soloCombat = 1

// opShop = BR eclvm.pas EclShop (list): the shop stocked from item list `list` of the game's tables, at the layout's
// PriceMultiplier (zone 1, sixteenths of a value); the script waits until the party leaves.
func (e *Engine) opShop() {
	e.vm.ReadArgs(1)
	stock := rules.StockShop{List: rules.ShopListID(byte(e.vm.Arg(1, e.host)))}
	if r, ok := e.layoutField(fieldPriceMult); ok {
		stock.Terms.PriceMultiplier = int32(e.host.Get(r))
	}
	e.takeover()
	e.push(cont{kind: kOpCombatEnd})
	e.openShop(stock)
}

// opDuel = BR eclvm.pas EclSoloCombat: Combat with only the selected member fighting. The ruleset
// disables the others for the fight (FightTerms.Duel) and settles only the duelist after it.
func (e *Engine) opDuel() {
	e.run.duelMode = soloCombat
	e.opCombat()
}
