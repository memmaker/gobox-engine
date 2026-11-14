package engine

import (
	"fmt"
	"gobox/rules"
	"math"
	"slices"
	"strconv"
	"strings"

	"gobox/game"
)

// The port's own script commands (extra, gobox ecl V1), placed by POR's
// import fixes where the script's own logic cannot follow the port (one
// party purse in 64-bit gold; Knock spent the rules' way, Felix d3).

// The layout fields the engine names itself (layout/global.txt).
const (
	fieldHours = "Hours" // the day-sky option's comparison (opCompare)
)

// field is the running script's field called name.
// opCastKnock is the cabinets' CastKnock(v) (area 14, Felix d3): Knock cast
// the rules' way (SpendKnock), the slot spent; v is the caster's place in
// the party, or 255 when no okay member has Knock memorized. It was the
// Spell command and the script clearing the slot itself.
func (e *Engine) opCastKnock() {
	e.vm.ReadArgs(1)
	out, err := e.ruleset().Resolve(nil, rules.SpendKnock{}, nil)
	if err != nil {
		panic(fmt.Sprintf("CastKnock: %v", err)) // the action never refuses
	}
	r := int64(noKnock)
	if len(out.Effects) > 0 {
		r = int64(slices.Index(e.partyIDs(), out.Effects[0].(rules.KnockCast).Who))
	}
	e.host.Set(e.vm.Ref(1), r)
}

// noKnock is CastKnock's answer when no one casts.
const noKnock = 255

// gold is the party's gold pieces.
func (e *Engine) gold() uint64 {
	v, err := e.ruleset().Fact(rules.NoChar, rules.PurseFact{})
	if err != nil {
		panic(fmt.Sprintf("native: purse: %v", err)) // the purse is always held
	}
	return v.(rules.Purse).Gold
}

// New Phlan's tavern game labels (area 0): TavernEvent is the "pickpocket"
// check and tavern tales, ResetVisits clears the visit flags, closes the
// picture and exits.
func (e *Engine) scriptLabel(l string) int { return e.vm.Script.Labels[l] }

// tavernHouseLimit is the original 1000 platinum in gold.
const tavernHouseLimit = 5000

// tavernGamble = ECL3 A501..A643, after "WHO WILL GAMBLE?" (opWho picks the
// gambler: the purse is shared): spin a top for
// a gold wager. 8-9 of a d10 (from 0) is green, the wager won; 3-7 blue,
// the wager back; 0-2 red, the wager lost. The script counted the selected
// member's platinum in a 16-bit word; here it is the party's gold.
// opTavernGamble is TavernGame(): the asks run from the next instruction
// on, and the game goes on at TavernEvent or ResetVisits.
func (e *Engine) opTavernGamble() {
	e.vm.PC++
	e.tavernGamble()
}

func (e *Engine) tavernGamble() {
	if !e.ruleset().Check(rules.Fit{Who: e.selected()}, nil).Yes { // [6C00] is 1 for an okay member
		e.scriptText(true, "NO ONE IS WILLING TO GAMBLE WITH SOMEONE WHO LOOKS THAT BAD.")
		e.vm.PC = uint16(e.scriptLabel("TavernEvent"))
		return
	}
	e.tavernRound()
}

// tavernRound is one spin, each question asked of the player in turn
// (the asks complete on the answers; RunECL waits for the last).
func (e *Engine) tavernRound() {
	e.scriptText(true, "YOU HAVE ")
	e.scriptText(false, strconv.FormatUint(e.gold(), 10))
	e.scriptText(false, " GP.")
	e.askContinue(askDone{kind: dTavernWager})
}

// tavernWager: the wager is asked.
func (e *Engine) tavernWager() {
	e.scriptText(true, "HOW MANY GOLD PIECES WILL YOU WAGER? (HOUSE LIMIT IS "+strconv.Itoa(tavernHouseLimit)+")")
	e.askNumber(askDone{kind: dTavernSpun})
}

// tavernSpun: the wager (typed s) is spun and told.
func (e *Engine) tavernSpun(s string) {
	n, _ := strconv.ParseInt(strings.TrimLeft(s, " "), 10, 32)
	wager := uint64(uint16(n))
	r := 0
	if tavernWagerOK(e.gold(), wager) {
		r = int(e.ruleset().Check(rules.RollRandom{N: 10}, nil).Value)
	}
	gold, text := tavernSpin(e.gold(), wager, r)
	e.do(rules.SetPurseGold{Gold: gold})
	e.scriptText(true, text)
	e.askContinue(askDone{kind: dTavernAsk})
}

// tavernAsk: the spin is told and had its key; go again?
func (e *Engine) tavernAsk() {
	e.scriptText(true, "WILL YOU GO AGAIN?")
	e.askChoice("", "~YES ~NO", false, false, askDone{kind: dTavernAgain})
	e.eclWait.ask.Esc, e.eclWait.ask.NoEsc = game.EscTo(1), false // NO: done gambling
}

// tavernAgain: the answer c (0: yes) to going again.
func (e *Engine) tavernAgain(c int) {
	e.clearPrompt()
	if c == 0 {
		e.tavernRound()
		return
	}
	e.vm.PC = uint16(e.scriptLabel("ResetVisits"))
}

func tavernWagerOK(gold, wager uint64) bool {
	return wager > 0 && wager <= gold && wager <= tavernHouseLimit
}

// tavernSpin is one spin of the top (r = Random(10)): the gold after it and
// the text.
func tavernSpin(gold, wager uint64, r int) (uint64, string) {
	switch {
	case !tavernWagerOK(gold, wager):
		return gold, "I CAN'T ACCEPT THAT WAGER."
	case r >= 8:
		return gold + min(wager, math.MaxUint64-gold), "IT'S GREEN.  YOU GET DOUBLE YOUR MONEY BACK."
	case r > 2:
		return gold, "IT'S BLUE.  YOU WIN."
	}
	return gold - wager, "RED. SORRY, YOU LOSE."
}
