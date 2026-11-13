package engine

import (
	"gobox/asset"
	"gobox/rules"
)

// Rules: affects, to-hit rolls, saving throws, damage and healing, ported
// from ovr25 and the START helpers it leans on (00BA:16EC ... 2637).

// CharMessage = 00BA:16EC (char, message, row, wait): the name and a
// message in the combat side panel, or in the box under the view.
func (e *Engine) CharMessage(cID rules.CharID, msg string, wait bool) {
	if e.beginFlag() == 5 {
		e.do(rules.CombatSay{Who: cID, Text: msg}) // extra: under the view, not in the side pane; the status row's latest line, as bottomMessage's: no stale note over it
	} else if wait { // extra: the original's flash of the text (MsgDelay, then cleared) is a toast
		e.say(e.ruleset().Sheet(cID).Name + " " + msg)
		return
	} else {
		r := 0x11
		if e.restLine {
			r = 0x12
		}
		e.blank(1, r, paneRight, paneBottom)
		e.text(1, r+1, e.nameColour(cID), e.ruleset().Sheet(cID).Name)
		e.textBlock(1, r+2, paneRight, paneBottom, asset.EGALightGreen, true, msg)
	}
	if wait {
		e.ClearMessageArea()
	}
}

// ClearMessageArea = 00BA:17E3.
func (e *Engine) ClearMessageArea() {
	if e.beginFlag() != 5 {
		e.blank(1, 18, winRight, winBottom)
	}
}
