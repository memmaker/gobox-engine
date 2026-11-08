package script

import "fmt"

// The canonical commands no form handles by name stay numbers in opNames. A
// canonical number below 0x80 is the first engine's byte; the commands later
// dialects added are from 0x80.
const (
	OpLoadWalls  byte = 0x37
	OpProtection byte = 0x3C
	OpClearExit  byte = 0x80 // clear, redraw and stop
)

// opNames are the commands' names in script text.
var opNames = map[byte]string{
	OpExit: "Exit", OpGoto: "Goto", OpGosub: "Gosub", OpCompare: "Compare", OpAdd: "Add", OpSub: "Sub", OpDiv: "Div", OpMul: "Mul",
	OpRandom: "Random", OpSave: "Save", 0x0A: "SelectMember", 0x0B: "LoadMonster", 0x0C: "SetupMonster", 0x0D: "Approach",
	OpPicture: "Picture", OpInputNumber: "InputNumber", OpInputString: "InputString", OpPrint: "Print", OpPrintClear: "PrintClear",
	OpReturn: "Return", OpCompareAnd: "CompareAnd", OpVMenu: "VMenu", OpIfEq: "IfEq", OpIfNe: "IfNe", OpIfLt: "IfLt", OpIfGt: "IfGt",
	OpIfLe: "IfLe", OpIfGe: "IfGe", OpClearMonsters: "ClearMonsters", OpPartyStr: "PartyStrength", 0x1E: "CheckParty", OpNewECL: "NewECL",
	OpLoadMap: "LoadMap", 0x22: "PartySurprise", 0x23: "Surprise", OpCombat: "Combat", OpOnGoto: "OnGoto", OpOnGosub: "OnGosub",
	OpTreasure: "Treasure", 0x28: "Rob", 0x29: "EncounterMenu", OpGetTable: "GetTable", OpHMenu: "HMenu", OpParlay: "Parlay", OpCall: "Call",
	0x2E: "Damage", OpAnd: "And", OpOr: "Or", 0x31: "RedrawView", OpFindItem: "FindItem", OpNewLine: "NewLine", 0x34: "Clock",
	OpSaveTable: "SaveTable", 0x36: "AddNPC", OpLoadWalls: "LoadWalls", OpProgram: "Program", 0x39: "Who", 0x3A: "Delay", 0x3B: "Spell",
	OpProtection: "Protection", 0x3D: "ClearBox",
	OpClearExit: "ClearExit",
}

// OpName is the command's name in script text.
func OpName(op byte) string {
	if n, ok := opNames[op]; ok {
		return n
	}
	return fmt.Sprintf("Op%02X", op)
}

// OpNamed is the command of a name (OpName), 0 for none.
func OpNamed(name string) byte {
	return opByName[name]
}

var opByName = func() map[string]byte {
	m := map[string]byte{}
	for op := range 256 {
		m[OpName(byte(op))] = byte(op)
	}
	return m
}()
