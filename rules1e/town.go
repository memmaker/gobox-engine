package rules1e

// A game's town economy: what the training hall charges and which services
// the temple sells at what price. One rule reads it; each game's importer
// writes its own values into rules.json ("town"); POR's are in the POR literal.
type TownData struct {
	TrainPrice int32           // gold the hall takes for a level (0: it trains free)
	Temple     []TempleService // the temple's services in the menu's order
	Bundles    bool            `json:",omitempty"` // a shop's bundle (an item with a count above 1) costs its count times the price (pod)
}

// TempleService is one temple service: what it does (Kind), the name the
// menu shows, the price in gold (0: free), and what the temple says of a
// member it would not help (Not; "": it always helps). Done is what it says of a member
// it helped ("" : "is cured."). Declines: it only says Not of a member it would not
// help and does not offer to cast anyway (Silver Blades). The wound cures (Wound) help
// only a member who needs them and cannot be chosen for another (Felix 2026-11-12).
type TempleService struct {
	Name          string
	Kind          TempleKind
	Price         int32
	Not           string
	Done          string `json:",omitempty"`
	Declines      bool   `json:",omitempty"`
	RestoresStats bool   `json:",omitempty"` // a disease cure puts the strength back (the weakening ends; pod's Cure Disease runs statAffects)
}

// TempleKind is what a temple service does.
type TempleKind string

const (
	TempleBlindness   TempleKind = "blindness"   // cure blindness
	TempleDisease     TempleKind = "disease"     // cure disease
	TempleLight       TempleKind = "light"       // cure light wounds
	TempleSerious     TempleKind = "serious"     // cure serious wounds
	TempleCritical    TempleKind = "critical"    // cure critical wounds
	TempleHeal        TempleKind = "heal"        // wounds down to a few points, blindness and disease
	TemplePoison      TempleKind = "poison"      // neutralize poison
	TempleRaiseDead   TempleKind = "raiseDead"   // raise the dead
	TempleRemoveCurse TempleKind = "removeCurse" // remove the curse of a member or an item
	TempleStone       TempleKind = "stone"       // stone to flesh
	TempleRepair      TempleKind = "repair"      // every member at once: all the others
)

// Wound: k is a wound cure (light, serious, critical, heal).
func (k TempleKind) Wound() bool {
	return k == TempleLight || k == TempleSerious || k == TempleCritical || k == TempleHeal
}

// porTown is POR's town (ovr17:29AC's fee, DS:00F1's services and ovr05's prices and refusals).
var porTown = TownData{TrainPrice: 1000, Temple: []TempleService{
	{Name: "Cure Blindness", Kind: TempleBlindness, Price: 1000, Not: "is not blind."},
	{Name: "Cure Disease", Kind: TempleDisease, Price: 1000, Not: "is not Diseased."},
	{Name: "Cure Light Wounds", Kind: TempleLight, Price: 100},
	{Name: "Cure Serious Wounds", Kind: TempleSerious, Price: 350},
	{Name: "Cure Critical Wounds", Kind: TempleCritical, Price: 600},
	{Name: "Neutralize Poison", Kind: TemplePoison, Price: 1000, Not: "is not poisoned."},
	{Name: "Raise Dead", Kind: TempleRaiseDead, Price: 5500, Not: "is not dead."},
	{Name: "Remove Curse", Kind: TempleRemoveCurse, Price: 3500, Not: "is not cursed."},
	{Name: "Stone to Flesh", Kind: TempleStone, Price: 2000, Not: "is not stoned."},
}}

// templeService is temple service i; false when the temple has no such.
func (e *RuleState) templeService(i int) (TempleService, bool) {
	t := e.T().Town.Temple
	if i < 0 || i >= len(t) {
		return TempleService{}, false
	}
	return t[i], true
}
