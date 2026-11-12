package pool

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/ecl/script"
	"gobox/ega"
	"gobox/game/play"
	"slices"
)

// opTreasure = ovr04:19B0 (7 money amounts, items): items < 0x80 is an
// ITEM<n>.DAX block of 0x3F-byte records, 0xFF none, otherwise
// items-0x80 random items. Each lands at the head of the DS:6810 list.
func (e *Engine) opTreasure() {
	e.vm.ReadArgs(8)
	var coins [7]uint16
	for i := range coins {
		coins[i] = uint16(e.vm.Arg(i+1, e.host))
	}
	e.do(core.SetLoot{Coins: coins}) // DS:67F4
	v := int(byte(e.vm.Arg(8, e.host)))
	switch {
	case v < 0x80:
		if _, err := e.ruleset().Resolve(nil, core.AddTreasureBlock{Block: int32(v)}, nil); err != nil {
			e.PromptRow24(ega.White, "Unable to find item file")
			e.Exit()
		}
		return
	case v == 0xFF:
		return
	}
	e.do(core.AddRandomTreasure{N: int32(v - 0x80), Who: e.selected()})
}

// treasureOfCoins are TreasureOf's coin operands: platinum, gems, jewelry, the last three of the seven kinds.
const (
	treasureOfCoins = 3
	firstOfCoins    = 4 // platinum's place among the seven
)

// opTreasureOf = pod O040:13E0 (platinum, gems, jewelry, count, items...): the coins add to the
// loot, the ITEM0 library's items (255: a random scroll) go to the head of the treasure.
func (e *Engine) opTreasureOf() {
	e.vm.ReadArgs(treasureOfCoins + 1)
	var coins [7]uint16
	for k := range treasureOfCoins {
		coins[firstOfCoins+k] = uint16(e.vm.Arg(k+1, e.host))
	}
	e.do(core.AddLoot{Coins: coins})
	n := int(byte(e.vm.Arg(treasureOfCoins+1, e.host)))
	e.vm.PC--
	e.vm.ReadArgs(n)
	ids := make([]int32, n)
	for k := range ids {
		ids[k] = int32(byte(e.vm.Arg(k+1, e.host)))
	}
	e.do(core.AddLibraryItems{IDs: ids})
}

// opItemSet = pod O040:251F (set): the set's library items, their names known, are all the
// treasure; the shop then sells them. A set out of range is set 1 (GameTables.TreasureSets).
func (e *Engine) opItemSet() {
	e.vm.ReadArgs(1)
	e.clearTreasure()
	sets := e.game().TreasureSets
	n := int(byte(e.vm.Arg(1, e.host)))
	if n < 1 || n >= len(sets) {
		n = 1
	}
	e.do(core.AddLibraryItems{IDs: sets[n], Known: true})
	e.takeover()
	e.push(cont{kind: kOpCombatEnd})
	e.Shop()
}

// opRob = ovr04:1DB1 (who, pct, items): take pct% of the money and items%
// of the items; who 0 is the selected member (and their share of the
// purse), else everyone. Operand 3 is the per-item chance, as the fixed
// Pascal EclRob (the original read operand 2 again).
func (e *Engine) opRob() {
	e.vm.ReadArgs(3)
	who := byte(e.vm.Arg(1, e.host))
	pct := int(byte(e.vm.Arg(2, e.host)))
	itemPct := int(byte(e.vm.Arg(3, e.host)))
	robbed := core.Party{e.selected()}
	if who != 0 {
		robbed = e.partyIDs()
	}
	e.setWorldWord(core.WordWho, e.sel)
	e.do(core.Rob{Robbed: robbed, Pct: int32(pct), ItemPct: int32(itemPct)})
}

