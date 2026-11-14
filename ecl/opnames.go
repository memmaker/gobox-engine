package ecl

import "fmt"

// The canonical commands no form handles by name stay numbers in opNames. A
// canonical number below 0x80 is the first engine's byte; the commands later
// dialects added are from 0x80.
const (
	OpSelectMember  byte = 0x0A
	OpLoadMonster   byte = 0x0B
	OpSetupMonster  byte = 0x0C
	OpApproach      byte = 0x0D
	OpCheckParty    byte = 0x1E
	OpPartySurprise byte = 0x22
	OpSurprise      byte = 0x23
	OpRob           byte = 0x28
	OpEncounterMenu byte = 0x29
	OpDamage        byte = 0x2E
	OpRedrawView    byte = 0x31
	OpClock         byte = 0x34
	OpAddNPC        byte = 0x36
	OpWho           byte = 0x39
	OpDelay         byte = 0x3A
	OpSpell         byte = 0x3B
	OpClearBox      byte = 0x3D
)

const (
	OpLoadWalls    byte = 0x37
	OpProtection   byte = 0x3C
	OpClearExit    byte = 0x80 // clear, redraw and stop
	OpPressKey     byte = 0x81
	OpJournal      byte = 0x82
	OpYesNo        byte = 0x83
	OpNoOp         byte = 0x84
	OpSound        byte = 0x85
	OpMenuKeys     byte = 0x86 // HMenu that also takes the key held
	OpFindNPC      byte = 0x87
	OpItemSet      byte = 0x88
	OpSpriteOff    byte = 0x89
	OpDump         byte = 0x8A
	OpFindAffect   byte = 0x8B
	OpDestroy      byte = 0x8C
	OpGainExp      byte = 0x8D
	OpTreasureOf   byte = 0x8E // Treasure with a counted list
	OpShade        byte = 0x8F
	OpPartyStat    byte = 0x90
	OpSpaceFight   byte = 0x91
	OpSkill        byte = 0x92
	OpSkillAll     byte = 0x93
	OpHasItem      byte = 0x94
	OpNPCJoin      byte = 0x95
	OpMenuCall     byte = 0x96
	OpPickMember   byte = 0x97
	OpLogbook      byte = 0x98
	OpOp64         byte = 0x99
	OpWallDist     byte = 0x9A
	OpFor          byte = 0x9B
	OpNext         byte = 0x9C
	OpOp72         byte = 0x9D
	OpDuel         byte = 0x9E
	OpShop         byte = 0x9F
	OpAreaKind     byte = 0xA0
	OpDamage6      byte = 0xA1
	OpNothing      byte = 0xA2
	OpLoadArea     byte = 0xA3 // LoadMap with no icons
	OpQuest        byte = 0xA4 // a quest handed over: the adventure's Quest value
	OpCastKnock    byte = 0xA5 // Knock cast by the first able member: its party place, or 255
	OpTavernGamble byte = 0xA6 // POR's tavern top spun for gold wagers
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
	OpClearExit: "ClearExit", OpPressKey: "PressKey", OpJournal: "Journal", OpYesNo: "YesNo", OpNoOp: "NoOp", OpSound: "Sound",
	OpMenuKeys: "MenuKeys", OpFindNPC: "FindNPC", OpItemSet: "ItemSet", OpSpriteOff: "SpriteOff", OpDump: "Dump",
	OpFindAffect: "FindAffect", OpDestroy: "DestroyItems", OpGainExp: "GainExperience", OpTreasureOf: "TreasureOf", OpShade: "Shade",
	OpPartyStat: "PartyStat", OpSpaceFight: "SpaceFight", OpSkill: "Skill", OpSkillAll: "SkillAll", OpHasItem: "HasItem",
	OpNPCJoin: "NPCJoin", OpMenuCall: "MenuCall", OpPickMember: "PickMember", OpLogbook: "Logbook", OpOp64: "Op64",
	OpWallDist: "WallDistance", OpFor: "For", OpNext: "Next", OpOp72: "Op72", OpDuel: "Duel", OpShop: "Shop", OpAreaKind: "AreaKind",
	OpDamage6: "Damage6", OpNothing: "Nothing", OpLoadArea: "LoadArea", OpQuest: "Quest", OpCastKnock: "CastKnock", OpTavernGamble: "TavernGame",
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
