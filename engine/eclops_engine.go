package engine

import (
	"fmt"
	"gobox/core"
	"gobox/ecl"
	"gobox/ecl/script"
	"gobox/ega"
	"os"
	"gobox/rules1e"
	"slices"
	"strconv"
	"strings"
)

// RunECL = ovr04:3171 (addr): run the script from instruction addr until
// it stops, then ret (what follows is the caller's cont). The VM's state is
// the engine's (vm, es): an instruction that needs the player (an ask)
// or opens a screen (a fight, a shop, the party menu) leaves it suspended
// and RunECL returns; it goes on once that is answered or the screens are
// gone.
func (e *Engine) RunECL(addr int) {
	e.vm.PC = uint16(addr)
	e.vm.Stop = false
	e.eclRun()
}

// eclRun runs the script on from vm.PC (RunECL's loop).
func (e *Engine) eclRun() {
	for !e.vm.Stop && !e.run.quit && !e.stopsHere() {
		if int(e.vm.PC) >= len(e.vm.Script.Code) {
			fmt.Fprintf(os.Stderr, "script: no instruction %d\n", e.vm.PC)
			break
		}
		e.vm.LastCommand = e.vm.Command
		e.vm.Fetch()
		m, base, takes := e.ext.opened, len(e.k), e.es.takes
		if f := e.nativeScript(); f != nil {
			f(e)
		} else if op := e.vm.Ops[e.vm.Command]; op != nil {
			op(e.vm, e.host)
		} else {
			// Not ported yet: report it and end this script run.
			fmt.Fprintf(os.Stderr, "ecl: opcode %02X on line %d not ported\n", e.vm.Command, e.vm.Script.Code[e.vm.PC].Line)
			e.vm.PC++
			e.vm.Stop = true
		}
		if h := e.es.hold; h != holdNone { // the instruction opened camp or the party menu: the script goes on once it is closed
			e.es.hold = holdNone
			switch h {
			case holdPartyMenu:
				e.push(cont{kind: kEclResume}, cont{kind: kDrawMain})
				e.partyMenu(true, true)
			case holdEncamp: // the script waits for camp, then exits (the watch is only for a broken rest)
				e.push(cont{kind: kEclResume}, cont{kind: kProgramExit, n: [3]int{e.es.holdPC}})
				e.Encamp()
			}
			return
		}
		// An instruction asking the player completes on the answer. The ask
		// opens last, over anything else the instruction opened; no op opens
		// a read too (only scriptText's journal refs do, and no asking op
		// prints those), so the read never ends up under the ask. An op that
		// did would have to open its ask after the read, as the old stack did.
		if w := e.eclWait; w != nil && w.seq == 0 {
			e.openAsk()
		}
		if e.es.takes != takes { // the op took the script over: its conts end in the resume
			return
		}
		if e.ext.opened != m { // it opened a question, a read or a wait: the script goes on once they are done
			e.k = slices.Insert(e.k, base, cont{kind: kEclResume})
			if !e.blocked() {
				e.ret()
			}
			return
		}
	}
	e.vm.Stop = false // DS:442E is cleared on the way out
	e.ret()
}

// takeover is an op handing the script over to the conts it pushes next
// (and the flow they follow): the resume goes at the bottom of them.
func (e *Engine) takeover() {
	e.es.takes++
	e.push(cont{kind: kEclResume})
}

// opExit = ovr04:0001 (EXIT).
func (e *Engine) opExit() {
	if e.es.textInView != 0 {
		e.newPage()
		e.blank(1, 1, 15, 15)
		if e.worldWord(core.WordIn3D) != 0 {
			e.blank(3, 3, 13, 13)
		}
	}
	if e.run.restoreSel {
		e.selectChar(e.run.savedSel)
		e.run.restoreSel = false
	}
	e.es.encounterShown = [2]byte{}
	e.es.picShown, e.es.textInView = 0, 0
	e.vm.Stop = true
	e.vm.PC++
	e.vm.Calls = nil
	e.msgRow, e.msgCol = 0x11, 1
	e.run.textInBox = 0
}

