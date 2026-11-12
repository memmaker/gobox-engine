// Package script is the one decompiled form of ECL event scripts: Decode
// turns an ECL block into a Script (instruction-numbered commands), Write
// and Parse carry it as text, MarkExits and FoldVendorGates add what the
// frontend needs. The differences between the games' byte code are data: a
// typed Version, chosen at import; nothing here knows a game.
package script

// FormatRevision is the text form's revision; a change makes every import stale.
const FormatRevision = 2

// VersionID names one byte-code dialect.
type VersionID int

// Cmd is how a Version encodes one byte-code command.
type Cmd struct {
	Op    byte // the canonical command (form.go, ops.go)
	Fixed int  // operands always there
	Count int  // the operand (from 1) holding how many more follow; 0 none
	Known bool // the original's dispatcher has a case for it
}

// ExitPlace is a menu instruction of the decoded script of an area.
type ExitPlace struct{ Area, Instr int }

// NoExitRow in ExitOverrides: every row of the menu is a real choice.
const NoExitRow = -1

// Version is one dialect: where its code sits and what its commands are.
type Version struct {
	ID            VersionID
	Base          int               // the address of the code's first byte
	Cmds          [256]Cmd          // by byte-code command
	ExitOverrides map[ExitPlace]int // corrections of the exit heuristic
	// LoadMapSecondUnused: the dialect's engine ignores LoadMap's second
	// operand, so the text leaves it out.
	LoadMapSecondUnused bool
	// Skip[b] is how many operands the dialect's engine reads when it steps
	// over byte-code command b unrun (an If that fails: parse_command's own
	// table). It differs from Cmds[b].Fixed where the original's table did
	// (TestSkipCounts names each difference).
	Skip [256]byte
	// Calls names the engine routine each CALL operand (its raw value, an
	// address in the original) runs; an operand not listed runs nothing.
	Calls map[uint16]CallFn
	// Programs names what each PROGRAM operand does; one not listed does nothing.
	Programs map[byte]ProgramFn
}

// ProgramFn is a PROGRAM sub-command, by name (ruling S14: each game's
// codes map onto one set).
type ProgramFn int

const (
	ProgramNone      ProgramFn = iota
	ProgramPartyMenu           // the party menu
	ProgramEnding              // the adventure's ending
	ProgramCamp                // camp (pod: with its camp events)
	ProgramQuit                // leave the game
	ProgramSaveQuit            // save, then the ending
)

// CallFn is an engine routine a script CALLs, by name.
type CallFn int

const (
	CallNone        CallFn = iota
	CallRedrawView         // redraw the view if the script changed it
	CallStepForward        // the party steps one square ahead
	CallWallAhead          // the wall the party faces into the wall variable
	CallDuelWin            // the duel's winning setup
	CallDuelLose           // the duel's losing setup
	CallDuelSound          // the duel's sound, by the sound variable
	CallAnimate            // the picture's next step
)

// skips makes a Skip table of the groups: operand count -> commands.
func skips(groups map[byte][]byte) (t [256]byte) {
	for n, cs := range groups {
		for _, c := range cs {
			t[c] = n
		}
	}
	return t
}

// Current reports whether a text carrying id and revision is this version's.
func (v *Version) Current(id VersionID, revision int) bool {
	return id == v.ID && revision == FormatRevision
}

// commands makes the commands with n fixed operands of the groups' lists; a
// command's canonical number is its byte value.
func commands(groups map[int][]byte) (t [256]Cmd) {
	for n, cs := range groups {
		for _, c := range cs {
			t[c] = Cmd{Op: c, Fixed: n, Known: true}
		}
	}
	return t
}

// counted sets commands whose last fixed operand counts the rest.
func counted(t *[256]Cmd, fixedAndCount int, cs ...byte) {
	for _, c := range cs {
		t[c] = Cmd{Op: c, Fixed: fixedAndCount, Count: fixedAndCount, Known: true}
	}
}

