package pool

import (
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/core/grid"
	"gobox/ecl"
	"gobox/ecl/script"
	"gobox/ega"
)

// InitAreaScript = ovr08:01B6.
func (e *Engine) InitAreaScript() {
	s := &e.es
	s.picShown, s.textInView, s.memberDied, s.memberGone = 0, 0, 0, 0
	s.groundShade, s.skyShade, s.skyChanged = 0x41, 9, 1
	s.encounterShown = [2]byte{}
	s.cloneIcon = 8
	e.vm.Calls = nil
	e.vm.Flags = [6]bool{}
	e.setWorldWord(core.WordPortrait, 0xFF)
	e.setWorldWord(core.WordRestEvery, 0)
	e.setWorldWord(core.WordRestChance, 0)
	e.setWorldWord(core.WordNoMagic, 0)
	s.events = e.vm.Script.Events
	e.setWorldWord(core.WordIn3D, 1)
	if !e.loadedGame {
		e.locals = fieldCells{} // area words 0x4A00..0x4A1F and game words 0x6E79..0x6E82
	} else if e.loadedLocals != nil { // nil: a load that kept the running game's state (slots.go)
		e.locals, e.loadedLocals = e.loadedLocals, nil
		if l := e.vm.Script.Layout; l != nil {
			foldElements(e.locals, l.Fields)
		}
		e.loadedGame = false
	} else {
		e.loadedGame = false
	}
}

// LoadArea = ovr08:0341: ECL<n>.DAX block areaId → ECL image (first word skipped).
func (e *Engine) LoadArea(areaID int) {
	e.loadingMonster = false
	e.clearPrompt()
	e.prompt(ega.White, "Loading...Please Wait")
	e.loadScript(areaID)
	e.clearPrompt()
}

// FreeSteps = ovr08:0482.
func (e *Engine) FreeSteps(x, y, dir int) int {
	if e.worldWord(core.WordIn3D) == 0 {
		e.setWorldWord(core.WordMonsterDist, 2)
		return 2
	}
	n := 0
	for n < 2 {
		if e.WallAt(x, y, dir) != 0 {
			break
		}
		n++
		switch dir {
		case 0:
			y--
		case 2:
			x++
		case 4:
			y++
		case 6:
			x--
		}
	}
	return n
}

// StepForward = ovr08:19DB: one square ahead. Within a map it never wraps:
// at the edge the party stays put. A step out through an open side
// (checkEdge set LeftMap) lands on the far edge as in the original, since
// that is the entry square the next area keeps (west out of Phlan, in at
// the east edge): an area change, not a wrap within the map.
func (e *Engine) StepForward() {
	x, y := e.posX+grid.StepX[e.moveDir()], e.posY+grid.StepY[e.moveDir()]
	if e.worldWord(core.WordLeftMap) != 0 {
		x, y = (x+16)%16, (y+16)%16
	}
	if x >= 0 && x <= 15 && y >= 0 && y <= 15 {
		e.posX, e.posY = x, y
	}
	e.es.cell = e.MapCell(e.posX, e.posY)
	e.es.wallAhead = e.WallAt(e.posX, e.posY, e.facing)
}

// HitMessage = ovr08:20CC (c, dmg).
func (e *Engine) HitMessage(h core.Hurt) {
	var msg string
	if h.Dies {
		msg = "  " + e.ruleset().Sheet(h.Unit).Name + " dies. "
	} else {
		msg = "  " + e.ruleset().Sheet(h.Unit).Name + " is hit FOR " + fmt.Sprintf("%3d", h.Damage) + " points of Damage."
	}
	if e.msgRow > 0x16 {
		e.msgRow = 0x11
		e.pageOn() // the original's "press <enter>/<return> to continue": gobox pages the lots
	}
	e.msgCol = 0x26
	e.paneText(ega.White, false, msg)
	e.blank(17, 1, paneRight, 15)
}

// initVM makes the script machine: the game's dialect (POR's by default), ecl.Default and its own ops,
// no area loaded.
func (e *Engine) initVM() {
	e.host = eclHost{e}
	e.engineFields = e.newEngineFields()
	e.vars, e.locals = fieldCells{}, fieldCells{}
	e.vm = ecl.New(e.host, e.eclLayout(), make([]byte, e.game().OriginalSave.CodeSize))
	e.eclVersion = e.game().Script
	if e.eclVersion == nil {
		e.eclVersion = script.V1
	}
	e.vm.SetVersion(e.eclVersion)
	e.vm.Script = &script.Script{}
	for b, c := range e.eclVersion.Cmds {
		if op := eclOps[c.Op]; c.Known && op != nil {
			e.vm.Ops[b] = func(*ecl.VM, ecl.Host) { op(e) }
		}
	}
}

