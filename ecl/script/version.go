// Package script is the one decompiled form of ECL event scripts: Decode
// turns an ECL block into a Script (instruction-numbered commands), Write
// and Parse carry it as text, MarkExits and FoldVendorGates add what the
// frontend needs. The differences between the games' byte code are data: a
// typed Version, chosen at import; nothing here knows a game.
package script

// FormatRevision is the text form's revision; a change makes every import stale.
const FormatRevision = 1

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
}

// Current reports whether a text carrying id and revision is this version's.
func (v *Version) Current(id VersionID, revision int) bool {
	return id == v.ID && revision == FormatRevision
}

// fixed makes the commands with n fixed operands of the groups' lists; a
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
	return &Version{ID: 1, Base: 0x9900, Cmds: t, Vars: v1Vars, ExitOverrides: v1ExitOverrides}
}()
