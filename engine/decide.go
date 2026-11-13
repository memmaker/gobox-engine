package engine

// What takes the player's decisions (docs/engine-redesign.md): no screen
// stack. Each thing the game waits on is a typed state slot that records
// when it opened (a sequence number): a question (eclWait), a script's wait
// (pending), camp (camp.on), the party menu (menu.on), exploring
// (ext.explore). The last one opened is on top (top). What follows what a
// call opened is a cont on the stack (cont.go); it goes on once nothing
// is open (closed).

// decider is the slot that takes the decisions.
type decider int

const (
	onNone    decider = iota
	onExplore         // Exploring: the adventuring commands (exploreVerb)
	onMenu            // MainMenu: the party menu (rule calls)
	onCamp            // Camp: camp (rule calls)
	onWait            // Pending: a script waits (pending.go)
	onAsk             // the Ask: a question waits for its answer (eclask.go)
)

// top is the slot opened last of those open.
func (e *Engine) top() decider {
	best, l := -1, onNone
	try := func(on bool, seq int, x decider) {
		if on && seq > best {
			best, l = seq, x
		}
	}
	try(e.ext.explore, e.ext.exploreSeq, onExplore)
	try(e.menu.on, e.menu.seq, onMenu)
	try(e.campOn(), e.campSeq, onCamp)
	try(e.pending != PendNone, e.waitSeq, onWait)
	try(e.eclWait != nil && e.eclWait.seq > 0, e.eclWaitSeq(), onAsk)
	return l
}

func (e *Engine) eclWaitSeq() int {
	if e.eclWait == nil {
		return 0
	}
	return e.eclWait.seq
}

// opens numbers a slot opening.
func (e *Engine) opens() int {
	e.ext.opened++
	return e.ext.opened
}

// blocked: something open waits for the player (a question, a script's
// wait, camp, the party menu). Nothing under it goes on until it is done.
func (e *Engine) blocked() bool {
	return e.eclWait != nil && e.eclWait.seq > 0 || e.pending != PendNone || e.campOn() || e.menu.on
}

// blockedRet goes on with what follows unless something waits.
func (e *Engine) blockedRet() {
	if !e.blocked() {
		e.ret()
	}
}

// closed ends a slot: when nothing else is open the cont on top of the
// stack goes on; then the slot under it draws itself again, unless that
// opened another or dropped the game (unwindReset).
func (e *Engine) closed() {
	m, r := e.ext.opened, e.resets
	if !e.blocked() {
		e.ret()
	}
	if e.ext.opened == m && e.resets == r {
		e.redrawTop()
	}
}

// redrawTop: what is on top draws itself again (what was over it went).
func (e *Engine) redrawTop() {
	switch e.top() {
	case onAsk:
		e.showAsk()
	case onWait:
		e.waitRedrawNow()
	case onCamp, onMenu:
		e.stateRedraw()
	case onExplore:
		if e.over {
			e.endExplore()
		} else {
			e.awaitCommand()
		}
	}
}