// opEncounterMenu = ovr04:1EF8 (sprite, max dist, picture, result, 5
// reactions, 3 texts, flee threshold, monster strength): the approach
// menu. result: 0 none, 1 combat, 2 party flees, 3 parlay.
func (e *Engine) opEncounterMenu() {
	e.es.inEncounterMenu = 1
	e.songMusic(songEncounter, 0) // extra
	e.run.textCleared = 0
	mr := e.ruleset().Check(core.MoveExtremes{}, nil)
	hi, lo := int(mr.High), int(mr.Low)
	e.vm.ReadArgs(14)
	e.run.sprite = int(byte(e.vm.Arg(1, e.host)))
	e.setWorldWord(core.WordMaxDist, int(e.vm.Arg(2, e.host)))
	e.run.closeup = int(byte(e.vm.Arg(3, e.host)))
	dst := e.vm.Ref(4)
	var react [5]byte
	for i := range react {
		react[i] = byte(e.vm.Arg(i+5, e.host))
	}
	var texts [4]string
	for i := 1; i <= 3; i++ {
		texts[i] = e.strAt(i)
	}
	e.es.enc = &encounter{dst: dst, react: react, texts: texts, hi: hi, lo: lo,
		fleeAt: int(byte(e.vm.Arg(13, e.host))), strength: int(byte(e.vm.Arg(14, e.host)))}
	e.setWorldWord(core.WordMonsterDist, int(e.FreeSteps(e.posX, e.posY, e.facing)))
	if e.worldWord(core.WordMaxDist) < e.worldWord(core.WordMonsterDist) {
		e.setWorldWord(core.WordMonsterDist, int(e.worldWord(core.WordMaxDist)))
	}
	e.encShow()
	e.encRound()
}

func (e *Engine) encShow() {
	e.ShowEncounter(e.run.sprite, e.run.closeup, int(byte(e.worldWord(core.WordMonsterDist))), &e.es.encounterShown)
}

func (e *Engine) encApproach() bool {
	if e.worldWord(core.WordMonsterDist) == 0 {
		return false
	}
	e.setWorldWord(core.WordMonsterDist, e.worldWord(core.WordMonsterDist)-1)
	e.encShow()
	return true
}

func (e *Engine) encWaitMsg() {
	e.paneText(ega.LightGreen, true, "Both sides wait.")
}

func (e *Engine) encFlee() {
	e.host.Set(e.es.enc.dst, 0)
	e.msgCol, e.msgRow = 1, 0x11
	e.paneText(ega.LightGreen, true, "The monsters flee.")
}

// encRound shows the monsters' text for the distance and asks the menu
// (again after WAIT or a step closer).
func (e *Engine) encRound() {
	en := e.es.enc
	dist := func() int { return e.worldWord(core.WordMonsterDist) }
	txt := ""
	anim := e.es.picShown != 0 && e.es.picAnimating != 0
	clear := true
	e.msgCol, e.msgRow = 1, 0x11
	pick := func(k, stop int) {
		for {
			txt = en.texts[k]
			k++
			if k > 3 {
				k = 1
			}
			if txt != "" || k == stop {
				return
			}
		}
	}
	switch dist() {
	case 0:
		for k := 1; ; {
			txt = en.texts[k]
			k++
			if txt != "" || k > 3 {
				break
			}
		}
	case 1:
		pick(2, 2)
	case 2:
		pick(3, 3)
	}
	if txt == "" {
		clear = false
	}
	e.paneText(ega.LightGreen, clear, txt)
	en.parlay = dist() == 0 || e.worldWord(core.WordIn3D) == 0
	menu := "~COMBAT ~WAIT ~FLEE ~ADVANCE"
	if en.parlay {
		menu = "~COMBAT ~WAIT ~FLEE ~PARLAY"
	}
	e.askChoice("", menu, false, anim, askDone{kind: dEncounter})
}

