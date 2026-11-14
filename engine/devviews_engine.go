//go:build !nodev

// Dev views: the headless checks' scenes (-dev, -battle, -fight, -at,
// -ecltrace). Left out of a release build (`-tags nodev`;
// devviews_off.go stands in for the dispatch hooks there).
package engine

import (
	"fmt"
	"gobox/adventure"
	"gobox/asset"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"os"
)

// DevScene returns a routine that sets up one screen for a headless check
// (snap ... game SLOT KEYS SCENE) and puts it up, or nil.
func (e *Engine) DevScene(name string) func() {
	switch name {
	case "shop":
		return func() {
			e.fixture(rules.FixtureShop, rules.NoChar)
			if h, ok := e.artFolder().FirstID(asset.DOS, e.set.File("HEAD")); ok {
				if b, ok := e.artFolder().FirstID(asset.DOS, e.set.File("BODY")); ok {
					e.ShowPortrait(h, b) // a shopkeeper's both halves, as a shop's script shows them (no 0, no body block 0)
				}
			}
			e.push(cont{kind: kDevBack})
			e.Shop()
		}
	case "hall", "hallpoor": // the training hall's party menu (hallpoor: no gold for the fee)
		return func() {
			f := rules.FixtureHall
			if name == "hallpoor" {
				f = rules.FixtureHallPoor
			}
			e.fixture(f, rules.NoChar)
			e.push(cont{kind: kDrawMain})
			e.partyMenu(true, true)
		}
	case "temple":
		return func() {
			e.fixture(rules.FixtureTemple, e.selected())
			e.push(cont{kind: kDevBack})
			e.Temple()
		}
	case "appraise":
		return func() {
			e.fixture(rules.FixtureAppraise, rules.NoChar)
			e.push(cont{kind: kDevBack})
			e.Shop()
		}
	case "combat", "wildcombat", "castfight", "missile", "autocombat", "meleefight", "fleefight", "winfight", "losefight": // losefight: a foe of 250 hp next to every member of 1 hp (the party falls: the last blow's body, the monsters' page, the party menu); winfight: one foe of 1 hp next to the party (a hit wins, the round's end asks Continue Battle:); meleefight: a foe next to every member (a bump attack on the first turn); fleefight: every member on a square beside the map's edge (a step off it asks Flee:); castfight: every member has Cure Light Wounds memorized (the inventory's Spells in a fight); missile: items for everyone (a sling, a bow and arrows, javelins), so the bar shows Use; autocombat: the computer fights for the party against eight (measuring the battle view)
		return func() {
			if name == "missile" {
				e.fixture(rules.FixtureMissile, rules.NoChar)
				e.selectMember(0)
			}
			if name == "wildcombat" { // the overland's terrain, from the dungeon save
				e.viewMode = 2
			}
			if name == "castfight" {
				e.fixture(rules.FixtureCastFight, rules.NoChar)
				e.do(rules.SetPortrait{Who: e.selected(), PNG: e.DevPortrait}) // the sheet's portrait, when DevPortrait is set
			}
			// Block 1 exists only in MON2CHA; take the first monster of
			// the current set (DS:5376) so any save can start a fight.
			idx := 1
			if id, ok := e.gdata().firstMonster(e.monsterSet().File("MON")); ok {
				idx = id
			}
			if e.DevMonster > 0 { // -fight N
				idx = e.DevMonster
			}
			count := 3
			if name == "meleefight" || name == "losefight" {
				count = e.partySize() // a foe for every member
			}
			if name == "winfight" {
				count = 1
			}
			if name == "autocombat" {
				count = 8
				e.fixture(rules.FixtureAutoCombat, rules.NoChar)
			}
			e.addMonsters(idx, count, 1)
			// as opSetupMonster with max distance 1: the monsters stand off
			// a block (a fresh game's distance is 0, right among the party)
			e.setWorldWord(rules.WordMonsterDist, min(e.FreeSteps(e.posX, e.posY, e.facing), 1))
			open := e.pushCombat
			if name == "meleefight" || name == "fleefight" || name == "winfight" || name == "losefight" {
				open = func() {
					e.pushCombat()
					e.do(rules.DevPlace{Melee: name != "fleefight"})
					if name == "winfight" || name == "losefight" {
						e.fixture(map[string]rules.FixtureID{"winfight": rules.FixtureWinFight, "losefight": rules.FixtureLoseFight}[name], rules.NoChar)
					}
				}
			}
			e.push(cont{kind: kDevFought})
			open()
		}
	case "overland": // the wilderness map, the party on it
		return func() {
			e.setWorldWord(rules.WordBeginFlag, 4)
			e.setWorldWord(rules.WordIn3D, 0)
			e.viewMode = 2
			e.setWorldWord(rules.WordMapX, 7)
			e.setWorldWord(rules.WordMapY, 29)
			e.loadAreaIcon(0, 9)
			e.areaMapIcons()
		}
	case "bestiary": // camp with a few kinds met for its Bestiary (CPIC blocks 1..5)
		return func() {
			for i, n := range []string{"KOBOLD", "ORC", "TROLL", "GOBLIN GUARD", "WYVERN"} {
				e.meetMonster(n, e.monsterSet().File("CPIC"), i+1)
			}
			e.CampMenu()
		}
	case "parlay": // an encounter's menu, then the talk-style menu a script's Parlay shows after PARLAY
		return func() { // as opEncounterMenu at distance 0 with a close-up, then the script's Parlay
			sprite, pic := 1, 1
			for _, f := range []struct {
				name string
				id   *int
			}{{"SPRIT", &sprite}, {"PIC", &pic}} {
				if id, ok := e.artFolder().FirstID(asset.DOS, e.set.File(f.name)); ok {
					*f.id = id
				}
			}
			e.es.inEncounterMenu = 1
			e.setWorldWord(rules.WordPortrait, 0xFF)
			e.ShowEncounter(sprite, pic, 0, &e.es.encounterShown)
			e.msgCol, e.msgRow = 1, 0x11
			e.paneText(game.InkLightGreen, true, "A band of kobolds blocks the way.")
			e.push(cont{kind: kDrawMain})
			e.devParlayAnswered(-1)
			e.openAsk()
		}
	case "password": // a script's InputString: the question in the text window, then the line typed (the remaster's pop-up)
		return func() {
			e.scriptText(true, "WHAT DO YOU SAY? (TYPE A SINGLE WORD)")
			e.askLine(0x28, false, askDone{kind: dDevPassword})
			e.push(cont{kind: kDrawMain})
			e.openAsk()
		}
	case "portrait": // the DevPortrait (-png) as the selected member's portrait: the sheet, then the view with the cards
		return func() {
			e.do(rules.SetPortrait{Who: e.selected(), PNG: e.DevPortrait})
			e.openThen(false, cont{kind: kDevLoot})
		}
	case "sheetcast": // the sheet's Spells with the DevPortrait (-png) as the portrait: the list is a page of its own (no portrait overlay, no tall row)
		return func() {
			e.selectMember(e.firstOkay())
			e.do(rules.SetPortrait{Who: e.selected(), PNG: e.DevPortrait})
			e.fixture(rules.FixtureSheetCast, e.selected())
			e.openThen(false, cont{kind: kDrawMain})
		}
	case "cast", "castko": // Cast on the command bar with Cure Light Wounds memorized; castko with the selected member unconscious
		return func() {
			f := rules.FixtureCastKO
			if name == "cast" { // an able caster
				f = rules.FixtureCast
				if i := e.firstOkay(); i >= 0 {
					e.selectMember(i)
				}
			}
			e.fixture(f, e.selected())
		}
	case "who": // a member pick on the view, as a script's "Who will search?"
		return func() { e.askMember("Who will search?", rules.NoChar, askDone{kind: dWho}); e.openAsk() }
	case "effects":
		return func() { // spell effects for the View page: many on member 1, one on member 2
			e.fixture(rules.FixtureEffects, rules.NoChar)
		}
	case "journal": // a text window naming journal entries, as ECL3 block 0 prints them
		return func() {
			e.paneText(game.InkLightGreen, true, "YOU RECORD IT IN YOUR JOURNAL.")
			e.showEntry(e.entryNamed("TavernTale3"))
			e.push(cont{kind: kDevJournal})
			e.blockedRet()
		}
	case "talljournal": // a journal entry with a picture taller than wide (entry 37): Up and Down scroll it
		return func() {
			e.paneText(game.InkLightGreen, true, "IT BECOMES PART OF YOUR JOURNAL.")
			e.showEntry(e.entryNamed("JournalEntry37"))
		}
	case "runes": // the Sokal Keep elf's scroll (ECL1 block 21, 9BD7): opClearBox, three opProtection lines
		return func() {
			e.setWin(WinText) // as opClearBox
			e.clearView()
			e.run.textInBox = 1
			e.msgCol, e.msgRow = 2, 5
			for _, l := range [][]byte{{0x56, 0x57, 0x6b, 0x6c, 0x71, 0x72}, {0x67, 0x68, 0x40, 0x41, 0x58, 0x59, 0x5c, 0x5d, 0x67, 0x68, 0x6b, 0x6c, 0x46, 0x47}, {0x67, 0x68, 0x4e, 0x4f, 0x48, 0x49, 0x67, 0x68, 0x69, 0x6a, 0x5a, 0x5b, 0x50, 0x51}} {
				for i := 0; i < len(l); i += 2 {
					e.cursorWrite(game.InkWhite, runeLetter(l[i]+1))
				}
				e.msgCol, e.msgRow = 2, e.msgRow+1
			}
			e.keyWait(cont{kind: kNop}) // the scroll stays up until a key (a pause outside a fight never waits)
		}
	case "inventory": // member 1 fully equipped, a party inventory to scroll
		return func() {
			e.fixture(rules.FixtureInventory, rules.NoChar)
		}
	case "lockeddoor": // a locked door ahead that no one can pick (the Pick menu's failure)
		return func() {
			for d := 0; d < 8; d += 2 { // face a wall
				if e.WallAt(e.posX, e.posY, d) != 0 && rules.OnMap(e.posX+rules.StepX[d], e.posY+rules.StepY[d]) {
					e.facing = d
					break
				}
			}
			e.setDoor(e.posX, e.posY, e.facing, 2)
			e.setDoor(e.posX+rules.StepX[e.facing], e.posY+rules.StepY[e.facing], (e.facing+4)%8, 2)
			e.run.doorTry = [3]byte{1, 1, 1}
			e.fixture(rules.FixtureNoPicker, rules.NoChar)
			e.Redraw()
		}
	case "dual": // the training hall with a human fighter who may change class (newClass)
		return func() {
			e.fixture(rules.FixtureDual, rules.NoChar)
			e.DevScene("hall")()
		}
	case "camp", "resting": // camp up; resting: a member is hurt and the rest has begun (StopRest)
		return func() {
			e.CampMenu()
			if name == "resting" {
				e.fixture(rules.FixtureHurt, rules.NoChar)
				e.campRest(adventure.RestFix)
			}
		}
	case "sheet", "items", "usable", "paladin", "drop": // drop: items with a dagger to drop (Drop It?); the selected member's sheet; items: and its inventory over it; usable: with a potion to drink as the last item; paladin: a paladin of level 3 (its powers)
		return func() {
			if name == "paladin" || name == "drop" || name == "usable" {
				e.fixture(map[string]rules.FixtureID{"paladin": rules.FixturePaladin, "drop": rules.FixtureDrop, "usable": rules.FixtureUsable}[name], e.selected())
			}
			e.OpenSheet(-1)
			if name != "sheet" && name != "paladin" {
				e.OpenItems(-1, false)
			}
		}
	case "learn": // the training hall with the hall's magic-user ready to rise (Train, then Learn)
		return func() {
			e.fixture(rules.FixtureLearn, rules.NoChar)
			e.sel = 4
			e.DevScene("hall")()
		}
	case "loot", "lootdetect": // lootdetect: the selected member has Detect Magic memorized
		return func() {
			if name == "lootdetect" {
				e.fixture(rules.FixtureLootDetect, e.selected())
			}
			e.fixture(rules.FixtureLoot, rules.NoChar)
			e.push(cont{kind: kDevFightEnd})
			e.pushAfterCombat()
		}
	}
	return nil
}

