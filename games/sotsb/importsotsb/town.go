package importsotsb

import "rules1e"

// Town is Silver Blades' town economy for the shared engine (rules.json town):
// training and the temple's ten services are free (no Pay in train.go or
// temple.go). Each service says what it did (Done) and why it did not help
// (Not) in the temple's own words (temple.go), and only says so: it never
// offers to cast anyway (Declines). The wound services help only the hurt (rules1e TempleKind.Wound).
func Town() rules1e.TownData {
	const nothingToHeal = "has no need of healing"
	svc := func(i int, k rules1e.TempleKind, not, done string) rules1e.TempleService {
		return rules1e.TempleService{Name: templeServices[i], Kind: k, Not: not, Done: done, Declines: true}
	}
	wound := func(i int, k rules1e.TempleKind, done string) rules1e.TempleService {
		return svc(i, k, nothingToHeal, done)
	}
	return rules1e.TownData{Temple: []rules1e.TempleService{
		svc(0, rules1e.TempleBlindness, "is not blind", "can see!"),
		svc(1, rules1e.TempleDisease, "is not Diseased", "is cured"),
		wound(2, rules1e.TempleLight, "is healed"),
		wound(3, rules1e.TempleSerious, "is healed"),
		wound(4, rules1e.TempleCritical, "is healed"),
		wound(5, rules1e.TempleHeal, "is fully healed"),
		svc(6, rules1e.TemplePoison, "is not poisoned", "is no longer poisoned"),
		svc(7, rules1e.TempleRaiseDead, "is not dead yet", "is ressurected"),
		svc(8, rules1e.TempleRemoveCurse, "is not cursed", "rejoices.  The curse has been lifted!"),
		svc(9, rules1e.TempleStone, "is not petrified", "is turned to flesh"),
	}}
}
