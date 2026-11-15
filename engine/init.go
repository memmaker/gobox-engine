package engine

import (
	"gobox/adventure"
	"gobox/asset"
	"gobox/game"
	"gobox/rules"
)

// Init = ovr12:002D: allocate the global records, set the start state and
// load the always-resident assets. Everything the DOS code zeroed with
// FillChar is zero-valued here.
func (e *Engine) Init() {
	clear(e.vm.Code)
	e.set = 1 // DS:5376 = 1
	e.resetAdventure()
	e.clearPrompt()
	e.prompt(game.InkLightGreen, "Loading...Please Wait")
	symbols := e.game().Symbols
	if symbols == nil {
		symbols = adventure.DefaultSymbols
	}
	for _, l := range symbols { // ovr35:003B
		e.LoadSymbolSet(l)
	}
	e.sqrpaci = [2]*asset.Image{blankTiles()}
	for _, l := range e.artNames().TileSheets { // POR: SQRPACI 1 and 2, 0464:0111 → DS:5E52, DS:5E56
		e.sqrpaci[l.First/0x80] = e.LoadPic(l.File, l.Block)
	}
	for _, l := range e.artNames().Tiles {
		e.LayTiles(l)
	}
	spr := e.artNames().CombatSprites
	for i := 0; i <= 0xB; i++ {
		e.LoadIconPic(spr, i, i+0xD) // ovr33:01C8
	}
	e.LoadIconPic(spr, 0x19, 0x19)
	e.do(rules.LoadItemTable{}) // ITEMS file, 2-byte header then the records → DS:5582
	e.clearPrompt()
	e.Randomize()           // 0AF8:1247
	e.do(rules.InitRules{}) // ovr23:3413 and the affect handler table
}

// resetAdventure puts the adventure's state back to a new game's: the
// game, area and aux records, New Phlan as the area, the start square.
// Init's part of it, and (extra) what the party menu needs again after the
// party was lost, where the original exited to DOS: the area and position
// the party died at would otherwise carry over to the next party's Begin
// (a wipe on the first square after an area change loaded the old area,
// whose block still named it, at the new one's square: inside the water).
func (e *Engine) resetAdventure() {
	e.do(rules.NewAdventure{})                    // game and area records DS:49D6, 49D2, an empty purse
	e.posX, e.posY, e.facing = 0xF, 1, 6          // the start square DS:6AAD/6AAE/6AAF (the Engine's)
	e.vars, e.locals = fieldCells{}, fieldCells{} // DS:49DA
	e.gs.areaKind = 0
	e.pos3, e.pos4 = 0, 0
	e.vm.PC = 0    // DS:49ED
	e.msgDelay = 1 // DS:49E2
	e.viewMode, e.prevViewMode = 1, 1
	e.viewBit0, e.viewBit1 = 1, 1                                                                                                                                        // DS:49F5/49F6
	e.mapPairA, e.mapPairB = [3]int{0, -1, -1}, [3]int{0, -1, -1}                                                                                                        // DS:6AAE+4i / 6AB0+4i, i = 2..3
	e.es.skyShade, e.es.groundShade = 0xE8, 0xDB                                                                                                                         // DS:6F69/6F68
	e.ext.seen, e.ext.voiced, e.ext.met, e.ext.beasts, e.ext.labels, e.ext.quests, e.ext.notes, e.ext.doors, e.ext.secrets = nil, nil, nil, nil, nil, nil, nil, nil, nil // extra: kept with the save, not the party menu
}
