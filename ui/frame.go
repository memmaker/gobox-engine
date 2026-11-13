// frame.go: the switch's skeleton (switch-spec.md). The presenter reads
// the game's play.State, which the engine built and shares read-only, and
// UI, the front's own state over it; nothing repacks the State.
package ui

import (
	"image"
	"time"

	"gobox/game"
	"gobox/game/play"
)

// UI is what gobox has up over the game's state that the rules do not
// know: its own questions, the aiming, the screens it opened, the paging,
// the pacing and the toasts. Only the front writes it; game facts never
// go here (they are play.State's).
type UI struct {
	grid       image.Point   // the page's size in cells (ActResize); 40x25 at least
	items      game.ItemsUp  // the inventory screen up, gobox's (OpenItems)
	itemsOf    int           // the member whose items are up (OpenItems), -1: the acting or selected one
	ask        *play.Ask     // gobox's own question up (a maneuver pick), nil none
	moves      []play.Action // the actions ask's options stand for
	refusalAsk *play.Ask     // a Refusal's question put instead (answered by its options' Do)
	aim        *squareAim    // an Aimed action's squares being picked
	spellAim   *aiming       // a combat spell's squares being picked (Combat.Aims)
	cursor     image.Point   // the free cursor's square
	read       *game.Read    // the handout up, kept so the pointer holds while it does
	sheet      int           // the member whose sheet is up, -1 none
	lookSheet  int           // the look-only sheet up (a fight's View): member+1, 0 none
	tiles      bool          // the weapon picker up is the pool bodies by original ref (TilePicker)
	speedUp    bool          // the Speed menu is up (SpeedMenu)
	speed      int           // the game speed (SetSpeed, 0-9): a fight's moment holds speed × 100 ms
	rest       time.Duration // camp's rest still to go, a chunk per Step (PaceRest)
	walk       *walking      // the auto-walk under way, a square per walkPace (PaceWalk)
	seq        uint64        // counts the changes: the front redraws on a new one
	log        []string      // the fight's message log, from the outcomes' text
	note       game.Note     // the status line's toast: the last refusal
	toast      FrontNote     // the front's own toast (Added to journal, a refusal over the items)
	lore       int           // journal notes seen (State.Lore), +1; 0: none counted yet
	title      titleStep     // the title as gobox steps it
}

// titleStep is where gobox has the title: its step, the credits and their page.
type titleStep struct {
	n       int
	credits bool
	page    int
}

// Moment is one thing that happened, as the presenter plays it: the
// rules' Report, how long it holds, and the board or party as they stood
// then. Board and Party are snapshots, never modified; nil Board: the
// live state's. A timed page or screen is Report.Page or Report.Screen,
// drawn over the live state. It replaces front.CombatEvent (switch).
type Moment struct {
	play.Report
	Ms    int           // how long it holds (0: no moment of its own)
	Key   bool          // held until a key ("press any key"), not for Ms
	Board *play.Combat  // in a fight: the board this moment shows
	Party []game.Member // with Board: the party's figures then
}

// state is the game's state now, as the engine built it (read-only; nil:
// no game set, SetEngine).
func (p *Presenter) state() play.State {
	if p.eng == nil {
		return nil
	}
	return p.eng.G.State()
}

// selected is the member the screens act on (selectedOf; the state's
// own when no game is set).
func (p *Presenter) selected() int {
	if p.eng == nil {
		return play.CommonOf(nil).Selected
	}
	return p.eng.r.selectedOf(p.state())
}

// ui is the front's own state over it (nil: no game set).
func (p *Presenter) ui() *UI {
	if p.eng == nil {
		return nil
	}
	return &p.eng.r.UI
}

// asking is the question up over the state: a Refusal's, else the game's
// or gobox's own (nil: none).
func (r *frontState) asking(ps play.State) *play.Ask {
	if r.refusalAsk != nil {
		return r.refusalAsk
	}
	return r.pending(ps)
}

// exploring is whether the party explores and nothing asks: the adventuring
// decisions (ActForward .. ActCastPick) are taken. A question or the door
// menu over the view, or a yes/no over it, takes the keys instead.
func (r *frontState) exploring(ps play.State) bool {
	if _, ok := ps.(play.Exploring); !ok || play.CommonOf(ps).Ask != nil || doorOf(ps) != nil {
		return false
	}
	return !OverMenu(r.modeOf(ps), r.asking(ps))
}

// exploring is r.exploring over the game's state now, and no question of
// the presenter's own up (Ask, the gate).
func (p *Presenter) exploring() bool {
	return p.eng != nil && p.q == nil && p.eng.r.exploring(p.state())
}

// pacing is what goes on with time now (nothing: no game set).
func (p *Presenter) pacing() game.Pacing {
	if p.eng == nil {
		return game.Pacing{}
	}
	return *p.eng.r.pacingOf(p.state())
}
