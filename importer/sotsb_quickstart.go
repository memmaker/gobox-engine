package importer

import (
	"gobox/game"
)

// quickParty suits the start: a dwarf fighter and a paladin in front, a
// cleric, a mage, a thief and a cleric/magic-user.
var sotsbQuickParty = []game.QuickPick{
	{Race: "Dwarf", Gender: "Male", Class: "Fighter", Align: "Lawful Good", Need: map[string]int{"str": 16, "con": 15}, Name: "Brann", Head: "head_022", Body: "body_008"},
	{Race: "Human", Gender: "Female", Class: "Paladin", Align: "Lawful Good", Need: map[string]int{"str": 15, "wis": 13, "cha": 14}, Name: "Elowen", Head: "head_012", Body: "body_020"},
	{Race: "Human", Gender: "Male", Class: "Cleric", Align: "Neutral Good", Need: map[string]int{"wis": 15, "con": 13}, Name: "Osric", Head: "head_001", Body: "body_007"},
	{Race: "Elf", Gender: "Male", Class: "Magic-User", Align: "Neutral Good", Need: map[string]int{"int": 15, "dex": 14}, Name: "Ilvan", Head: "head_030", Body: "body_017"},
	{Race: "Halfling", Gender: "Female", Class: "Thief", Align: "Neutral Good", Need: map[string]int{"dex": 15, "con": 12}, Name: "Tilly", Head: "head_023", Body: "body_010"},
	{Race: "Half-Elf", Gender: "Female", Class: "Cleric/Magic-User", Align: "Neutral Good", Need: map[string]int{"wis": 13, "int": 13}, Name: "Maren", Head: "head_018", Body: "body_045"},
}
