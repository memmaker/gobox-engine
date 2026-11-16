package engine

import (
	"fmt"
	"gobox/game"
	"gobox/rules"
	"slices"
)

// combatSetup = ovr11:1E31.
func (e *Engine) combatSetup() {
	e.as(e.sel)
	e.world.Morale = min(e.world.Morale, moraleCap) // ovr11:1E31
	e.world.AttackedAlly = 0
	f := e.fight()
	f.Board = rules.Board{Straight: e.Deskew.Load()}
	e.ruleset().BeginFight(f, nil)
	e.FreeAnimPic()
	e.port = portrait{headID: 0xFF, bodyID: 0xFF} // 0464:2595 x2
	e.clearPrompt()
	e.ext.note = "" // extra's note goes with the row the original clears (camp's "rudely interrupted")
	e.prompt(game.InkLightGreen, "A battle begins...")
	e.battleMusic()
	e.combatMapSetup()
	e.placeCombatants()
	e.loadMemberIcons()
	e.spr = nil // 0464:248E (1, 4, 3, 0x18): iconToSprite sizes it
	e.combatUseHook()
	e.as(e.sel)
	if _, err := e.ruleset().Resolve(e.fight(), rules.FightPlaced{}, nil); err != nil {
		panic(fmt.Sprintf("combat setup: %v", err))
	}
	e.setPhase(rules.PhaseFight)
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
	e.layPlaceTiles()
	if _, err := e.ruleset().Resolve(e.fight(), e.groundTerms(), nil); err != nil {
		panic(fmt.Sprintf("combat ground: %v", err))
	}
}

// layPlaceTiles is LoadCombatTileSet's sheet by the place (the game's PlaceTiles): the arena's tiles are a fresh set with
// the sheet laid in it, in the sheet's own colours.
func (e *Engine) layPlaceTiles() {
	traits, _ := e.placeTraits()
	for _, l := range e.artNames().PlaceTiles {
		if !l.Fits(traits) {
			continue
		}
		im := e.LoadPic(l.File, l.Block)
		if im == nil {
			panic(fmt.Sprintf("combat tiles: no %s block %d in the art", l.File, l.Block))
		}
		e.sqrpaci[0] = blankTiles()
		e.sqrpaci[0].Palette = im.Palette
		e.LayTiles(l)
		return
	}
}

// loadMemberIcons loads each party unit's whole-picture icon (saveload.pas LoadCharIconAnim: the sheet's block the record's
// icon names, into the slot of its orderNumber); a game that composes its members' icons has no sheet.
func (e *Engine) loadMemberIcons() {
	sheet := e.artNames().MemberIcons
	if sheet == "" {
		return
	}
	for _, u := range e.fact(rules.NoChar, rules.UnitsFact{}).(rules.Units) {
		if u.Party {
			e.LoadIconPic(sheet, e.fact(u.ID, rules.IconPicsFact{}).(rules.IconPics).Body, u.Icon)
		}
	}
}

// groundTerms is what the rules lay the combat ground from.
func (e *Engine) groundTerms() rules.BuildGround {
	g := rules.BuildGround{Walls: e.gmap(), Dungeon: e.viewMode == 1, NoRemains: e.game().NoRemains(int(e.gs.areaKind)), MapX: e.world.MapX, MapY: e.world.MapY, Boss: e.world.Boss}
	g.Block = int32(e.gs.areaKind) // the ECL block in play (the field's name is older than its use)
	g.Place, _ = e.placeTraits()
	if r, ok := e.layoutField(fieldWildTerrain); ok {
		g.Terrain = rules.TerrainID(e.host.Get(r))
	}
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
	o, err := e.ruleset().Resolve(e.fight(), rules.PlaceFight{Walls: walls}, nil)
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