// opCompare = ovr04:0126: strings when either operand is one.
func (e *Engine) opCompare() {
	e.vm.ReadArgs(2)
	if e.vm.ArgType[1] >= ecl.ArgText || e.vm.ArgType[2] >= ecl.ArgText {
		e.vm.SetTextFlags(e.vm.Text[1], e.vm.Text[2])
		return
	}
	a := e.vm.Arg(1, e.host)
	b := e.vm.Arg(2, e.host)
	hours, hasHours := e.layoutField(fieldHours)
	if hasHours && (e.vm.ArgType[1] == ecl.ArgVarByte || e.vm.ArgType[1] == ecl.ArgVarWord) && e.vm.Ref(1) == hours && e.vm.ArgType[2] == ecl.ArgImmediate && b == 14 && e.DaySky.Load() {
		// option (U3): the scripts pick the day sky by hour < 14; day is 6:00-20:00 instead.
		// Night is 15, not 14, so the pier's "Hours > 14" (no ship picture after dark) holds too.
		a = 15
		if e.ruleset().Check(core.IsDay{}, nil).Yes {
			a = 0
		}
	}
	e.vm.SetFlags(a, b)
}

// opLoadCharacter = ovr04:02E0: select party member n (bit 7: also refresh).
func (e *Engine) opLoadCharacter() {
	e.run.restoreSel = true
	e.vm.ReadArgs(1)
	v := byte(e.vm.Arg(1, e.host))
	refresh := v&0x80 != 0 && e.eclVersion.Behaviour.SelectMemberDrop
	n := int(v & 0x7F)
	e.es.memberMissing = n >= e.partySize()
	if !e.es.memberMissing {
		e.sel = n
	}
	if refresh && e.es.memberDied != 0 && e.es.memberGone != 0 {
		if e.run.savedSel == e.selected() {
			e.run.restoreSel = false
		}
		e.RemoveSelected(false, true) // ovr17:3291
		e.es.memberDied, e.es.memberGone = 0, 0
	}
}

// opLoadMonster = ovr04:0450 (monster, count, icon block): append count
// copies of the monster to the party list for the coming fight.
func (e *Engine) opLoadMonster() {
	e.vm.ReadArgs(3)
	if e.run.monCount >= 0x3F {
		return
	}
	e.addMonsters(int(e.vm.Arg(1, e.host)), int(byte(e.vm.Arg(2, e.host))), int(byte(e.vm.Arg(3, e.host))))
}

func (e *Engine) addMonsters(idx, count, icon int) {
	sel := e.sel
	if count == 0 {
		count = 1
	}
	r := e.ruleset()
	if _, err := r.Resolve(nil, core.NewGroup{}, nil); err != nil {
		panic(err)
	}
	if _, err := r.Spawn(core.MonsterID(idx), nil); err != nil { // ovr08:0411
		e.PromptRow24(ega.White, "Unable to load monster")
		e.Exit()
		return
	}
	tmplID := e.partyIDs()[e.partySize()-1]
	e.LoadIconPic("CPIC", icon, int(e.es.cloneIcon)) // ovr33:01C8
	e.meetMonster(e.ruleset().Sheet(tmplID).Name, e.monsterSet().File("CPIC"), icon)
	// The template itself is the first monster (ovr04:055B); each clone is
	// a copy of it made by Spawn (its Order too).
	e.do(core.SetOrder{Who: tmplID, Order: int32(e.es.cloneIcon)})
	e.run.monCount++
	for i := 2; i <= count && e.run.monCount < 0x3F; i++ {
		if _, err := r.Spawn(core.MonsterID(idx), nil); err != nil {
			panic(err)
		}
		e.run.monCount++
	}
	e.es.cloneIcon++
	e.run.monstersLoaded = 1
	e.sel = sel
}

