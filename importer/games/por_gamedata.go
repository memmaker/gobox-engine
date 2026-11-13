package games

import (
	"gobox/asset"
	"gobox/engine"
	"gobox/game"
	"image/color"
)

// doorColour is a light brown (Amiga palette 0xA84): it stands out
// against the white walls and the grey, black and blue floors.
const doorColour = asset.AmigaBase + 16*asset.AmigaPortrait + 6

// gameData is POR's declarative data for gobox (game.Data), written as
// data/gamedata.json: the series, the game id and the spell tips (SpellRules by spell id).
func porGameData() game.Data {
	d := game.Data{Series: "pool", ID: "por", SpellRules: map[int]string{}}
	d.VerbChords = map[string]string{"Manual": "space", "Magic": "alt+m"}                 // the computer turn's poll (ovr10:0FC1): Space takes the party back, Alt-M switches its casting
	d.VerbKeys = map[string]rune{"View": 'V'}                                             // combatMenu's View (ovr09:069E), gobox's sheet on the fight bar
	d.VerbOrder = []string{"Move", "View", "Aim", "Use", "Cast", "Turn", "Quick", "Done"} // combatMenu's bar (ovr09:069E-07D4)
	// Quick Start: two fighters in front, two clerics, two magic-users (Sleep is
	// the spell that wins the early fights) and a thief for locks and traps.
	d.QuickParty = []game.QuickPick{
		{Race: "Dwarf", Gender: "Male", Class: "Fighter", Align: "Lawful Good", Need: map[string]int{"str": 18, "con": 17}, Name: "Thorgrim", Head: "head_009", Body: "body_004"},
		{Race: "Human", Gender: "Female", Class: "Cleric", Align: "Lawful Good", Need: map[string]int{"wis": 17, "con": 15, "str": 15}, Name: "Mirelle", Head: "head_011", Body: "body_009"},
		{Race: "Half-Elf", Gender: "Male", Class: "Cleric/Fighter", Align: "Neutral Good", Need: map[string]int{"str": 17, "wis": 16, "con": 15}, Name: "Aldric", Head: "head_018", Body: "body_011"},
		{Race: "Elf", Gender: "Female", Class: "Fighter/Magic-User", Align: "Chaotic Good", Need: map[string]int{"str": 16, "int": 17, "dex": 16}, Name: "Sylvaine", Head: "head_022", Body: "body_012"},
		{Race: "Human", Gender: "Male", Class: "Magic-User", Align: "Neutral Good", Need: map[string]int{"int": 18, "dex": 16, "con": 14}, Name: "Corwin", Head: "head_008", Body: "body_016"},
		{Race: "Halfling", Gender: "Male", Class: "Fighter/Thief", Align: "Neutral Good", Need: map[string]int{"str": 16, "dex": 17, "con": 15}, Name: "Pip", Head: "head_025", Body: "body_001"},
	}

	d.IconParts, d.DungeonTiles, d.MapCell = engine.IconParts[:], engine.DungeonTiles, engine.MapCell
	d.Rests = []game.RestKind{ // camp's rests: until fixed while someone heals or studies, an hour, until night or day
		{Text: "Until fixed", Key: 'F', N: engine.RestFix, Fix: true},
		{Text: "Wait an hour", Key: 'W', N: engine.RestWait},
		{Text: "Until night", Key: 'N', N: engine.RestUntil, Night: "Until day", NightKey: 'D'},
	}
	d.DoorInk, d.SelectInk = asset.EGAIdentity.RGBA(doorColour), asset.EGARGBA(10)
	// TITLE picture 2's "Pool of Radiance" (yellow with a light blue shadow, above row 108)
	// and its subtitle (light magenta, rows 108 to 160) in the Amiga intro's colours: gold
	// (0xF90) with a dark red (0x600) outline, the intro's credits palette entries 2 and 1.
	d.Logo = &game.Logo{Sub: 108, Bottom: 160, Keep: []uint8{asset.EGAYellow, asset.EGALightBlue}, SubInk: asset.EGALightMagenta,
		Gold: color.RGBA{0xff, 0x99, 0, 0xff}, Outline: color.RGBA{0x66, 0, 0, 0xff}}
	d.Bindings = []game.Binding{{Ctx: "Any time", Keys: "F12", Desc: "art: Amiga, DOS", Mouse: "Esc>Graphics", Pad: "B, then pick"}}
	for id, t := range SpellRules {
		if t != "" {
			d.SpellRules[id] = t
		}
	}
	return d
}
