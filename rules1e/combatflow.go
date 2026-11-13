package rules1e

import (
	"gobox/core"
	"testing"
)

// The fight as a state machine (docs/engine-redesign.md, step 5): the
// script waits on the fight (PendFight); the presenter steps it a turn at
// a time (CombatStep), and a member's turn is its rule calls
// (TurnChoice). The turn and aiming are phases of the fight (cs.turn,
// cs.aim), not screens; a yes/no it needs (Flee?, Attack Ally?) is the
// pending question (cs.ask, then playask.go); its Yes makes the call again sure. What the
// rules show is play's facts (e.events; the original's Delay and MsgDelay are
// gobox's pacing).

// Pause is a moment the original held the page as it stood (the presenter
// paces the fight itself: gobox's), and a key wait outside a fight is a bug.
func (e *RuleState) Pause(ms int, key bool) {
	if key && !e.Fight.On && testing.Testing() {
		panic("pause: a key wait outside a fight never waits (BUG-CLASSES 11): use keyWait")
	}
}

// factAttack records a's swing at b: a hit's harm is shown (Attacked's
// Damage), so the ApplyDamage it causes adds no Struck (present).
func (e *RuleState) factAttack(a, b *Character, hit bool, shown int) {
	if !e.Fight.On {
		return
	}
	e.Emit(core.Attacked{From: unit(e.CombatIndex(a)), To: unit(e.CombatIndex(b)), Hit: hit, Damage: int32(shown)})
}

// factHarm records c's harm: a spell's, a trap's; a hit's own is its Attacked.
func (e *RuleState) factHarm(c *Character, dmg int) {
	if !e.Fight.On {
		return
	}
	e.Emit(core.Struck{Unit: unit(e.CombatIndex(c)), Damage: int32(dmg)})
}

// factFell records c's fall.
func (e *RuleState) factFell(c *Character) {
	if !e.Fight.On {
		return
	}
	e.Emit(core.Fell{Unit: unit(e.CombatIndex(c))})
}

// FightState is the fight's part of the rule state: whether one is up, the
// aims a spell was given, the maneuver armed, the forced and free moves.
type FightState struct {
	On     bool          // a fight is up
	Aims   []AimAnswer   // the aims taken before a spell is cast, in the order its targeting asks for them
	Choice int32         // extra: the reply to the spell's own question (core.SpellAim.Choices), carried in like Aims
	Armed  Armed         // extra: the maneuver the actor's next weapon attack carries (combatrules.go)
	Aimed  core.BoardPos // extra: the square an AimArea maneuver was aimed at (Aimed)
	forced bool          // extra: MoveTo's steps: no Moved hook, no moves spent
	free   bool          // extra: a free attack (an opportunity attack) is being made (Opportunity)
	Key    AiKey         // the key the computer turn's polls take (computerTurn, aiKeys)
}

// AimAnswer is one aim the player took (playerAim's results).
type AimAnswer struct {
	Found bool
	T     AimTarget
}

// aimNow is chooseSpellTarget's aim: the next one aimSpell took (none
// left: not found, as a cancelled aim).
func (e *RuleState) aimNow(c *Character, rng, allowGround int) (found bool, t AimTarget) {
	if len(e.Fight.Aims) > 0 {
		a := e.Fight.Aims[0]
		e.Fight.Aims = e.Fight.Aims[1:]
		return a.Found, a.T
	}
	return
}