// encounterAnswered goes on with the menu's choice: a result for the
// script, or the menu again.
func (e *Engine) encounterAnswered(choice int) {
	en := e.es.enc
	dst := en.dst
	if en.parlay && choice == 3 {
		choice = 4
	}
	if choice < 0 || choice > 4 {
		e.encRound()
		return
	}
	again := false
	// advance: PARLAY at distance 0, else one step closer.
	advance := func() {
		if e.worldWord(core.WordMonsterDist) == 0 {
			e.host.Set(dst, 3)
			return
		}
		e.encApproach()
		again = true
	}
	switch en.react[choice] {
	case 0:
		if choice == 2 && en.lo >= en.fleeAt {
			e.host.Set(dst, 2)
		} else {
			e.host.Set(dst, 1)
		}
	case 1:
		switch choice {
		case 0:
			e.host.Set(dst, 1)
		case 1:
			again = true
			e.encWaitMsg()
		case 2:
			e.host.Set(dst, 2)
		case 3:
			if !e.encApproach() {
				e.encWaitMsg()
			}
			again = true
		case 4:
			advance()
		}
	case 2:
		if choice == 0 {
			if en.strength > en.hi {
				e.encFlee()
			} else {
				e.host.Set(dst, 1)
			}
		} else {
			e.encFlee()
		}
	case 3:
		switch choice {
		case 0:
			e.host.Set(dst, 1)
		case 1, 3:
			if !e.encApproach() {
				e.encWaitMsg()
			}
			again = true
		case 2:
			e.host.Set(dst, 2)
		case 4:
			advance()
		}
	case 4:
		switch choice {
		case 0:
			e.host.Set(dst, 1)
		case 1, 3, 4:
			advance()
		case 2:
			e.host.Set(dst, 2)
		}
	}
	if again {
		e.encRound()
		return
	}
	e.clearPrompt()
	e.es.inEncounterMenu = 0
}

// opParlay = ovr04:25CB (5 results, dst): the talk-style menu.
func (e *Engine) opParlay() {
	e.vm.ReadArgs(6)
	var r [5]byte // the original keeps the results as bytes
	for i := range r {
		r[i] = byte(e.vm.Arg(i+1, e.host))
	}
	dst := e.vm.Ref(6)
	e.askChoice(" ", "~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE", false, false, askDone{kind: dParlay, dst: dst, r: r})
}

// opCall = ovr04:2CCF (addr): runs the engine routine the dialect names
// for the operand (script.Version.Calls).
func (e *Engine) opCall() {
	e.vm.ReadArgs(1)
	switch e.eclVersion.Calls[uint16(e.vm.ArgN[1])] { // a routine number, not memory
	case script.CallRedrawView:
		e.redrawChanged()
	case script.CallDuelWin:
		e.DuelSetup(1) // ovr08:1A6F
	case script.CallDuelLose:
		e.DuelSetup(0)
	case script.CallDuelSound:
		switch e.es.soundVar {
		case 8:
			e.cue(core.CueStep)
		case 0xA:
			e.cue(core.CueFireball) // DS:262A
		}
	case script.CallStepForward:
		e.StepForward()
	case script.CallWallAhead:
		if e.viewMode == 1 {
			e.es.wallAhead = e.WallAt(e.posX, e.posY, e.facing)
		}
	case script.CallPause: // cotab gameDelay, sotsb pauseForMessage
		e.askPause(int(e.msgDelay)*0xE1, askDone{})
	}
	// ponytail: CallNextFrame (pod's picture frame step) waits for pod's
	// switch to pool, where its picture frames live.
}

// redrawChanged is CALL's RedrawView: the cell under the party again, and
// the view when the script changed it.
func (e *Engine) redrawChanged() {
	e.es.cell = e.MapCell(e.posX, e.posY)
	s := &e.es
	if e.run.scriptLoad != 0 && (s.picShown != 0 || s.spriteLoaded != 0 || s.skyChanged != 0 || s.posChanged != 0 || s.portraitChanged != 0) {
		e.Redraw()
		s.portraitChanged, s.skyChanged, s.posChanged, s.picShown, s.spriteLoaded = 0, 0, 0, 0, 0
	}
}

