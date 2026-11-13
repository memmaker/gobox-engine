package importer

import (
	"gobox/rules1e"
)

// Town is Pools of Darkness' town economy for the shared engine (rules.json
// town): training and the temple are free; the temple's eleven services are the
// spells of templeServiceSpells, and Repair mends the whole party at once
// (templeRepair).
func PodTown() rules1e.TownData {
	kinds := [len(podTempleServices)]rules1e.TempleKind{rules1e.TempleBlindness, rules1e.TempleDisease, rules1e.TempleLight,
		rules1e.TempleSerious, rules1e.TempleCritical, rules1e.TempleHeal, rules1e.TemplePoison, rules1e.TempleRaiseDead,
		rules1e.TempleRemoveCurse, rules1e.TempleStone}
	nots := map[rules1e.TempleKind]string{rules1e.TempleBlindness: "is not blind.", rules1e.TempleDisease: "is not Diseased",
		rules1e.TemplePoison: "is not poisoned.", rules1e.TempleRaiseDead: "is not dead.", rules1e.TempleRemoveCurse: "is not cursed.",
		rules1e.TempleStone: "is not stoned."}
	t := rules1e.TownData{Bundles: true} // shopPrice: a bundle costs its count times the price
	for i, name := range podTempleServices {
		t.Temple = append(t.Temple, rules1e.TempleService{Name: name, Kind: kinds[i], Not: nots[kinds[i]], RestoresStats: kinds[i] == rules1e.TempleDisease}) // O0BB:2F08's cure runs statAffects(Str) after the weakening goes
	}
	t.Temple = append(t.Temple, rules1e.TempleService{Name: "Repair", Kind: rules1e.TempleRepair})
	return t
}
