package rules

// PlaceTraits is what a ruleset tests of the place a fight is in: the canonical value of a game's area-kind
// field (a layout's enc= maps the game's byte to and from it). The zero value is a place with no kind at all.
// Each trait names the Buck Rogers combatst.pas / cbtacts.pas test it stands for (zone0 areaKind).
type PlaceTraits uint8

const (
	// TraitHasKind: areaKind <> 0 (charhlp.pas:1417, eclmenu.pas:416: the facing is doubled).
	TraitHasKind PlaceTraits = 1 << iota
	// TraitWilderness: areaKind in [VENUS_WILDERNESS $A2, MARS_WILDERNESS $A8] (combat.pas:387 the jetpack,
	// combatst.pas:628 the wilderness arena).
	TraitWilderness
	// TraitZeroGravity: areaKind and ZERO_GRAVITY $10 (cbtacts.pas:1063, 1990; combatai.pas:656).
	TraitZeroGravity
	// TraitLowTech: areaKind = $88 or $82, the 'lotek' dungeon: floor terrain 7, not 45 (combatst.pas:314).
	TraitLowTech
	// TraitScatterTiles: areaKind and $40, the dungeon arena scatters tile groups (combatst.pas:318).
	TraitScatterTiles
	// TraitSecondTileSet: areaKind and 8, the dungeon arena's second wall tile set (combatst.pas:85).
	TraitSecondTileSet
	// TraitTerrainShift: areaKind and $F = 8, the wilderness terrain is 10 higher (combatst.pas:485).
	TraitTerrainShift
)

// Has reports whether every trait in t is set.
func (p PlaceTraits) Has(t PlaceTraits) bool { return p&t == t }
