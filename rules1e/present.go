package rules1e

import (
	"gobox/core"
	"gobox/core/goldbox"
)

// The rules' side of the drawing (ARCHITECTURE.md step S): rule code shows
// nothing; it records core.Effects (emit) in the order they happen, and the
// engine's side applies them in that order (present). A computer turn hands
// back its Outcome (computerTurn).

// unit is a combatant's index (DS:65B9, 1-based) as a board unit.
func unit(i int) core.UnitID { return core.UnitID(i) }

// Emit records f for the presenter.
func (e *RuleState) Emit(f core.Effect) { e.Fx = append(e.Fx, f) }

// notice records a line of the rules' text with no one's name (the
// original's BottomMessage), cut to the original's row.
func (e *RuleState) notice(msg string) { e.CombatSay(nil, goldbox.Cut(msg, msgMax), true) }

// narrate records a line of the rules' text as it stands (the remaster's note).
func (e *RuleState) narrate(s string) { e.Emit(core.Narrated{Text: s}) }

// Sound records a rule's Sound (the original's PlaySound), by the Sound
// pack's event that names it.
func (e *RuleState) Sound(cue string) { e.Emit(core.Sounded{Cue: core.SoundCue(cue)}) }

// say1 is the text of a message of c, as the original wrote it after the name.
func say1(c *Character, msg string) string { return core.LogJoin(name(c), msg) }