// opSetupMonster = ovr04:03B4 (sprite, max distance, picture).
func (e *Engine) opSetupMonster() {
	sides := e.eclVersion.Behaviour.SetupMonsterSides
	if sides {
		e.vm.ReadArgs(4)
		e.run.monsterSides = byte(e.vm.Arg(4, e.host))
	} else {
		e.vm.ReadArgs(3)
	}
	e.run.sprite = int(byte(e.vm.Arg(1, e.host)))
	e.setWorldWord(core.WordMaxDist, int(e.vm.Arg(2, e.host)))
	e.run.closeup = int(byte(e.vm.Arg(3, e.host)))
	e.setWorldWord(core.WordMonsterDist, int(e.FreeSteps(e.posX, e.posY, e.facing)))
	if sides && e.worldWord(core.WordIn3D) == 0 { // outdoors the distance asked becomes 2 (pod O055:03D7)
		e.setWorldWord(core.WordMaxDist, maxEncounterDist)
	}
	if e.worldWord(core.WordMaxDist) < e.worldWord(core.WordMonsterDist) {
		e.setWorldWord(core.WordMonsterDist, int(e.worldWord(core.WordMaxDist)))
	}
	e.ShowEncounter(e.run.sprite, e.run.closeup, int(byte(e.worldWord(core.WordMonsterDist))), &e.es.encounterShown)
}

// maxEncounterDist: an encounter is at most two squares away (FreeSteps).
const maxEncounterDist = 2

// opApproach = ovr04:0817: the monster comes one step closer.
func (e *Engine) opApproach() {
	if d := e.worldWord(core.WordMonsterDist); d > 0 {
		e.setWorldWord(core.WordMonsterDist, int(d-1))
		e.ShowEncounter(e.run.sprite, e.run.closeup, int(byte(e.worldWord(core.WordMonsterDist))), &e.es.encounterShown)
	}
	e.vm.PC++
}

// opPicture = ovr04:0856 (block): show a picture, or with 0xFF put the
// view back.
func (e *Engine) opPicture() {
	e.vm.ReadArgs(1)
	v := int(byte(e.vm.Arg(1, e.host)))
	if e.worldWord(core.WordIn3D) == 0 && v != 0xFF {
		e.blank(3, 3, 13, 13)
	}
	if v != 0xFF {
		e.es.encounterShown[1] = 1
		e.es.picShown = 1
		if face := e.worldWord(core.WordPortrait); face == 0xFF {
			e.LoadAnimPic("PIC", v, false)
			e.DrawAnimFirst()
		} else {
			e.ShowPortrait(int(face), v)
		}
		return
	}
	if e.prevViewMode > 1 || e.viewMode == 1 {
		if e.es.picShown != 0 || e.es.spriteLoaded != 0 {
			e.Redraw()
			e.es.picShown, e.es.spriteLoaded = 0, 0
			e.es.picAnimating = 1
		}
	}
	e.es.encounterShown = [2]byte{}
}

// opInputNumber = ovr04:0951.
func (e *Engine) opInputNumber() {
	e.vm.ReadArgs(2)
	dst := e.vm.Ref(2)
	e.askNumber(askDone{kind: dNumber, dst: dst})
}

// opInputString = ovr04:0994.
func (e *Engine) opInputString() {
	e.vm.ReadArgs(2)
	dst := e.vm.Ref(2)
	max := textCols
	if e.eclVersion.Behaviour.InputTextOperand {
		max = int(byte(e.vm.Arg(1, e.host)))
	}
	e.askLine(max, false, askDone{kind: dString, dst: dst})
}

// opPrint = ovr04:0A09 (PRINT 0x11, PRINTCLEAR 0x12).
func (e *Engine) opPrint() {
	e.vm.ReadArgs(1)
	if e.vm.ArgType[1] < 0x80 {
		e.vm.Text[1] = strconv.Itoa(int(e.vm.Arg(1, e.host))) // 00BA:1073
	}
	e.scriptText(e.cmdIs(script.OpPrintClear), e.vm.Text[1])
}

// scriptText is opPrint's text: clear starts the text window over.
func (e *Engine) scriptText(clear bool, s string) {
	e.run.textCleared = 0
	orig := s
	s = platinumPrices.Replace(s)
	e.vm.Text[1] = s
	if clear {
		e.msgRow, e.msgCol = 0x11, 1
	}
	e.paneText(ega.LightGreen, clear, s)
	e.speak(voiceKey(orig)) // the narration was recorded from the scripts' own text
	if clear {
		e.ext.printed = ""
	}
	e.ext.printed += s
}

