package pool

import (
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/ega"
)

// Items: names, lists and the START helpers around them.

// BottomMessage = 00BA:1693: the string on row 24 for the message delay.
func (e *Engine) BottomMessage(s string) { e.bottomMessage(s, ega.LightGreen) }

// refuse (extra): BottomMessage for a refusal, the text red on the combat
// status row; outside combat the original's green on row 24.
func (e *Engine) refuse(s string) { e.bottomMessage(s, ega.LightRed) }

// bottomMessage: BottomMessage with the colour of the combat status row.
// Extra: in combat the message goes to the View's Log, elsewhere it is the
// window's note (a toast the remaster shows on every page; the original's
// row 24 is under the text box there, and its delay held the game).
func (e *Engine) bottomMessage(s string, colour int) {
	s = goldbox.Cut(s, textCols)
	e.say(s)
	if e.beginFlag() == 5 && e.ruleset().Check(core.BoardUp{}, nil).Yes {
		e.do(core.CombatSay{Text: s})
	}
}
