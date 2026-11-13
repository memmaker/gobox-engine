package engine

import (
	"gobox/rules"
)

// waitData is what a wait keeps for its page.
type waitData struct {
	greeting [paneRows][]textCell // PendShop, PendTemple: the keeper's greeting, kept on the pane
	anim     bool                 // PendLoot: the animation setting to put back
}

// wait suspends the script on p until resume answers it; what each kind
// draws again (waitRedrawNow), does after a rule call (waitEachNow) and
// does once it is answered (leaveWait) are by p.
func (e *Engine) wait(p Pending) {
	e.pending, e.waitSeq = p, e.opens()
}

// resume answers the wait: it is left, and what follows it runs.
func (e *Engine) resume() {
	e.leaveWait()
	e.closed()
}

// leaveWait closes the wait and does what its kind does then.
func (e *Engine) leaveWait() {
	p := e.pending
	e.pending = PendNone
	switch p {
	case PendLoot:
		e.animInput = e.pw.anim
	case PendFight:
		e.combatEnd()
	case PendDoor:
		if w := e.ext.door; w != nil {
			e.ext.doorThrough = w.through
		}
		e.ext.door = nil
	case PendShop:
		e.emit(rules.Sounded{Cue: rules.SoundCue("store_leave")})
	case PendTemple:
		e.blank(0, 0, lastCol, promptRow) // ovr05:0C18
		e.DrawMainScreen()
	}
}

// waitRedrawNow draws the wait's page again when what was over it goes.
func (e *Engine) waitRedrawNow() {
	switch e.pending {
	case PendLoot:
		e.lootRedraw()
	case PendFight:
		e.turnOn()
	case PendShop:
		e.DrawMainScreen()
		e.keepPane(e.pw.greeting)
	case PendTemple:
		e.keepPane(e.pw.greeting)
	}
}

// waitEachNow is the wait's work after each rule call on it.
func (e *Engine) waitEachNow() {
	if e.pending == PendTemple {
		e.keepPane(e.pw.greeting)
	}
}

// waiting: the script waits on one of ps, its wait on top (nothing over it).
func (e *Engine) waiting(ps ...Pending) bool {
	if e.top() != onWait || e.view.on || e.snap.invOn { // the sheet or the inventory over it
		return false
	}
	for _, p := range ps {
		if e.pending == p {
			return true
		}
	}
	return false
}

// Leave = the shop's and the temple's Exit: the valued gem or jewel is
// sold and the script goes on.
func (e *Engine) Leave() error {
	if !e.waiting(PendShop, PendTemple) {
		return rules.ErrNotNow
	}
	e.sellValued()
	e.resume()
	return nil
}

// Appraise values a gem (jewelry: a jewel) in the shop or the temple.
func (e *Engine) Appraise(jewelry bool) error {
	if !e.waiting(PendShop, PendTemple) {
		return rules.ErrNotNow
	}
	e.appraiseValue(jewelry)
	e.waited()
	return nil
}

// KeepAppraised keeps (keep) or sells the one appraised.
func (e *Engine) KeepAppraised(keep bool) error {
	if !e.waiting(PendShop, PendTemple) {
		return rules.ErrNotNow
	}
	e.keepValued(keep)
	e.waited()
	return nil
}

// waited runs the wait's each after a rule call on it.
func (e *Engine) waited() { e.waitEachNow() }