// opVMenu = ovr04:0F08 (result, prompt, count, items...): a vertical list.
func (e *Engine) opVMenu() {
	e.run.textCleared = 0
	e.vm.ReadArgs(3)
	dst := e.vm.Ref(1)
	prompt := e.vm.Text[1]
	n := int(byte(e.vm.Arg(3, e.host)))
	e.vm.PC--
	e.vm.ReadArgs(n)
	list := newList(n)
	e.msgCol, e.msgRow = 1, 0x11
	e.paneText(ega.LightGreen, true, prompt)
	for i, p := range list {
		s := e.strAt(i + 1)
		if len(s) > 0x28 {
			s = s[:0x28]
		}
		p.Text, p.Flag = s, 0
	}
	e.askList(list, askDone{kind: dList, dst: dst})
	e.eclWait.ask.Text = prompt // (extra) the remaster shows it over the list, not in a text box
	e.eclWait.ask.NoEsc = true
	e.scriptExit()
}

// listAnswered completes VMenu: line i of list is the result.
func (e *Engine) listAnswered(dst script.Ref, i int) {
	e.host.Set(dst, int64(uint16(i)))
	e.clearPane()
}

// opNewECL = ovr04:0BD1 (area): switch to another area's script.
func (e *Engine) opNewECL() {
	e.vm.ReadArgs(1)
	e.setLastBlock(int(e.gs.areaKind))
	v := byte(e.vm.Arg(1, e.host))
	e.gs.areaKind = v
	e.eclFile = e.set.File("ECL") + ".dax" // DS:547C
	// ovr18:22D5 asked for another game disk when the file was missing; all
	// files are in one folder here.
	e.LoadArea(int(v))
	e.InitAreaScript()
	e.vm.Stop = true
	e.run.areaChanged = true
	e.run.scriptLoad = 0
}

// opLoadArea = pod O040:0C58 (map, walls): in a 3D area the GEO map, and the walls
// unless they are already loaded (255: none; the map from 127 on: none).
func (e *Engine) opLoadArea() {
	e.vm.ReadArgs(2)
	e.run.scriptLoad, e.run.geoSet, e.run.walldefSet = 1, 1, 1
	geo, walls := int(byte(e.vm.Arg(1, e.host))), int(byte(e.vm.Arg(2, e.host)))
	if geo < geoNone && e.worldWord(core.WordIn3D) != 0 {
		e.setWorldWord(core.WordGeo, geo)
		e.LoadGeo(geo)
		e.setWorldWord(core.WordHoldMove, 0)
	}
	if walls != wallsNone && walls != e.mapPairA[0] {
		e.LoadWalldef(walls, 1)
	}
}

// LoadArea's operands that load nothing.
const (
	geoNone   = 127 // from it on: no map
	wallsNone = 255
)

// opLoadFiles = ovr04:0C74 (0x21 LOAD FILES, 0x37 LOAD PIECES).
func (e *Engine) opLoadFiles() {
	e.vm.ReadArgs(3)
	e.run.scriptLoad = 1
	var a [4]int
	for i := 1; i <= 3; i++ {
		a[i] = int(byte(e.vm.Arg(i, e.host)))
	}
	if e.cmdIs(script.OpLoadMap) {
		e.run.geoSet = 1
		if a[1] != 0xFF && a[1] != 0x7F && e.worldWord(core.WordIn3D) != 0 {
			e.setWorldWord(core.WordGeo, int(a[1]))
			e.LoadGeo(a[1])
			e.setWorldWord(core.WordHoldMove, 0)
		}
		if a[3] != 0xFF && e.worldWord(core.WordIn3D) == 0 {
			e.layAreaTiles() // ovr33:0022
		}
	} else {
		e.run.walldefSet = 1
		switch {
		case a[1] == 0x7F:
			e.LoadWalldef(0, 1)
		case e.worldWord(core.WordWallSplitA) != 0 && e.worldWord(core.WordWallSplitB) != 0:
			e.LoadWalldef(a[1], 1)
			e.LoadWalldef(a[3], 3)
		default:
			for i := 1; i <= 3; i++ {
				if a[i] != 0xFF {
					e.LoadWalldef(a[i], i)
				} else {
					e.mapPairA[i-1], e.mapPairB[i-1] = -1, -1
				}
			}
		}
	}
	if e.run.needFrame != 0 && e.run.walldefSet != 0 && e.run.geoSet != 0 {
		e.DrawMainScreen()
		e.run.needFrame = 0
	}
}

