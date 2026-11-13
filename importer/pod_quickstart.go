package importer

import (
	"gobox/game"
)

// quickParty: two fighters in front (one with thief skills for locks),
// a cleric, a cleric/fighter, a magic-user and a fighter/magic-user.
var podQuickParty = []game.QuickPick{
	{Race: "Dwarf", Gender: "Male", Class: "Fighter", Align: "Lawful Good", Need: map[string]int{"str": 17, "con": 16}, Name: "Thorgrim", Head: "head_022", Body: "body_008"},
	{Race: "Human", Gender: "Female", Class: "Cleric", Align: "Lawful Good", Need: map[string]int{"wis": 16, "con": 14}, Name: "Mirelle", Head: "head_012", Body: "body_020"},
	{Race: "Half-Elf", Gender: "Male", Class: "Cleric/Fighter", Align: "Neutral Good", Need: map[string]int{"str": 15, "wis": 15}, Name: "Aldric", Head: "head_001", Body: "body_007"},
	{Race: "Elf", Gender: "Female", Class: "Fighter/Magic-User", Align: "Chaotic Good", Need: map[string]int{"str": 15, "int": 16}, Name: "Sylvaine", Head: "head_018", Body: "body_045"},
	{Race: "Human", Gender: "Male", Class: "Magic-User", Align: "Neutral Good", Need: map[string]int{"int": 17, "dex": 15}, Name: "Corwin", Head: "head_030", Body: "body_017"},
	{Race: "Halfling", Gender: "Male", Class: "Fighter/Thief", Align: "Neutral Good", Need: map[string]int{"str": 15, "dex": 16}, Name: "Pip", Head: "head_005", Body: "body_010"},
}
