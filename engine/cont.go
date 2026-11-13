package engine

import (
	"gobox/rules"
)

// What follows a flow is data, not a closure: a cont on the engine's
// stack. A step that goes on once a script run (or a camp, a door) ends
// pushes its cont and starts it; the run ends with ret, which pops the
// top cont and does it. A flow that waits for the player simply returns;
// the answer's flow calls ret when it is done. unwindReset empties the
// stack.

type contKind int

const (
	kGameBegan     contKind = iota // GameLoop's start events are done
	kStepEnd                       // stepEnd
	kAreaChanged                   // areaChanged
	kLookDone                      // lookDone
	kSearched                      // the look's events are done (n: opened, x, y)
	kSearchRestore                 // the search pace goes back
	kStepPre                       // the pre-step script is done (flag: split)
	kMoved                         // the step's move is done (flag: split)
	kPreEnd                        // RunEvent{Before}'s script is done (pre, n[0]: the facing kept)
	kAreaStart                     // the new area's start events are done (sel)
	kAreaStep                      // its step events (sel)
	kAreaSettled                   // its square events (sel)
	kCamped                        // the camp event is done
	kCampEnd                       // camp is left
	kCampOver                      // the after-camp event is done
	kMoveStep                      // the locked door's answer: the step goes on
	kStepEvents                    // the square's events are done
	kEclResume                     // the script goes on
	kProgramExit                   // opProgram's encamp is over: the script exits (n[0]: its pc)
	kOpCombatEnd                   // opCombat's screens are gone
	kOpFought                      // opCombat's fight is over
	kOpAfterFight                  // opCombat's AfterCombat is over
	kDevBack                       // a dev scene: the 3D view comes back
	kDevFought                     // a dev scene: the fight is over
	kDevFightEnd                   // a dev scene: its AfterCombat is over
	kDevJournal                    // the journal dev scene
	kDrawMain                      // the main screen is drawn again
	kGameLoop                      // the party menu is left: adventure begins
	kShopBrowse                    // the shop's inventory is closed
	kDevLoot                       // the portrait dev scene: the loot screen follows the sheet
	kNop                           // nothing: a key's wait is all there is to it
	kAfterShared                   // AfterCombat: the hired shares are taken
	kAfterResults                  // AfterCombat: the results page had its key
	kLootDone                      // AfterCombat: the treasure screen is left
	kAfterEnd                      // AfterCombat is over (the monsters' page had its key)
	kVaultDone                     // the vault screen is left
)

// cont is what to do next, with what it needs.
type cont struct {
	kind contKind
	flag bool
	n    [3]int
	sel  rules.CharID
	pre  *preStep
}

// push puts cs on the stack, the last one on top (done first).
func (e *Engine) push(cs ...cont) { e.k = append(e.k, cs...) }

// ret does the cont on top of the stack (nothing when it is empty).
func (e *Engine) ret() {
	n := len(e.k)
	if n == 0 {
		return
	}
	c := e.k[n-1]
	e.k = e.k[:n-1]
	switch c.kind {
	case kGameBegan:
		e.gameBegan(c.flag)
	case kStepEnd:
		e.stepEnd()
	case kAreaChanged:
		e.areaChanged()
	case kLookDone:
		e.lookDone()
	case kSearched:
		e.searched(c.n[0], c.n[1], c.n[2])
	case kSearchRestore:
		e.setWorldWord(rules.WordSearch, int(e.run.searchSave))
		e.lookDone()
	case kStepPre:
		e.stepPre(c.flag)
	case kMoved:
		e.moved(c.flag)
	case kPreEnd:
		e.preEnd(c.pre, c.n[0])
	case kAreaStart:
		e.areaStart(c.sel)
	case kAreaStep:
		e.areaStep(c.sel)
	case kAreaSettled:
		e.areaSettled(c.sel)
	case kCamped:
		e.push(cont{kind: kCampEnd})
		e.CampMenu()
	case kCampEnd:
		e.campEnded(e.campInterrupted)
	case kCampOver:
		e.campOver()
	case kMoveStep:
		e.moveStep(e.ext.doorThrough)
	case kStepEvents:
		e.ext.stepping = false
		e.ret()
	case kEclResume:
		e.eclRun()
	case kProgramExit:
		e.vm.PC = uint16(c.n[0])
		e.opExit()
		e.ret()
	case kOpCombatEnd:
		e.opCombatEnd()
	case kOpFought:
		e.opFought()
	case kOpAfterFight:
		e.opAfterFight()
	case kDrawMain:
		e.DrawMainScreen()
		e.ret()
	case kGameLoop:
		e.GameLoop()
	case kShopBrowse:
		e.snap.shop = false
		e.uncovered()
	case kNop:
		e.ret()
	case kAfterShared:
		e.afterShared()
	case kAfterResults:
		e.afterResults()
	case kLootDone:
		e.do(rules.ClearTreasure{})
		e.afterEnd()
	case kAfterEnd:
		e.afterEnd()
	case kVaultDone:
		e.vaultDone()
	default:
		e.devCont(c) // the dev scenes' steps (not in -tags nodev)
	}
}