// opNoOp: a command the dialect's engine reads and does nothing for.
func (e *Engine) opNoOp() { e.vm.PC++ }

// opPressKey = pod O040:1641: wait for a key.
func (e *Engine) opPressKey() {
	e.vm.PC++
	e.askContinue(askDone{})
}

// opYesNo = pod O040:1677: equal when the answer is Yes.
func (e *Engine) opYesNo() {
	e.vm.PC++
	e.askYesNo(&play.Question{YesNo: true, Default: 1, Replies: []string{"No", "Yes"}, NoEsc: true}, false, askDone{kind: dYesNo})
}

// opClearExit = pod O040:24AE: the view back, then Exit.
func (e *Engine) opClearExit() {
	e.redrawChanged()
	e.opExit()
}

// opGainExperience = sotsb O050:37F4 (whole party, points): the points go
// to everyone (or the selected member), split among each one's classes.
// Ruling S8: as given (pod's are multiplied by 100 at import), any member
// whatever its state.
func (e *Engine) opGainExperience() {
	e.vm.ReadArgs(2)
	whom := e.partyIDs()
	text := "Congratulations the party gains experience!"
	if byte(e.vm.Arg(1, e.host)) == 0 {
		whom = core.Party{e.selected()}
		text = "Congratulations " + e.name(e.selected()) + " gains experience!"
	}
	e.do(core.GainExperience{Whom: whom, Points: int32(uint16(e.vm.Arg(2, e.host)))})
	e.scriptText(true, text)
	e.askPause(int(e.msgDelay)*0xE1, askDone{})
}

// opSpell = sotsb O050:3410 (spell, slot, who): who in the party first has
// the spell memorized: its slot (from 0; spellNone when no one) and the
// member's place (the last member's when no one). Ruling S11: the one spell
// list's id, no range search; POR gains the handler (area14's one Spell runs as the engine's castKnock).
func (e *Engine) opSpell() {
	e.vm.ReadArgs(3)
	ids := e.partyIDs()
	a := e.ruleset().Check(core.Memorized{Party: ids, Spell: e.spellID(int(byte(e.vm.Arg(1, e.host))))}, nil)
	slot, who := int64(spellNone), max(len(ids)-1, 0)
	if a.Yes {
		slot, who = int64(a.Value), slices.Index(ids, a.Who)
	}
	e.host.Set(e.vm.Ref(2), slot)
	e.host.Set(e.vm.Ref(3), int64(who))
}

// spellNone is Spell's slot result when no one has the spell.
const spellNone = 255

// DestroyItems' operand: the item type, +128 for magical items only.
const (
	destroyMagicalOnly = 128
	destroyTypeMask    = 127
)

// opDestroyItems = pod O040:2066 (type): the party's items of the type go,
// readied ones put away first. Ruling S9: unready always, pod's +128
// magical-only flag, for all games.
func (e *Engine) opDestroyItems() {
	e.vm.ReadArgs(1)
	v := int(byte(e.vm.Arg(1, e.host)))
	id, err := e.itemID(v & destroyTypeMask)
	if err != nil { // an item type without an id: no one carries it (as opFindItem)
		return
	}
	e.do(core.DestroyItems{Whom: e.partyIDs(), Item: id, MagicalOnly: v&destroyMagicalOnly != 0})
}

// opDump = cotab O03C:3249: the selected member leaves the party; the one
// selected after is the one Exit puts back.
func (e *Engine) opDump() {
	e.vm.PC++
	e.RemoveSelected(false, true)
	e.run.savedSel = e.selected()
}

// DAMAGE's first operand (ruling 3, cotab's): a count of random members,
// or flags; its save operand: the save type, or with saveSelected the
// selected member saves (type 0: none).
const (
	damageFlagged   = 0x80
	damageAll       = 0x40
	damageAnyway    = 0x20 // everyone, no saving throw
	damageSaveHurts = 0x10 // a made save still hurts
	damageBonusMask = 0x1F // the save bonus
	saveSelected    = 0x80
	saveTypeMask    = 7
)

