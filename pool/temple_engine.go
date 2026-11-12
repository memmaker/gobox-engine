package pool

import (
	"errors"
	"gobox/core"
	"slices"
)

// Cure is temple cure x for member n (n < 0: the selected one), the
// player agreed to the price.
func (e *Engine) Cure(n, x int) error {
	if !e.waiting(PendTemple) || x < 0 || x >= len(templeCures) {
		return core.ErrNotNow
	}
	cID := e.actor(n)
	defer e.waited()
	if o, err := e.ruleset().Resolve(nil, core.Pay{Gold: uint64(templeCures[x].price)}, nil); err != nil { // ovr05:00BF
		if slices.Contains(o.Effects, core.Effect(core.Refused{Why: core.NoMoney})) {
			return errNoMoney
		}
		return err
	}
	e.emit(core.Sounded{Cue: core.SoundCue("recover")})
	e.say(e.ruleset().Sheet(cID).Name + " is cured.")
	e.setWorldWord(core.WordWho, e.sel)
	e.do(core.GiveTempleCure{Who: cID, Cure: int32(x)})
	return nil
}

// ViewMember = the temple's View: member n's sheet (n < 0: the selected one).
func (e *Engine) ViewMember(n int) error {
	fight := e.waiting(PendFight) && e.cs.turn != nil && e.cs.actor != core.NoChar
	if !fight && !e.waiting(PendTemple, PendShop) { // BUG-CLASSES 36: the shop's and the fight's View too, the game's own sheet (State.SheetPage) in full
		return core.ErrNotNow
	}
	if fight {
		n = e.partyIndex(e.cs.actor) // the fight's View: the active member's
	}
	e.actor(n)
	e.openSheet()
	return nil
}

// Temple = ovr05:0D09: the script waits on the temple (PendTemple) until
// the party leaves (Leave).
func (e *Engine) Temple() {
	e.songMusic(songTemple, 0) // extra
	greeting := e.paneCopy()   // extra: the priests' greeting stays, as the shop
	e.setWorldWord(core.WordBeginFlag, 1)
	e.DrawMainScreen()
	e.keepPane(greeting)
	e.pw.greeting = greeting
	e.wait(PendTemple)
}

// templeCures = the services of DS:011A, their prices and the "is not"
// messages of ovr05:0013 (none: always needed).
var templeCures = [9]struct {
	price uint16
	not   string
}{
	{1000, "is not blind."}, {1000, "is not Diseased."}, {100, ""}, {350, ""}, {600, ""},
	{1000, "is not poisoned."}, {5500, "is not dead."}, {3500, "is not cursed."}, {2000, "is not stoned."},
}

// errNoMoney is a payment the purse cannot meet (core.NoMoney), as the temple words it.
var errNoMoney = errors.New("Not enough money.")