// opGetTable = ovr04:0E5E (base, index, dst).
func (e *Engine) opGetTable() {
	e.vm.ReadArgs(3)
	i := byte(e.vm.Arg(2, e.host))
	e.host.Set(e.vm.Ref(3), e.host.Get(e.vm.Ref(1).Plus(int(i))))
}

// opHMenu = ovr04:10D1 (result, count, items...): a one-line hotkey menu.
func (e *Engine) opHMenu() {
	e.vm.ReadArgs(2)
	dst := e.vm.Ref(1)
	n := int(byte(e.vm.Arg(2, e.host)))
	e.vm.PC--
	e.vm.ReadArgs(n)
	allowEnter := n == 1
	if n == 1 {
		if e.vm.Text[1] == "PRESS <RETURN> OR BUTTON TO CONTINUE" {
			e.vm.Text[1] = "PRESS <ENTER>/<RETURN> TO CONTINUE"
		}
	}
	anim := e.es.picShown != 0 && e.es.picAnimating != 0
	menu := ""
	for i := 1; i <= n-1; i++ {
		menu += "~" + e.strAt(i) + " "
		menu = menu[:min(len(menu), 0x32)]
	}
	menu += "~" + e.strAt(n)
	menu = menu[:min(len(menu), 0x32)]
	e.askChoice("", menu, allowEnter, anim, askDone{kind: dChoiceVar, dst: dst})
	e.scriptExit()
}

// opClearMonsters = ovr04:125D.
func (e *Engine) opClearMonsters() {
	e.vm.PC++
	e.clearTreasure()
}

// clearTreasure is ClearMonsters' work (pod's clearTreasure too): no monsters, no loot.
func (e *Engine) clearTreasure() {
	e.run.monCount = 0
	e.run.monstersLoaded = 0
	e.es.cloneIcon = 8
	e.do(core.ClearLoot{})
}

// opPartyStrength = ovr04:12C6.
func (e *Engine) opPartyStrength() {
	e.vm.ReadArgs(1)
	total := e.ruleset().Check(core.PartyStrength{Party: e.partyIDs()}, nil).Value
	e.host.Set(e.vm.Ref(1), int64(uint16(total)))
}

// opCheckParty = ovr04:1435 (what, special, min, max, avg, count): a party
// census. what 0 looks for a special ability, 0x6BA7 averages +0x79 and
// 0x6C1B averages +0x11C. The fourth result is the count for every game
// (S-diffs 12, sotsb's form): the members with the ability, or the members
// counted; found is count > 0. POR's original wrote found there (its two
// scripts are patched at import, ScriptFixes).
func (e *Engine) opCheckParty() {
	e.vm.ReadArgs(6)
	what := censusNone
	var fact core.Fact
	switch {
	case e.vm.ArgType[1] != 1:
		if e.vm.Arg(1, e.host) == 0 {
			what = censusSpecial
		}
	default:
		for name, f := range censusFacts {
			if r, err := e.vm.Script.Layout.Named(name); err == nil && e.vm.Ref(1) == r { // a word the game's layout lacks is no census
				what, fact = censusFact, f
			}
		}
	}
	special := byte(e.vm.Arg(2, e.host))
	var dst [7]script.Ref
	for i := 3; i <= 6; i++ {
		dst[i] = e.vm.Ref(i)
	}
	lo, hi, avg, count := byte(0xFF), byte(0), byte(0), 0
	store := func() { // ovr04:13D7
		e.host.Set(dst[3], int64(uint16(lo)))
		e.host.Set(dst[4], int64(uint16(hi)))
		e.host.Set(dst[5], int64(uint16(avg)))
		e.host.Set(dst[6], int64(uint16(count)))
	}
	census := func(f core.Fact) {
		party := e.partyIDs()
		a := e.ruleset().Check(core.Census{Party: party, Fact: f}, nil)
		lo, hi, avg, count = byte(a.Low), byte(a.High), byte(a.Avg), len(party)
		store()
	}
	switch what {
	case censusSpecial:
		for _, id := range e.partyIDs() {
			if e.ruleset().Check(core.HasAffect{Whom: []core.CharID{id}, Affect: core.ScriptAffect(special)}, nil).Yes { // 00BA:2199
				count++
			}
		}
		lo, hi, avg = 0, 0, 0
		store()
	case censusFact:
		census(fact)
	}
}