// devBack: as opCombat after a shop or temple, the 3D view comes back.
func (e *Engine) devBack() {
	e.setWorldWord(rules.WordBeginFlag, 4)
	e.DrawMainScreen()
	e.ret()
}

// devFightEnd ends a dev scene's fight as opCombat's ECL loop does: a lost
// party (run.quit, combatResult) ends the game through stepEnd, as a
// script fight's does; else the explore screen (BUG-CLASSES 53).
func (e *Engine) devFightEnd() {
	if e.run.quit {
		e.stepEnd()
		return
	}
	e.DrawMainScreen()
	e.ret()
}

// devParlayAnswered is the dev encounter menu's step: WAIT a message, then
// the menu again (c -1: the first); else the talk-style menu.
func (e *Engine) devParlayAnswered(c int) {
	if c == 1 {
		e.paneText(game.InkLightGreen, true, "Both sides wait. The kobolds eye the party with suspicion.")
	}
	if c < 0 || c == 1 {
		e.askChoice("", "~COMBAT ~WAIT ~FLEE ~PARLAY", false, false, askDone{kind: dDevParlay})
		return
	}
	e.clearPrompt()
	e.es.inEncounterMenu = 0
	e.askChoice(" ", "~HAUGHTY ~SLY ~NICE ~MEEK ~ABUSIVE", false, false, askDone{})
}

