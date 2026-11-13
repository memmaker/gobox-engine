package present

import (
	"gobox/front"
	"gobox/game"
	"gobox/gpu"
	"image"

	"gobox/blobber"
	"gobox/input"
)

// The adventuring commands: the presenter's own list, each a decision
// sent for the selected member (View.Exploring).

type command struct {
	text string
	key  byte
	act  front.ActKind
}

// commands are the decisions the command menu offers on the state up.
func (p *Presenter) commands() []command {
	var c []command
	if p.flags().Interact { // Explicit Interactions: the square's event waits (Space)
		c = append(c, command{"Interact", ' ', front.ActInteract})
	}
	if blobber.Settings.NoGrid { // No Grid's right click is mouselook: the options for the mouse
		c = append(c, command{"Options", 0, actOptions})
	}
	return append(c, []command{
		{"Camp", 'A', front.ActCamp}, // not C (Cast) nor E (every screen's Exit)
		{"Sweep", 'W', front.ActSweep},
		{"Search", 'S', front.ActSearch},
		{"Map", 'M', front.ActMap}, // the keys' own commands, for the mouse and the pad
		{"Overview", 'O', front.ActOverview},
		{"Journal", 'J', actJournal},
		{"Clue Book", 'B', actClueBook},
		{"Help?", '?', front.ActHelp}}...)
}

// memberCommands are the selected member's keys while exploring, kept
// off the command menu (a card's context menu has them for the mouse and
// the pad).
func (p *Presenter) memberCommands() []command {
	c := []command{{"Items", 'I', front.ActInventory}, {"View", 'V', front.ActMember}}
	if p.flags().CanCast {
		c = append(c, command{"Cast", 'C', front.ActCastPick})
	}
	return c
}

// actJournal is the Journal command's: the presenter's own page (journal.go), no decision.
// actOptions is the Options command's: the Esc menu.
const actJournal, actOptions front.ActKind = -1, -2

// runCommand sends command i of the command menu.
func (p *Presenter) runCommand(i int) {
	if i >= 0 && i < len(p.cmds) {
		p.runAct(p.cmds[i].act)
	}
}

// runAct sends a command's decision, or opens the journal.
func (p *Presenter) runAct(a front.ActKind) {
	switch a {
	case actJournal:
		p.openJournal()
		return
	case front.ActOverview:
		p.overview = true
		return
	case front.ActHelp:
		p.openHelp(p.shownView)
		return
	case front.ActMap:
		p.openMap()
		return
	case actClueBook:
		p.openClue(p.shownView)
		return
	case actOptions:
		p.menu = false
		p.view.OpenOptions()
		return
	}
	p.send(a, -1)
}

// exploreKey turns key k into an adventuring decision; false when k is
// not one.
func (p *Presenter) exploreKey(k uint16) bool {
	c := byte(k)
	if c >= 'a' && c <= 'z' {
		c -= 'a' - 'A'
	}
	a := actOf(k)
	if kp := keypad(k); kp != input.None { // the digits: the keypad's arrows
		a = kp
	}
	if p.window() == game.WindowMap { // the area map: a step each way
		if d, ok := map[input.Act]int{input.Up: 0, input.Right: 1, input.Down: 2, input.Left: 3}[a]; ok {
			p.send(front.ActGo, d)
			return true
		}
	} else if act, ok := map[input.Act]front.ActKind{input.Up: front.ActForward, input.Down: front.ActBackward, input.Left: front.ActTurnLeft, input.Right: front.ActTurnRight}[a]; ok {
		if gpu.IsKeyPressed(gpu.KeyShift) && (a == input.Left || a == input.Right) { // Shift: a step aside instead of a turn
			act = map[input.Act]front.ActKind{input.Left: front.ActStrafeLeft, input.Right: front.ActStrafeRight}[a]
		}
		p.send(act, 0)
		return true
	}
	if k >= 128 {
		return false
	}
	for _, cm := range append(p.commands(), p.memberCommands()...) {
		if cm.key == c {
			p.runAct(cm.act)
			return true
		}
	}
	return false
}

// exploreClick takes a left click at (x, y) on the 3D view while the party
// explores (no menu up): the minimap opens the full map, anywhere else
// opens the command menu. The mouse never walks or turns.
func (p *Presenter) exploreClick(x, y int) {
	if _, ok := p.minimap(); ok && p.mapAt.Min.X < p.mapAt.Max.X && image.Pt(x, y).In(p.mapAt) {
		p.openMap()
		return
	}
	p.menu, p.sel = true, 0
}