// ECL virtual-machine support, ported from ovr08. The script's variables
// are the typed fields of scriptvars.go; the save's records are the game's
// data (GameTables.OriginalSave), the code image's base and size with them.
const (
	areaScratchWords = 0x20 // the area record's words 0x4A00..0x4A1F: locals, cleared by a fresh area (SaveEditor.EnterArea)
)

// extractHotkeys = ovr08:156D: "~Yes ~No" → menu "yes no", hotkeys "YN".
func extractHotkeys(menu string) (hot, out string) {
	b := []byte(menu)
	var res []byte
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c == '~' {
			if i+1 < len(b) {
				h := goldbox.Upper(b[i+1])
				if len(hot) < 0x28 {
					hot += string(h)
				}
				res = append(res, h)
				i++
			}
			continue
		}
		if isHotkey(c) {
			c += 'a' - 'A' // digits too: '0'..'9' turn into 'P'..'Y', as in the original
		}
		res = append(res, c)
	}
	return hot, string(res)
}

const (
	holdNone      holdKind = iota
	holdPartyMenu          // the party menu
	holdEncamp             // camp, then the script exits
)

// eclLayout: the VM needs only the code image's range, the game's (GameTables.OriginalSave).
func (e *Engine) eclLayout() ecl.Layout {
	s := e.game().OriginalSave
	if s.CodeSize == 0 {
		panic("gametables.json OriginalSave: no code image (run the game's import)")
	}
	return ecl.Layout{Code: ecl.Range{First: uint16(s.CodeBase), Last: uint16(s.CodeBase + s.CodeSize - 1)}}
}

// eclState mirrors the DS globals used by the interpreter.
type eclState struct {
	enc                                *encounter // the approach menu while it is asked
	takes                              int        // ops that took the script over (takeover)
	hold                               holdKind   // extra: the instruction opened something the script waits on (the party menu, camp)
	holdPC                             int        // holdEncamp: the instruction's pc, where the script exits
	groundShade                        byte       // DS:6F68 (init 0x41)
	skyShade                           byte       // DS:6F69 (init 9)
	specialFB                          uint16     // read-only specials 0xFB/0xFC/0xB1; DS:6F6A
	specialFC                          uint16     // DS:6F6C
	specialB1                          uint16     // DS:6F6E
	cloneIcon                          byte       // CPIC id for ROLF clone (8); DS:6F83
	events                             [5]int     // DS:49E3,49E5,49E7,49E9,49EB
	wallAhead                          byte       // DS:6AB0
	cell                               byte       // DS:6AB1
	encounterShown                     [2]byte    // DS:84C8
	soundVar, scriptVarB8, scriptVarB9 uint16     // DS:84CC/84CE/84D0
	memberDied, memberGone             byte       // DS:84D2/84D3 (DS:84D4 is run.duelMode)
	memberMissing                      bool       // the last SELECT MEMBER matched nobody (cotab characterNotFound: CotabSelStatus reads 0 once)
	picShown                           byte       // DS:84E0
	picAnimating                       byte       // DS:84E1
	textInView                         byte       // DS:84E2
	spriteLoaded                       byte       // DS:84E3
	skyChanged                         byte       // DS:84E5
	posChanged                         byte       // DS:84E6
	portraitChanged                    byte       // DS:84E8
	inEncounterMenu                    byte       // DS:84E9
	shownFace                          byte       // DS:84EA
	picHead, picBody                   byte       // DS:49FD/49FE
}

// holdKind is what an instruction opened that the script waits for.
type holdKind int

// eclHost is what the shared VM asks of the game.
type eclHost struct{ e *Engine }

func (h eclHost) Get(r script.Ref) int64 { return h.e.getVar(r) }

func (h eclHost) Set(r script.Ref, v int64) { h.e.setVar(r, v) } // cut to the field's width

func (h eclHost) Text(r script.Ref) string { return h.e.getText(r) }

func (h eclHost) SetText(r script.Ref, s string) { h.e.setText(r, s) }

func (h eclHost) Random(n int) int { return h.e.rollRandom(n) }

// Exit: POR's RETURN with no GOSUB goes on with the next instruction
// (ovr04:0A9F).
func (h eclHost) Exit(*ecl.VM) {}
