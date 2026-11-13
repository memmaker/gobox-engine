package engine

import (
	"fmt"
	"gobox/core"
	"slices"
)

// rosterSel is the member the roster highlights: SelectCharacter's
// candidate while it waits (-1: none), else the selected one.
func (e *Engine) rosterSel() int {
	switch {
	case e.rosterNone:
		return -1
	case e.rosterHi != core.NoChar:
		return slices.Index(e.partyIDs(), e.rosterHi)
	}
	return e.sel
}

// selected is the selected member, an empty one in an empty party: the
// Engine's one placeholder, emptied on each call, so its CharID is one id
// and never a new registry entry.
func (e *Engine) selected() core.CharID {
	if e.partySize() == 0 {
		o, err := e.ruleset().Resolve(nil, core.EmptyMember{}, nil)
		if err != nil {
			panic(fmt.Sprintf("selected: %v", err)) // never refused
		}
		return o.Effects[0].(core.EmptyMemberIs).Who
	}
	return e.partyIDs()[e.sel]
}

// selectMember makes member n the selected one. Every switch the player
// makes goes through it: portraits, roster clicks, F1-F8, the keypad.
func (e *Engine) selectMember(n int) {
	if n >= 0 && n < e.partySize() {
		e.sel = n
	}
}

// actor selects member n (if there is one) and returns the selected member.
func (e *Engine) actor(n int) core.CharID {
	e.selectMember(n)
	return e.selected()
}

// as hands member n to the rules as the actor of the next call (WordWho).
func (e *Engine) as(n int) { e.setWorldWord(core.WordWho, n) }

// join adds c to the party as the rules' AddToParty, which makes it the
// actor (DS:5D92), and selects it.
func (e *Engine) join(cID core.CharID) {
	e.do(core.JoinParty{Who: cID})
	e.follow()
}

// follow selects the rules' actor after a call that moved it.
func (e *Engine) follow() { e.sel = e.worldWord(core.WordWho) }

// PartyMenu = ovr17:0161: the party menu takes decisions (MainMenu)
// until the player begins adventuring or loads a game that starts play
// (beginFlag != 0), then GameLoop (the cont it pushes) runs. extra: the front end
// lists the functions in its own order and asks its questions (Quit,
// Drop); there is no Save (F9 saves).
func (e *Engine) PartyMenu() {
	e.push(cont{kind: kGameLoop})
	e.partyMenu(false, true)
}

// partyMenu puts the party menu up; hall: a script's (the training hall);
// chain: the cont on top of the stack goes on once it is closed.
func (e *Engine) partyMenu(hall, chain bool) {
	if hall || !e.songMusic(songTitle, 0) { // extra: the title plays on under the main menu
		e.Music(3) // Amiga 0x2D35C
	}
	e.set = 3      // DS:5376 = 3
	if e.Explore { // -explore: straight into the 3D view (debugging the renderer); GameLoop clears it
		if e.partySize() == 0 && e.ExploreSlot != "" && e.loadFrom(e.ExploreSlot) { // -slot: a saved game instead
			e.Explore = false
			e.loadedGame = e.beginFlag() != 0 // DS:49F8, as the party menu's Load (LoadNamed): the game resumes, not afresh
			e.ret()                           // GameLoop
			return
		}
		if e.partySize() == 0 {
			if err := e.QuickStart(); err != nil {
				e.say(err.Error())
			}
		}
		if e.partySize() > 0 {
			e.setWorldWord(core.WordBeginFlag, 4)
			if chain {
				e.ret()
			}
			return
		}
	}
	debug := 0
	if e.FreeTraining { // -freetrain
		debug = 1
	}
	e.setWorldWord(core.WordDebug, debug)
	e.ext.hall = hall // a script's (the training hall): the game goes on under it
	e.setWorldWord(core.WordBeginFlag, 0)
	e.partyRedraw()
	e.menu.on, e.menu.chain, e.menu.seq = true, chain, e.opens()
}

// inParty: the party menu's own decisions are open.
func (e *Engine) inParty() bool {
	return e.onTop() == onMenu && !e.view.on && !e.snap.invOn && e.creation == nil && e.worldWord(core.WordTraining) == 0 && e.eclWait == nil
}

// onTop is camp's or the party menu's on flag when it is up with nothing
// opened over it (nil: neither).
func (e *Engine) onTop() decider {
	if t := e.top(); t == onCamp || t == onMenu {
		return t
	}
	return onNone
}

// campOn: camp is up (CampMenu).
func (e *Engine) campOn() bool { return e.worldWord(core.WordCampOn) != 0 }