// opDamage = ovr04:2775 (flags, dice, sides, bonus, save): damage against
// one member, everyone (flags 0x40) or flags random members, as the rules'
// ScriptDamage (ruling 3: cotab's rule; 0x10 = a made save still hurts).
// With flags 0x80 a saving throw (type save&7, bonus flags&0x1F) avoids
// it; 0x20 means no save when hitting everyone.
func (e *Engine) opDamage() {
	sel := e.sel
	e.vm.ReadArgs(5)
	flags := byte(e.vm.Arg(1, e.host))
	save := byte(e.vm.Arg(5, e.host))
	d := core.ScriptDamage{
		Dice: int32(byte(e.vm.Arg(2, e.host))), Sides: int32(byte(e.vm.Arg(3, e.host))), Bonus: int32(byte(e.vm.Arg(4, e.host))),
		All: flags&damageAll != 0,
	}
	if e.partySize() > 0 {
		d.Selected = e.selected()
	}
	if flags&damageFlagged == 0 {
		d.Random, d.Count, d.HitBonus = true, int32(flags), int32(save)
	} else {
		d.SaveBonus, d.SaveHurts = int32(flags&damageBonusMask), flags&damageSaveHurts != 0
		typ := int(save & saveTypeMask)
		switch {
		case d.All && flags&damageAnyway != 0:
			typ = -1
		case !d.All && save&saveSelected != 0:
			d.BySelected = true
			typ-- // 0: no save
		}
		if typ >= 0 && typ < len(goldbox.SaveOrder) {
			d.Save = goldbox.SaveOrder[typ]
		}
	}
	o, err := e.ruleset().Resolve(core.Scripted{Party: e.partyIDs()}, d, nil)
	if err != nil {
		panic(fmt.Sprintf("script: Damage: %v", err))
	}
	for _, f := range o.Effects {
		if h, ok := f.(core.Hurt); ok {
			e.HitMessage(h)
		}
	}
	e.run.quit = true
	for _, id := range e.partyIDs() {
		if e.fact(id, core.FactActive) == core.Yes(true) {
			e.run.quit = false
		}
	}
	if e.run.quit {
		e.Music(5) // Amiga 0x416C2
		e.newPage()
		e.blank(1, 1, winRight, winBottom)
		e.endPage("The entire party is killed!", CellRect{X1: 1, Y1: 1, X2: winRight, Y2: winBottom})
	}
	e.sel = sel
	ms := 0
	if e.run.quit {
		ms = 3000
	}
	e.askPause(ms, askDone{kind: dPauseKey}) // was PromptRow24's any key
}

// opRedrawView = ovr04:2AA2.
func (e *Engine) opRedrawView() {
	e.vm.PC++
	if e.es.spriteLoaded != 0 {
		e.Redraw()
		e.es.spriteLoaded, e.es.picShown = 0, 0
	}
}

// opFindItem = ovr04:266A (item): eq when someone carries it.
func (e *Engine) opFindItem() {
	e.vm.ReadArgs(1)
	n := int(e.vm.Arg(1, e.host))
	e.vm.Flags = [6]bool{}
	e.vm.Flags[1] = true
	// An item type without an id is carried by no one: the original's
	// type compare (EclFindItem) finds nothing, so the flags stay "not found".
	id, err := e.itemID(n)
	if err != nil {
		return
	}
	if e.ruleset().Check(core.HasItem{Whom: e.partyIDs(), Item: id}, nil).Yes {
		e.vm.Flags[0], e.vm.Flags[1] = true, false
	}
}

// opNewLine = ovr04:2AEB: text cursor to the start of the next row.
func (e *Engine) opNewLine() {
	e.vm.PC++
	e.msgCol = 1
	e.msgRow++
	e.run.lineHasText = 0
}