// opPartySurprise = ovr04:1629 (a, b): a = someone has +0x2F of 4 or 10;
// b is always 0.
func (e *Engine) opPartySurprise() {
	e.vm.ReadArgs(2)
	var a byte
	if e.ruleset().Check(core.HasClass{Party: e.partyIDs(), Class: rules1e.CoreClassRanger}, nil).Yes {
		a = 1
	}
	e.host.Set(e.vm.Ref(1), int64(uint16(a)))
	e.host.Set(e.vm.Ref(2), 0)
}

// opSurprise = ovr04:16C9 (a, b, c, d): roll who is surprised into
// DS:6DCB (0 none, 1 party, 2 monsters; the 3 "both" result is always
// overwritten by 2 in the original).
func (e *Engine) opSurprise() {
	e.vm.ReadArgs(4)
	var v [5]int
	for i := 1; i <= 4; i++ {
		v[i] = int(byte(e.vm.Arg(i, e.host)))
	}
	s1 := int(int8(byte(v[4] + 2 - v[1])))
	s2 := int(int8(byte(v[2] + 2 - v[3])))
	d1, d2 := e.rollDice(1, 6), e.rollDice(1, 6)
	r := 0
	if d1 <= s1 {
		if d2 <= s2 {
			r = 3
		} else {
			r = 1
		}
	}
	if d2 <= s2 {
		r = 2
	}
	e.setWorldWord(core.WordSurprise, int(r))
}

// opCombat = ovr04:179B: run the fight (or the shop/temple/after-combat
// screen when no fight is pending). The screens it puts up suspend the VM
// (eclRun) until they are gone.
func (e *Engine) opCombat() {
	e.vm.PC++
	e.takeover()
	e.push(cont{kind: kOpCombatEnd})
	if e.run.monstersLoaded == 0 && e.run.duelMode == 0 {
		if open := e.noMonsterScreen(); open != nil {
			open()
		} else {
			e.pushAfterCombat() // ovr06:1527
		}
		return
	}
	if e.eclVersion.Behaviour.AreaMonstersJoin {
		e.addAreaMonsters()
	}
	d := e.FreeSteps(e.posX, e.posY, e.facing)
	if d < e.worldWord(core.WordMonsterDist) {
		e.setWorldWord(core.WordMonsterDist, int(d))
	}
	e.push(cont{kind: kOpFought})
	e.pushCombat() // ovr09:007A
}

// noMonsterScreen is the screen Combat opens with no monsters: the dialect's first one
// whose word is set (it clears the word); nil: none.
func (e *Engine) noMonsterScreen() func() {
	for _, s := range e.eclVersion.Behaviour.NoMonsterCombat {
		w := combatScreenWords[s]
		if e.worldWord(w) != 1 {
			continue
		}
		e.setWorldWord(w, 0)
		switch s {
		case script.CombatShop:
			return e.Shop // ovr07:053E
		case script.CombatTemple:
			return e.Temple // ovr05:0D09
		case script.CombatVault:
			return e.Vault
		default:
			panic(fmt.Sprintf("ecl: Combat: no screen for %v", s))
		}
	}
	return nil
}

// combatScreenWords is the world word that asks for each screen.
var combatScreenWords = map[script.CombatScreen]core.WorldWord{
	script.CombatShop: core.WordShop, script.CombatTemple: core.WordTemple, script.CombatVault: core.WordVault}

// areaMonsterFor is the monster the area screen adds to a fight; none past the table.
func (t AreaMonsters) areaMonsterFor(screen int) (AreaMonster, bool) {
	if screen < 0 || screen >= len(t.ByScreen) {
		return AreaMonster{}, false
	}
	return t.ByScreen[screen], true
}

