package engine

import (
	"fmt"
	"gobox/ega"
	"gobox/rules"
	"image"
	"slices"
)

// combatSetup = ovr11:1E31.
func (e *Engine) combatSetup() {
	e.as(e.sel)
	e.ruleset().BeginFight(rules.Fight{Board: rules.Board{Straight: e.Deskew.Load()}, Terms: rules.FightTerms{Duel: e.duel(), PartyAt: image.Pt(e.posX, e.posY), PartyFacing: rules.Facing(e.facing)}}, nil)
	e.FreeAnimPic()
	e.port = portrait{headID: 0xFF, bodyID: 0xFF} // 0464:2595 x2
	e.clearPrompt()
	e.ext.note = "" // extra's note goes with the row the original clears (camp's "rudely interrupted")
	e.prompt(ega.LightGreen, "A battle begins...")
	e.battleMusic()
	e.combatMapSetup()
	e.placeCombatants()
	e.spr = nil // 0464:248E (1, 4, 3, 0x18): iconToSprite sizes it
	if p := e.partyIDs(); len(p) > 0 {
		at := e.at(p[0])
		e.cview.X, e.cview.Y = at.X-3, at.Y-3
	}
	e.combatUseHook()
	e.as(e.sel)
	if _, err := e.ruleset().Resolve(rules.Fight{}, rules.FightPlaced{}, nil); err != nil {
		panic(fmt.Sprintf("combat setup: %v", err))
	}
	e.setWorldWord(rules.WordBeginFlag, 5)
}

// combatMapSetup = ovr11:1244: the tile sets and the map.
func (e *Engine) combatMapSetup() {
	e.snap.dungeon = e.viewMode == 1
	lays := e.artNames().WildTiles
	if e.viewMode == 1 {
		lays = e.artNames().DungeonTiles
	}
	for _, l := range lays {
		e.LayTiles(l)
	}
	if _, err := e.ruleset().Resolve(rules.Fight{}, e.groundTerms(), nil); err != nil {
		panic(fmt.Sprintf("combat ground: %v", err))
	}
	e.ccursor = freshCursor // a fresh combat map's
}

// groundTerms is what the rules lay the combat ground from.
func (e *Engine) groundTerms() rules.BuildGround {
	g := rules.BuildGround{Walls: e.gmap(), Dungeon: e.viewMode == 1, NoRemains: e.gs.areaKind == 0xA}
	switch e.viewMode {
	case 3:
		g.MapShift = 0xD
	case 4:
		g.MapShift = 0x1A
	}
	return g
}

// placeCombatants = combatBlocks and placeCombatants (ovr11:12E1, :19F5),
// the rules' PlaceFight; an extra with no room leaves, selected as
// RemoveSelected leaves it (the one before it).
func (e *Engine) placeCombatants() {
	ids := e.partyIDs()
	var walls rules.MapSides
	if e.viewMode <= 1 {
		walls = e.gmap()
	}
	o, err := e.ruleset().Resolve(rules.Fight{}, rules.PlaceFight{Walls: walls}, nil)
	if err != nil {
		panic(fmt.Sprintf("combat placement: %v", err))
	}
	for _, f := range o.Effects {
		if l, ok := f.(rules.Left); ok {
			i := slices.Index(ids, l.Unit)
			ids = slices.Delete(ids, i, i+1)
			e.sel = max(i-1, 0)
		}
	}
}