// opClock = ovr04:2AC1 (n): advance the clock n minutes. (The original
// passes an uninitialised local as the operand number to ArgValue; operand 1
// is what the scripts mean.)
//
// Ruling S6 (Felix 2026-11-11): one handler, count + unit; the first
// engine's one-operand form is unit = minutes.
func (e *Engine) opClock() {
	n := e.eclVersion.Cmds[e.vm.Command].Fixed
	e.vm.ReadArgs(n)
	unit := int32(clockMinutes)
	if n > 1 {
		unit = int32(int8(byte(e.vm.Arg(2, e.host))))
	}
	e.as(e.sel)
	e.do(core.AdvanceClock{Count: int32(byte(e.vm.Arg(1, e.host))), Unit: unit})
}

// clockMinutes is CLOCK's unit for minutes (the clock's field 1).
const clockMinutes = 1

// opAddNPC = ovr04:2B7E (monster, morale).
func (e *Engine) opAddNPC() {
	e.vm.ReadArgs(2)
	v := int(byte(e.vm.Arg(1, e.host)))
	n := e.partySize()
	e.AddMonsterToParty(v)  // ovr18:1262
	if e.partySize() == n { // fixed, as the Pascal reference (EclAddNpc): a full party took no one; the original made the selected member the NPC (Fatima in area13)
		return
	}
	id := e.selected()
	if e.game().AddNPCIcon {
		e.LoadIconPic("CPIC", v, e.partyIndex(id))
	}
	if err := e.ruleset().SetFact(id, core.FactControl, core.Number(byte(e.vm.Arg(2, e.host))>>1|0x80)); err != nil {
		panic("opAddNPC: " + err.Error())
	}
	side := core.SideParty
	if v == 0x18 {
		side = core.SideMonsters
	}
	if err := e.ruleset().SetFact(id, core.FactSideOnly, side); err != nil {
		panic("opAddNPC: " + err.Error())
	}
	e.applyItems(id) // 00BA:0BB8
}

// opProgram = ovr04:2E10 (n): the sub-command the dialect names (script.Version.Programs).
func (e *Engine) opProgram() {
	e.vm.ReadArgs(1)
	v := byte(e.vm.Arg(1, e.host))
	if e.run.restoreSel {
		e.selectChar(e.run.savedSel)
		e.run.restoreSel = false
	}
	switch e.eclVersion.Programs[v] { // the screens put up suspend the VM until they are gone
	case script.ProgramPartyMenu:
		e.es.hold = holdPartyMenu
	case script.ProgramEnding:
		e.ProgramEight() // ovr19:02A4
	case script.ProgramCamp:
		e.es.hold, e.es.holdPC = holdEncamp, int(e.vm.PC)
	case script.ProgramQuit: // cotab 3, pod 2 (O040:1F1F): back to DOS; the front ends the run (play.Ended), as pod's Exit
		e.event(play.Ended{})
	case script.ProgramSaveQuit: // pod 1 (O043:0000): the save screen (the player picks the slot), then the ending
		e.event(play.SaveOffered{})
		e.ProgramEight()
	}
}

// opWho = ovr04:2B33 (prompt): choose a party member.
func (e *Engine) opWho() {
	e.vm.ReadArgs(1)
	e.clearPane()
	start := e.selected()
	// extra: one party purse, so the first non-NPC pays (inn, ECL3 A1B4) or
	// gambles (tavern, A501), an okay one if there is one
	if e.purseWho[e.vm.Cur()] {
		pick := -1
		ids := e.partyIDs()
		active := func(id core.CharID) bool { return e.fact(id, core.FactActive) == core.Yes(true) }
		for i, id := range ids {
			if e.fact(id, core.FactControl).(core.Number) < 0x80 && (pick < 0 || active(id) && !active(ids[pick])) {
				pick = i
			}
		}
		if pick >= 0 {
			e.sel = pick
			return
		}
	}
	e.askMember(e.strAt(1), start, askDone{kind: dWho, who: start}) // 00BA:2C48 (var DS:5D92)
}