// addAreaMonsters = pod O055:1740: the area screen's own monsters join the fight, standing aside.
func (e *Engine) addAreaMonsters() {
	t := e.game().AreaMonsters
	m, ok := t.areaMonsterFor(e.worldWord(core.WordAreaScreen))
	if !ok {
		return
	}
	e.addMonsters(m.Monster, t.Count, m.Icon)
	ids := e.partyIDs()
	for _, id := range ids[len(ids)-t.Count:] {
		if err := e.ruleset().SetFact(id, core.FactStatus, core.Number(t.Status)); err != nil {
			panic("addAreaMonsters: " + err.Error())
		}
		control := e.fact(id, core.FactControl).(core.Number)
		if err := e.ruleset().SetFact(id, core.FactControl, control|core.Number(t.Control)); err != nil {
			panic("addAreaMonsters: " + err.Error())
		}
	}
}

// opFought: the fight is over; AfterCombat (ovr06:1527).
func (e *Engine) opFought() {
	e.push(cont{kind: kOpAfterFight})
	e.pushAfterCombat()
}

// opAfterFight: AfterCombat is over.
func (e *Engine) opAfterFight() {
	if e.worldWord(core.WordIn3D) == 0 {
		e.layAreaTiles()
		e.loadAreaIcon(0, 9)
		e.loadAreaIcon(1, 0xA)
		e.loadAreaIcon(2, 0xB)
	}
	e.run.duelMode = 0
	e.ret()
}

// opCombatEnd is opCombat's end, after its screens.
func (e *Engine) opCombatEnd() {
	if e.viewMode == 1 {
		e.setWorldWord(core.WordBeginFlag, 4)
	} else {
		e.setWorldWord(core.WordBeginFlag, 3)
	}
	e.setWorldWord(core.WordSearch, e.worldWord(core.WordSearch)&1)
	e.es.encounterShown = [2]byte{}
	e.es.picShown = 0
	e.DrawMainScreen()
	e.ret()
}

// strAt is text operand k (DS:6FC8 + 0x100*k), "" outside them.
func (e *Engine) strAt(k int) string {
	if k >= 0 && k < len(e.vm.Text) {
		return e.vm.Text[k]
	}
	return ""
}

// platinumPrices (extra) reprices the scripts' platinum in gold: the purse
// holds gold and copper, and a script's platinum is 5 of its gold.
var platinumPrices = strings.NewReplacer(
	"1 PLATINUM PIECE", "5 GOLD PIECES", // inn, harbor master (ECL3)
	"ENOUGH PLATINUM", "ENOUGH GOLD",
	// the training hall's notice (ECL11 ReadNotices): paladins and rangers train with the fighters
	"AND FIGHTERS AND 'ROGUES' TO GO EAST.", "AND FIGHTERS, PALADINS, RANGERS AND 'ROGUES' TO GO EAST.")

// census is what opCheckParty counts: its first operand, the address of
// the field whose value it averages (a variable operand) or 0 (a number).
type census int

const (
	censusNone    census = iota // another address or number: nothing stored
	censusSpecial               // 0: a special ability
	censusFact                  // a selected member's fact (censusFacts)
)

// censusFacts are the fields opCheckParty averages over the party, by the
// layout's name: the thief skills the game's scripts name and the movement
// (por SelFindTraps 0x6BA7, SelMove 0x6C1B; cotab's census covers all eight
// thief skills, cvThiefFirst..Last).
var censusFacts = map[string]core.Fact{
	"SelPickPockets":  core.SkillFact{Skill: rules1e.PickPockets},
	fieldSelFindTraps: core.SkillFact{Skill: rules1e.FindTraps},
	"SelThiefSkill6":  core.SkillFact{Skill: rules1e.Climb},
	fieldSelMove:      core.FactMove,
}

// eclOps are the commands the engine runs by name (the rest are
// ecl.Default's); initVM puts each at its dialect's byte code.
var eclOps map[byte]eclOp

// eclRun holds the ovr04 globals that are not part of the ovr08 interpreter.
type eclRun struct {
	restoreSel     bool        // DS:4430: put DS:4436 back as the selected character
	savedSel       core.CharID // DS:4436
	scriptLoad     byte        // DS:4431
	needFrame      byte        // DS:4432: main screen not drawn yet
	walldefSet     byte        // DS:4433
	geoSet         byte        // DS:4434
	monCount       byte        // DS:4435: monsters appended to the party list
	areaChanged    bool        // DS:442F
	quit           bool        // DS:49FF
	searchSave     byte        // DS:84DD
	textInBox      byte        // DS:84DE
	textCleared    byte        // DS:84E4
	monstersLoaded byte        // monsters were loaded; DS:84E7
	duelMode       byte        // DS:84D4
	sprite         int         // DS:6F81
	closeup        int         // DS:6F82
	doorTry        [3]byte     // DS:6DDE..6DE0: bash, pick, knock still allowed
	lineHasText    byte        // DS:84DF
	monsterSides   byte        // SetupMonster's sides (pod; bits 0..3): where the monsters come from in combat
}