// V1 is the dialect of the first engine (Pool of Radiance's ECL*.DAX).
var V1 = func() *Version {
	t := commands(map[int][]byte{
		0:  {OpExit, 0x0D, OpReturn, OpIfEq, OpIfNe, OpIfLt, OpIfGt, OpIfLe, OpIfGe, OpClearMonsters, OpCombat, 0x31, OpNewLine, 0x3A, 0x3D},
		1:  {OpGoto, OpGosub, 0x0A, OpPicture, OpPrint, OpPrintClear, OpPartyStr, OpNewECL, OpCall, OpFindItem, 0x34, OpProgram, 0x39, 0x3C},
		2:  {OpCompare, OpRandom, OpSave, OpInputNumber, OpInputString, 0x22, 0x36},
		3:  {OpAdd, OpSub, OpDiv, OpMul, 0x0B, 0x0C, OpLoadMap, 0x28, OpGetTable, OpAnd, OpOr, OpSaveTable, OpLoadWalls, 0x3B},
		4:  {OpCompareAnd, 0x23},
		5:  {0x2E},
		6:  {0x1E, OpParlay},
		8:  {OpTreasure},
		14: {0x29},
	})
	counted(&t, 3, OpVMenu)
	counted(&t, 2, OpOnGoto, OpOnGosub, OpHMenu)
	// the pager (Felix 2026-11-11, one for every game): unused by the first
	// engine, so at Pools of Darkness' own number
	set(&t, 31, OpJournal, 1, 0)
	// a quest handed over (the quest journal; Felix 2026-11-11: never by
	// matching its printed text): unused by every Gold Box dialect
	set(&t, 66, OpQuest, 1, 0)
	// SkipEclCmd's counts (por-re src/eclmenu.pas:1041; Felix 2026-11-11:
	// POR steps over a command as the original did)
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 29, 32, 45, 50, 52, 54, 56, 57, 60},
		2: {3, 8, 9, 12, 15, 16, 31, 34},
		3: {4, 5, 6, 7, 11, 33, 40, 42, 47, 48, 53, 55, 59},
		4: {20, 35}, 5: {46}, 6: {30, 44}, 8: {39}, 14: {41},
	})
	return &Version{ID: 1, Base: 0x9900, Cmds: t, ExitOverrides: v1ExitOverrides, LoadMapSecondUnused: true, Skip: skip,
		// ovr04:2FE8's routine numbers, operand - 0x7FFF in the original
		Calls: map[uint16]CallFn{0x2C90: CallRedrawView, 0x8000: CallDuelWin, 0x8001: CallDuelLose, 0xBA03: CallDuelSound,
			0xC01E: CallStepForward, 0xC018: CallWallAhead},
		Programs: map[byte]ProgramFn{0: ProgramPartyMenu, 8: ProgramEnding, 9: ProgramCamp}}
}()

// set makes byte-code command b the canonical command op with these operands.
func set(t *[256]Cmd, b, op byte, fixed, count int) {
	t[b] = Cmd{Op: op, Fixed: fixed, Count: count, Known: true}
}

// V2 is the dialect of the second engine (Curse of the Azure Bonds' and
// Secret of the Silver Blades' ECL*.DAX): the first's, with 65 added.
var V2 = func() *Version {
	t := V1.Cmds
	set(&t, 49, OpSpriteOff, 0, 0)
	set(&t, 62, OpDump, 0, 0)
	set(&t, 63, OpFindAffect, 1, 0)
	set(&t, 64, OpDestroy, 1, 0)
	set(&t, 65, OpGainExp, 2, 0)
	set(&t, 52, 0x34, 2, 0)
	// parse_command's counts (cotab O04C:1FD2, sotsb's); 65 is sotsb's only
	// (cotab's scripts have no GainExperience).
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 29, 32, 45, 50, 52, 54, 56, 57, 60, 63, 64},
		2: {3, 8, 9, 15, 16, 31, 34, 65},
		3: {4, 5, 6, 7, 11, 12, 33, 40, 42, 47, 48, 53, 55, 59},
		4: {20, 35}, 5: {46}, 6: {30, 44}, 8: {39}, 14: {41},
	})
	// CALL's addresses: Curse of the Azure Bonds' (O03C:2F23) and Secret of
	// the Silver Blades' (O050:33C6) differ for RedrawView and DuelSound;
	// none collides, so the one dialect carries both.
	calls := map[uint16]CallFn{11792: CallRedrawView, 11723: CallRedrawView, 32768: CallDuelWin, 32769: CallDuelLose,
		45568: CallDuelSound, 45571: CallDuelSound, 49182: CallStepForward, 49176: CallWallAhead, 26627: CallAnimate}
	// PROGRAM: cotab's 0, 3 quit, 8 ending, 9 camp; sotsb uses 0 and 9
	programs := map[byte]ProgramFn{0: ProgramPartyMenu, 3: ProgramQuit, 8: ProgramEnding, 9: ProgramCamp}
	return &Version{ID: 2, Base: 32768, Cmds: t, Skip: skip, Calls: calls, Programs: programs}
}()

