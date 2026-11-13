package present

import (
	"gobox/front"
	"gobox/game"
	"log"
	"sync"
)

// Applying actions: every Action the presenter's Actions returns is the
// game's to apply, in order (por: Engine.Apply), after the queued rule
// calls. They are the page's own decisions (ActOK, ActClose, ActChoose,
// ActText, ActYes, ActNo, ActLead, ActMember, the command bar's and the
// exploration keys' kinds) and the screen-wide passes (ActHead,
// ActPortrait, ActFont, ActSound, ActResize, ActMainMenu, ActNote,
// ActNoteRemove, ActIcon). A kind the game does not apply it
// hands to Unhandled.

// Unhandled reports an action the game did not apply: a log line by
// default; a test may set it to collect them.
// A kind is logged once per run (a resize sends one every frame of a drag).
var Unhandled = func(a front.FrontAction) {
	if _, seen := unhandledSeen.LoadOrStore(a.Kind, true); !seen {
		log.Printf("front: action %d (N %d) not applied by the game", a.Kind, a.N)
	}
}

var unhandledSeen sync.Map

// RuleHost is what a screen needs of the front end that shows it: the rule
// calls it queues run on the game between frames.
type RuleHost interface {
	Call(f func(*classicGame))
	Edit(f func(*classicGame)) // Call, after which the laid-out pages are laid out again
	Ask(q *front.Question)     // put q up (nil: none)
	Asking() *front.Question   // the question up (nil: none)
	Shown() Shown              // what the last frame showed
}

// Shown is what a screen needs to know of the view on screen.
type Shown struct {
	Ask, Powers, Sheet, Inventory, Battle bool        // up
	Camp, MainMenu                        bool        // up
	Treasure                              *front.FrontLoot // the loot inventory's (Inventory.Treasure)
	Door                                  *game.Door
	Resting                               bool // camp's rest goes on
	Party                                 int  // the members
}
