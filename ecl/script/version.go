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
	Vars          map[int]string    // the words the dialect's engine names
	ExitOverrides map[ExitPlace]int // corrections of the exit heuristic
	// LoadMapSecondUnused: the dialect's engine ignores LoadMap's second
	// operand, so the text leaves it out.
	LoadMapSecondUnused bool
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
	return &Version{ID: 1, Base: 0x9900, Cmds: t, Vars: v1Vars, ExitOverrides: v1ExitOverrides, LoadMapSecondUnused: true}
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
	return &Version{ID: 2, Base: 32768, Cmds: t, Vars: map[int]string{0x7F6C: varShop, 0x7EE2: varTemple}}
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
	return &Version{ID: 3, Base: 32768, Cmds: t, Vars: map[int]string{0x0014: varTemple}}
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