// fixture has the ruleset set the party up for a dev scene (core.Fixture);
// who is the selected member, for the temple.
func (e *Engine) fixture(id rules.FixtureID, who rules.CharID) {
	if _, err := e.ruleset().Resolve(nil, rules.Fixture{ID: id, Who: who}, nil); err != nil {
		panic(fmt.Sprintf("dev scene %v: %v", id, err))
	}
}

// DevAt (-at BLOCK,X,Y) enters script block block at x, y, as a script's
// NEW ECL does (opNewECL), and runs the area's start events (AreaChange).
func (e *Engine) DevAt(block, x, y int) {
	e.setLastBlock(int(e.gs.areaKind))
	e.gs.areaKind = byte(block)
	e.eclFile = e.set.File("ECL") + ".dax"
	e.LoadArea(block)
	e.InitAreaScript()
	e.posX, e.posY = x, y
	e.AreaChange()
}

// TraceScripts (-ecltrace) prints each script command and its operands to stderr.
func (e *Engine) TraceScripts() {
	e.vm.Trace = func(vm *ecl.VM, n int) {
		fmt.Fprintf(os.Stderr, "ecl %04X op %02X\n", vm.PC, vm.Command)
		argsTrace(e.host, func(s string) { fmt.Fprint(os.Stderr, s) })(vm, n)
	}
}

