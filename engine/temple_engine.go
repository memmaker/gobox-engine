package engine

import (
	"errors"
	"gobox/adventure"
	"gobox/rules"
	"slices"
)

// cure is temple cure x for member n (n < 0: the selected one), the
// player agreed to the price.
func (e *Engine) cure(n, x int) error {
	sv, ok := e.templeService(rules.NoChar, x)
	if !e.waiting(PendTemple) || !ok {
		return rules.ErrNotNow
	}
	cID := e.actor(n)
	defer e.waited()
	if sv.Price > 0 { // a temple that heals free has no price to pay
		if o, err := e.ruleset().Resolve(nil, rules.Pay{Gold: uint64(sv.Price)}, nil); err != nil { // ovr05:00BF
			if slices.Contains(o.Effects, rules.Effect(rules.Refused{Why: rules.NoMoney})) {
				return errNoMoney
			}
			return err
		}
	}
	e.emit(rules.Sounded{Cue: rules.SoundCue("recover")})
	done := sv.Done
	if done == "" {
		done = "is cured."
	}
	e.say(e.ruleset().Sheet(cID).Name + " " + done)
	e.as(e.sel)
	e.do(rules.GiveTempleCure{Who: cID, Cure: int32(x)})
	return nil
}

// Temple = ovr05:0D09: the script waits on the temple (PendTemple) until
// the party leaves (Leave).
func (e *Engine) Temple() {
	e.songMusic(adventure.SongTemple, 0) // extra
	greeting := e.paneCopy()             // extra: the priests' greeting stays, as the shop
	e.setPhase(rules.PhaseShop)
	e.DrawMainScreen()
	e.keepPane(greeting)
	e.pw.greeting = greeting
	e.wait(PendTemple)
}

// templeService is the temple's service x as member who needs it (core.NoChar: for no one);
// false when the game's temple has no such service (rules.json town).
func (e *Engine) templeService(who rules.CharID, x int) (rules.TempleCure, bool) {
	v, err := e.ruleset().Fact(who, rules.TempleCureFact{Cure: x})
	if err != nil {
		return rules.TempleCure{}, false
	}
	return v.(rules.TempleCure), true
}

// errNoMoney is a payment the purse cannot meet (core.NoMoney), as the temple words it.
var errNoMoney = errors.New("Not enough money.")