// opDelay = ovr04:271F.
func (e *Engine) opDelay() {
	e.vm.PC++
	e.askPause(int(e.msgDelay)*0xE1, askDone{})
}

// opProtection = ovr04:2E7D (base): the glyph string stored in the
// variables from base on, each value+1 a symbol id, until one is <= 1.
// Extra: the original draws the runes; the letters they spell are shown
// instead, one per rune.
func (e *Engine) opProtection() {
	e.vm.ReadArgs(1)
	e.es.encounterShown = [2]byte{}
	e.es.picShown = 0
	if e.run.textInBox != 0 && e.msgRow > 0xD {
		e.msgCol, e.msgRow = 2, 5
		e.es.textInView = 1
	}
	e.run.lineHasText = 1
	base := e.vm.Ref(1)
	for i := 0; ; i += 2 {
		v := byte(e.host.Get(base.Plus(i)) + 1)
		if v <= 1 {
			break
		}
		e.char(e.msgCol, e.msgRow, ega.White, runeLetter(v), true)
		e.msgCol++
	}
	e.msgRow++
}

// opClearBox = ovr04:2B14: blank the view window for text.
func (e *Engine) opClearBox() {
	e.vm.PC++
	e.snap.win = WinText
	e.blank(1, 1, 15, 15)
	e.run.textInBox = 1
}

// runeLetter is the letter an Espruar rune stands for: each rune is two
// glyphs, its first glyph id g. A..O start at 0x41, the later letters
// three glyphs further on (the Sokal Keep scroll: LUX, SAMOSUD, SHESTNI).
func runeLetter(g byte) byte {
	if g >= 0x60 {
		return 'A' + (g-0x44)/2
	}
	return 'A' + (g-0x41)/2
}

// encounter is the approach menu's state while it is asked: the
// instruction's operands and what the party's strength was.
type encounter struct {
	dst      script.Ref
	react    [5]byte
	texts    [4]string
	fleeAt   int
	strength int
	hi, lo   int
	parlay   bool // the menu offers PARLAY, not ADVANCE
}

// FindNPC's results (pod O040:1C84).
const (
	npcNotFound = 255 // no hired NPC of that name
	npcNotOkay  = 128 // added to the place: the NPC is not up
	npcControl  = 128 // FactControl from here on: an NPC
)

// opFindNPC = pod O040:1C84 (name, dst): the party place of the hired NPC
// named name (+npcNotOkay when not up), or npcNotFound.
func (e *Engine) opFindNPC() {
	e.vm.ReadArgs(2)
	name, r := e.strAt(1), int64(npcNotFound)
	for i, id := range e.partyIDs() {
		if e.name(id) == name && e.fact(id, core.FactControl).(core.Number) >= npcControl {
			r = int64(i)
			if e.fact(id, core.FactActive) != core.Yes(true) {
				r |= npcNotOkay
			}
			break
		}
	}
	e.host.Set(e.vm.Ref(2), r)
}

// opSound = pod O040:24CD (n): what the game's number n plays (GameTables.ScriptSounds).
func (e *Engine) opSound() {
	e.vm.ReadArgs(1)
	s, ok := e.game().ScriptSounds[byte(e.vm.Arg(1, e.host))]
	switch {
	case !ok:
	case s.Music:
		e.songMusic(s.Song, 0)
	default:
		e.cue(s.Cue)
	}
}

// menuKeysPrompt heads MenuKeys' menu (pod O040:0FE1, command 49).
const menuKeysPrompt = "what do you do?"

// opMenuKeys = pod's HMenu 49: "what do you do?", then the HMenu.
func (e *Engine) opMenuKeys() {
	e.scriptText(false, menuKeysPrompt)
	e.opHMenu()
}