// V3 is Pools of Darkness' dialect.
var V3 = func() *Version {
	t := V2.Cmds
	set(&t, 12, 0x0C, 4, 0)
	set(&t, 29, OpPressKey, 0, 0)
	set(&t, 31, OpJournal, 1, 0)
	set(&t, 33, OpLoadArea, 2, 0)
	set(&t, 34, OpYesNo, 0, 0)
	set(&t, 35, OpClearExit, 0, 0)
	set(&t, 39, OpTreasureOf, 4, 4)
	set(&t, 41, OpNoOp, 0, 0)
	set(&t, 44, OpSound, 1, 0)
	set(&t, 49, OpMenuKeys, 2, 2)
	set(&t, 55, OpFindNPC, 2, 0)
	set(&t, 58, OpNoOp, 0, 0)
	set(&t, 60, OpItemSet, 1, 0)
	// Pools of Darkness' parse_command counts (O055:1464)
	skip := skips(map[byte][]byte{
		1: {1, 2, 10, 14, 17, 18, 31, 32, 44, 45, 50, 52, 54, 56, 57, 60, 63, 64},
		2: {3, 8, 9, 15, 16, 65, 33, 55},
		3: {4, 5, 6, 7, 11, 40, 42, 47, 48, 53, 59},
		4: {12, 20}, 5: {46}, 6: {30},
	})
	// pod renumbered its routines (O040:1E4A)
	return &Version{ID: 3, Base: 32768, Cmds: t, Skip: skip, Calls: map[uint16]CallFn{1: CallRedrawView, 2: CallStepForward, 3: CallAnimate},
		Programs: map[byte]ProgramFn{0: ProgramPartyMenu, 1: ProgramSaveQuit, 2: ProgramQuit, 9: ProgramCamp}}
}()

// V4 is the Buck Rogers engine's dialect (Countdown to Doomsday, Matrix Cubed).
var V4 = func() *Version {
	var t [256]Cmd
	for b, c := range V1.Cmds {
		if b < 0x40 {
			t[b] = c
		}
	}
	for _, c := range []struct {
		b, op        byte
		fixed, count int
	}{
		{12, 0x0C, 4, 0}, {29, OpShade, 2, 0}, {30, OpPartyStat, 6, 0}, {31, OpSpaceFight, 4, 0}, {34, OpSkill, 3, 0}, {35, OpSkillAll, 3, 0},
		{39, OpTreasureOf, 2, 2}, {41, OpPressKey, 0, 0}, {44, OpYesNo, 0, 0}, {45, OpCall, 1, 0}, {46, 0x2E, 5, 0}, {49, OpMenuKeys, 2, 2},
		{50, OpHasItem, 1, 0}, {52, 0x34, 2, 0}, {54, OpNPCJoin, 2, 0}, {55, OpLoadWalls, 3, 0}, {56, OpMenuCall, 1, 0}, {57, OpPickMember, 1, 0},
		{58, 0x3A, 0, 0}, {59, OpNothing, 0, 0}, {60, OpNothing, 0, 0}, {61, 0x31, 0, 0}, {62, OpDump, 0, 0}, {63, OpLogbook, 2, 0}, {64, OpOp64, 2, 0},
		{65, OpGainExp, 2, 0}, {66, OpClearExit, 0, 0}, {67, OpSound, 1, 0}, {68, OpNothing, 0, 0}, {69, OpWallDist, 2, 0}, {70, OpFor, 2, 0}, {71, OpNext, 0, 0},
		{72, OpOp72, 1, 0}, {73, OpDamage6, 6, 0}, {74, OpDuel, 0, 0}, {75, OpShop, 1, 0}, {76, OpAreaKind, 2, 0},
	} {
		set(&t, c.b, c.op, c.fixed, c.count)
	}
	return &Version{ID: 4, Base: 32768, Cmds: t}
}()

// Versions are the dialects by ID.
var Versions = map[VersionID]*Version{V1.ID: V1, V2.ID: V2, V3.ID: V3, V4.ID: V4}