// stateRedraw draws camp or the party menu again when what was over
// it went; false: neither is on top.
func (e *Engine) stateRedraw() bool {
	switch e.onTop() {
	case onCamp:
		e.DrawMainScreen()
	case onMenu:
		e.partyRedraw()
	default:
		return false
	}
	return true
}

// closeParty: play starts; what follows the party menu runs.
func (e *Engine) closeParty() {
	e.menu.on = false
	if e.menu.chain {
		e.menu.chain = false
		e.ret()
	}
}

// begin = the party menu's BEGIN Adventuring (the hall's Leave).
func (e *Engine) begin() error {
	if !e.inParty() || !slices.Contains(e.partyFunctions(), partyBegin) {
		return core.ErrNotNow
	}
	e.setWorldWord(core.WordBeginFlag, 4)
	e.ext.hall = false
	e.clearPrompt()
	e.closeParty()
	return nil
}

// removeMember = the party menu's Remove: member n (< 0: the selected
// one) goes back to the saved characters.
func (e *Engine) removeMember(n int) error {
	if !e.inParty() || e.partySize() == 0 {
		return core.ErrNotNow
	}
	if err := e.SaveCharacter(e.actor(n), ""); err != nil {
		return err // not saved: the character stays in the party
	}
	e.RemoveSelected(false, true)
	return nil
}

// lead makes member n the selected one (the party menu, camp, a wait).
func (e *Engine) lead(n int) error {
	if n < 0 || n >= e.partySize() || e.worldWord(core.WordBeginFlag) == 5 {
		return core.ErrNotNow
	}
	e.selectMember(n)
	return nil
}

// QuitGame = the party menu's Quit (the player confirmed).
func (e *Engine) QuitGame() error {
	if !e.inParty() {
		return core.ErrNotNow
	}
	e.Exit()
	return nil
}

// partyRedraw clears the party menu's page: the menu and the roster are
// the presenter's (MainMenu, Party).
func (e *Engine) partyRedraw() {
	e.snap.saves = len(e.slots()) > 0
	e.ClearScreen()
}

// partyFunctions are the party menu's decisions open now (ovr17:0161's
// enabled entries). extra: Quick Start and Forget (a saved character)
// without a party, Load only then and with saved games.
func (e *Engine) partyFunctions() []partyFn {
	if e.partySize() == 0 {
		fs := []partyFn{partyQuickStart, partyAdd, partyCreate, partyForget, partyQuit}
		if e.snap.saves {
			fs = append([]partyFn{partyLoad}, fs...)
		}
		return fs
	}
	fs := []partyFn{partyBegin, partyAdd, partyCreate, partyDrop, partyModify}
	if e.mayTrain() && (e.worldWord(core.WordTrainable) > 0 || e.worldWord(core.WordDebug) != 0) { // only when the selected member would rise
		fs = append(fs, partyTraining)
	}
	if e.ext.hall && len(e.classOffers(e.selected())) > 0 {
		fs = append(fs, partyNewClass)
	}
	return append(fs, partyView, partyRemove, partyQuit)
}

// RemoveSelected = ovr17:3291 (keepCount, freeIcon): the selected member
// leaves the party list; the one before it (or the new first) becomes
// selected. Unless keepCount, the party count (game +0x67C) drops.
func (e *Engine) RemoveSelected(keepCount, freeIcon bool) {
	if e.partySize() == 0 {
		return
	}
	i := e.sel
	if i >= e.partySize() {
		return
	}
	if freeIcon {
		e.freeIcon(byte(e.fact(e.partyIDs()[i], core.IconFact{}).(core.Number))) // ovr33:014E
	}
	e.ruleset().Dismiss(e.partyIDs()[i])
	if i > 0 {
		e.sel = i - 1
	} else {
		e.sel = 0
	}
	if !keepCount {
		e.setWorldWord(core.WordPartyCount, e.worldWord(core.WordPartyCount)-1)
	}
}

// partyMenu: the party menu is up (menu.on), under the creation sheet,
// the training hall, the sheet and questions too.
type partyMenu struct {
	on    bool
	chain bool // what beginning (or a loaded game) goes on with: the cont on top of the stack
	seq   int  // when it opened (top)
}

// mayTrain is whether the selected member would rise at the training hall now.
func (e *Engine) mayTrain() bool {
	return e.fact(e.selected(), core.TrainFact{}).(core.TrainCard).Rises
}

// applyItems = por ApplyItems (00BA:0BB8) on id, the selected member acting.
func (e *Engine) applyItems(id core.CharID) {
	e.setWorldWord(core.WordWho, e.sel)
	e.must(core.ApplyItems{Who: id})
}
