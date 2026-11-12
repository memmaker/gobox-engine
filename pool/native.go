package pool

import (
	"fmt"
	"gobox/core"
	"math"
	"slices"
	"strconv"
	"strings"

	"gobox/ecl/script"
	"gobox/game/play"
)

// Native scripts (extra): stretches of the ECL scripts rewritten in Go,
// where the script's own logic cannot follow the port (one party purse in
// gold and copper). RunECL runs one instead of the commands at its label;
// it leaves vm.PC at the label the script would have gone on to.

type scriptAt struct {
	area  int    // the area script
	label string // its label
}

var nativeScripts = map[scriptAt]func(*Engine){
	{0, "TavernGamble"}: (*Engine).tavernGamble,
	{14, "CastKnock"}:   (*Engine).castKnock,
}

// The layout fields the engine names itself (layout/global.txt).
const (
	fieldScratch6 = "Scratch6" // castKnock's answer
	fieldHours    = "Hours"    // the day-sky option's comparison (opCompare)
)

// field is the running script's field called name.
func (e *Engine) field(name string) script.Ref {
	r, err := e.vm.Script.Layout.Named(name)
	if err != nil {
		panic(err) // the game's layout lacks a word the engine names: a broken game folder
	}
	return r
}

// castKnock is the cabinets' CAST KNOCK (area 14, Felix d3): Knock cast the
// native way (spendKnock), on at KnockOpens with the caster's place in the
// party in Scratch6, or at NoKnock. It was the Spell command and the script
// clearing the slot itself.
func (e *Engine) castKnock() {
	out, err := e.ruleset().Resolve(nil, core.SpendKnock{}, nil)
	if err != nil {
		panic(fmt.Sprintf("native: Knock: %v", err)) // the action never refuses
	}
	if len(out.Effects) == 0 {
		e.vm.PC = uint16(e.scriptLabel("NoKnock"))
		return
	}
	c := out.Effects[0].(core.KnockCast).Who
	e.host.Set(e.field(fieldScratch6), int64(uint16(slices.Index(e.partyIDs(), c))))
	e.vm.PC = uint16(e.scriptLabel("KnockOpens"))
}

// gold is the party's gold pieces.
func (e *Engine) gold() uint64 {
	v, err := e.ruleset().Fact(core.NoChar, core.PurseFact{})
	if err != nil {
		panic(fmt.Sprintf("native: purse: %v", err)) // the purse is always held
	}
	return v.(core.Purse).Gold
}

// purseWhos are the Who commands that spend the party's one purse: a label's
// instruction plus by.
var purseWhos = []struct {
	scriptAt
	by int
}{
	{scriptAt{0, "PayPlatinum"}, 0},   // the inn, ECL3 A1B4
	{scriptAt{0, "TavernGamble"}, -1}, // the tavern, ahead of A501
}

func (e *Engine) nativeScript() func(*Engine) { return e.native[int(e.vm.PC)] }

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
func (e *Engine) tavernGamble() {
	if !e.ruleset().Check(core.Fit{Who: e.selected()}, nil).Yes { // [6C00] is 1 for an okay member
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
		r = int(e.ruleset().Check(core.RollRandom{N: 10}, nil).Value)
	}
	gold, text := tavernSpin(e.gold(), wager, r)
	e.do(core.SetPurseGold{Gold: gold})
	e.scriptText(true, text)
	e.askContinue(askDone{kind: dTavernAsk})
}

// tavernAsk: the spin is told and had its key; go again?
func (e *Engine) tavernAsk() {
	e.scriptText(true, "WILL YOU GO AGAIN?")
	e.askChoice("", "~YES ~NO", false, false, askDone{kind: dTavernAgain})
	e.eclWait.ask.Esc, e.eclWait.ask.NoEsc = play.EscTo(1), false // NO: done gambling
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