// devCont runs the dev scenes' continuation steps.
func (e *Engine) devCont(c cont) {
	switch c.kind {
	case kDevBack:
		e.devBack()
	case kDevFought:
		e.push(cont{kind: kDevFightEnd})
		e.pushAfterCombat()
	case kDevFightEnd:
		e.devFightEnd()
	case kDevJournal:
		e.showEntry(e.entryNamed("JournalEntry26"))
		e.showEntry(e.entryNamed("JournalEntry4"))
		e.ret()
	case kDevLoot:
		e.do(rules.DevTreasure{})
		e.push(cont{kind: kDrawMain})
		e.lootMenu()
	}
}

// devAnswered runs the dev scenes' answered prompts.
func (e *Engine) devAnswered(d askDone, v ansv) {
	switch d.kind {
	case dDevParlay:
		e.devParlayAnswered(v.n)
	case dDevPassword:
		e.scriptText(true, "YOU SAY '"+v.s+"'.")
		e.askContinue(askDone{})
	}
}

// okay is whether member i's status is okay.
func (e *Engine) okay(i int) bool {
	return e.ruleset().Sheet(e.partyIDs()[i]).Status == rules.StatusOkay
}

// firstOkay is the first member whose status is okay, or -1.
func (e *Engine) firstOkay() int {
	for i := range e.partySize() {
		if e.okay(i) {
			return i
		}
	}
	return -1
}

// firstMonster is the lowest monster id of file; false: none.
func (d *gameData) firstMonster(file string) (int, bool) {
	d.monster(file, 0)
	first, ok := 0, false
	for id := range d.monsters[file] {
		if !ok || id < first {
			first, ok = id, true
		}
	}
	return first, ok
}

// blockedRet goes on with what follows unless something waits.
func (e *Engine) blockedRet() {
	if !e.blocked() {
		e.ret()
	}
}

// entryNamed is the id of the entry the scripts call name (a dev view's).
func (e *Engine) entryNamed(name string) ecl.EntryID {
	for id, en := range e.entries() {
		if en.Name == name {
			return id
		}
	}
	panic("no text entry " + name)
}

// argsTrace is a ready Trace printing operands the way the games' scriptTrace does.
func argsTrace(h ecl.Host, print func(string)) func(*ecl.VM, int) {
	return func(vm *ecl.VM, n int) {
		for k := 1; k <= n; k++ {
			print(fmt.Sprintf("    arg%d type %d %d %v = %d\n", k, vm.ArgType[k], vm.ArgN[k], vm.ArgRef[k], vm.Arg(k, h)))
		}
	}
}
