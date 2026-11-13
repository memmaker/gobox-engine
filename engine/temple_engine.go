package engine

import (
	"errors"
	"gobox/core"
	"slices"
)

// cure is temple cure x for member n (n < 0: the selected one), the
// player agreed to the price.
func (e *Engine) cure(n, x int) error {
	sv, ok := e.templeService(core.NoChar, x)
	if !e.waiting(PendTemple) || !ok {
		return core.ErrNotNow
	}
	cID := e.actor(n)
	defer e.waited()
	if sv.Price > 0 { // a temple that heals free has no price to pay
		if o, err := e.ruleset().Resolve(nil, core.Pay{Gold: uint64(sv.Price)}, nil); err != nil { // ovr05:00BF
			if slices.Contains(o.Effects, core.Effect(core.Refused{Why: core.NoMoney})) {
				return errNoMoney
			}
			return err
		}
	}
	e.emit(core.Sounded{Cue: core.SoundCue("recover")})
	done := sv.Done
	if done == "" {
		done = "is cured."
	}
	e.say(e.ruleset().Sheet(cID).Name + " " + done)
	e.setWorldWord(core.WordWho, e.sel)
	e.do(core.GiveTempleCure{Who: cID, Cure: int32(x)})
	return nil
}

// viewMember = the temple's View: member n's sheet (n < 0: the selected one).
func (e *Engine) viewMember(n int) error {
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

// templeService is the temple's service x as member who needs it (core.NoChar: for no one);
// false when the game's temple has no such service (rules.json town).
func (e *Engine) templeService(who core.CharID, x int) (core.TempleCure, bool) {
	v, err := e.ruleset().Fact(who, core.TempleCureFact{Cure: x})
	if err != nil {
		return core.TempleCure{}, false
	}
	return v.(core.TempleCure), true
}

// errNoMoney is a payment the purse cannot meet (core.NoMoney), as the temple words it.
var errNoMoney = errors.New("Not enough money.")