func init() {
	eclOps = map[byte]eclOp{
		script.OpExit: (*Engine).opExit, script.OpCompare: (*Engine).opCompare,
		script.OpSelectMember: (*Engine).opLoadCharacter, script.OpLoadMonster: (*Engine).opLoadMonster,
		script.OpSetupMonster: (*Engine).opSetupMonster, script.OpApproach: (*Engine).opApproach, script.OpPicture: (*Engine).opPicture,
		script.OpInputNumber: (*Engine).opInputNumber, script.OpInputString: (*Engine).opInputString, script.OpPrint: (*Engine).opPrint,
		script.OpPrintClear: (*Engine).opPrint, script.OpVMenu: (*Engine).opVMenu,
		script.OpClearMonsters: (*Engine).opClearMonsters, script.OpPartyStr: (*Engine).opPartyStrength,
		script.OpCheckParty: (*Engine).opCheckParty, script.OpJournal: (*Engine).opJournal, script.OpNewECL: (*Engine).opNewECL,
		script.OpLoadMap: (*Engine).opLoadFiles, script.OpPartySurprise: (*Engine).opPartySurprise, script.OpSurprise: (*Engine).opSurprise,
		script.OpCombat: (*Engine).opCombat, script.OpGetTable: (*Engine).opGetTable, script.OpHMenu: (*Engine).opHMenu,
		script.OpLoadWalls: (*Engine).opLoadFiles, script.OpLoadArea: (*Engine).opLoadArea,
		script.OpQuest: (*Engine).opQuest, // the port's (gobox ecl V1)
		// ovr04 0x25..0x3D (eclops2_engine.go)
		script.OpRob: (*Engine).opRob, script.OpEncounterMenu: (*Engine).opEncounterMenu, script.OpParlay: (*Engine).opParlay,
		script.OpCall: (*Engine).opCall, script.OpDamage: (*Engine).opDamage, script.OpRedrawView: (*Engine).opRedrawView,
		script.OpTreasure: (*Engine).opTreasure, script.OpTreasureOf: (*Engine).opTreasureOf, script.OpItemSet: (*Engine).opItemSet,
		script.OpFindItem: (*Engine).opFindItem, script.OpFindAffect: (*Engine).opFindAffect, script.OpNewLine: (*Engine).opNewLine, script.OpClock: (*Engine).opClock,
		script.OpAddNPC: (*Engine).opAddNPC, script.OpProgram: (*Engine).opProgram, script.OpWho: (*Engine).opWho,
		script.OpDelay: (*Engine).opDelay, script.OpProtection: (*Engine).opProtection,
		script.OpClearBox: (*Engine).opClearBox,
		// the later dialects' (fold S2): SpriteOff is the first engine's 0x31
		script.OpSpriteOff: (*Engine).opRedrawView, script.OpNoOp: (*Engine).opNoOp, script.OpPressKey: (*Engine).opPressKey,
		script.OpYesNo: (*Engine).opYesNo, script.OpClearExit: (*Engine).opClearExit, script.OpGainExp: (*Engine).opGainExperience,
		script.OpSpell: (*Engine).opSpell, script.OpDestroy: (*Engine).opDestroyItems, script.OpDump: (*Engine).opDump,
		script.OpFindNPC: (*Engine).opFindNPC, script.OpSound: (*Engine).opSound, script.OpMenuKeys: (*Engine).opMenuKeys,
	}
}

// cmdIs reports whether the command being run is the canonical op.
func (e *Engine) cmdIs(op byte) bool { return e.vm.Command == e.vm.Raw[op] }

// eclOp is an instruction only this game runs, or runs its own way (the rest
// are ecl.Default's).
type eclOp func(e *Engine)
