package importer

import (
	"os"
	"path/filepath"

	"gobox/ecl"
)

// brLayout is BR1's and BR2's partial script layout (ecl.PartialLayout):
// the zone bytes the engine reads by name at fight setup (core.FightTerms),
// at the same GBVM address in both games (buckroger-re and buckrogers2-re
// types.pas TZone0, TZone1). Every other address stays raw.
const brLayout = `# Buck Rogers script layout (gobox ecl/script Layout, Partial): from len width owner name.
4BE6 1 byte game AreaKind  # zone0 areaKind (BR2 areaType): $A2, $A8 Venus or Mars wilderness -> FightTerms.Area
7EE3 1 byte game SkipsCombatItems  # zone1: 1 keeps the fallen's items from treasure -> FightTerms.NoItemLoot
7EE6 1 byte game RevivesParty  # zone1: non-zero stands the downed party up with 1 HP -> FightTerms.RevivesParty
7F6D 1 byte game PriceMultiplier  # zone1 priceMultiplier: sixteenths of an item's value a shop asks, 0 the value -> Shopping.Terms.PriceMultiplier
`

// WriteBRLayout writes brLayout into a scripts folder (the adventure's, the
// run-time cache): before the scripts' import, which reads it.
func WriteBRLayout(scripts string) error {
	p := filepath.Join(scripts, ecl.PartialLayout)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, []byte(brLayout), 0o644)
}
