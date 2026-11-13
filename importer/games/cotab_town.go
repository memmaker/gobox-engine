package games

import "gobox/rules1e"

// Town is cotab's town economy for the shared engine (rules.json town): the
// hall's fee and the temple's ten priced services (templeServices, templePrices,
// templeNot), as the engine's kinds.
func CotabTown() rules1e.TownData {
	kinds := [len(cotabTempleServices)]rules1e.TempleKind{rules1e.TempleBlindness, rules1e.TempleDisease, rules1e.TempleLight,
		rules1e.TempleSerious, rules1e.TempleCritical, rules1e.TempleHeal, rules1e.TemplePoison, rules1e.TempleRaiseDead,
		rules1e.TempleRemoveCurse, rules1e.TempleStone}
	t := rules1e.TownData{TrainPrice: 1000}
	for i, name := range cotabTempleServices {
		t.Temple = append(t.Temple, rules1e.TempleService{Name: name, Kind: kinds[i], Price: int32(templePrices[i]), Not: templeNot[i]})
	}
	return t
}
