package engine

import (
	"slices"

	"gobox/ecl"
	"gobox/rules"
)

func init() {
	eclOps[ecl.OpWallDist] = (*Engine).opWallDistance
	eclOps[ecl.OpSkill] = (*Engine).opSkill
	eclOps[ecl.OpSkillShow] = (*Engine).opSkill
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
