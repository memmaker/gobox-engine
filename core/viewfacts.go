package core

// WordFact is the ruleset's word for a character's fact Of, as its sheet
// shows it (por: RaceNames, GenderNames, AlignmentNames, ClassName); the
// value is a Text.
type WordFact struct{ Of PlainFact }

func (WordFact) fact() {}

// Fixture asks the ruleset to set its characters up for a dev scene (por
// DevScene): the records a headless check of that screen needs. Who is the
// selected member, for the fixtures that change one.
type Fixture struct {
	ID  FixtureID
	Who CharID
}

// FixtureID names a dev scene's set-up.
type FixtureID string

const (
	FixtureShop       FixtureID = "shop"       // a shop's stock and unknown items for Id
	FixtureHall       FixtureID = "hall"       // experience and gold for training
	FixtureHallPoor   FixtureID = "hallpoor"   // experience, no gold for the fee
	FixtureTemple     FixtureID = "temple"     // a wounded member and gold for healing
	FixtureAppraise   FixtureID = "appraise"   // gems and jewelry to appraise
	FixtureMissile    FixtureID = "missile"    // missile weapons readied by everyone who may
	FixtureCastFight  FixtureID = "castfight"  // Cure Light Wounds memorized by everyone
	FixtureAutoCombat FixtureID = "autocombat" // the computer fights for everyone
	FixtureWinFight   FixtureID = "winfight"   // the placed foes at 1 hp
	FixtureLoseFight  FixtureID = "losefight"  // the placed foes strong, the party at 1 hp
	FixtureCast       FixtureID = "cast"       // Cure Light Wounds memorized by the able and Who
	FixtureCastKO     FixtureID = "castko"     // as cast, Who unconscious
	FixtureSheetCast  FixtureID = "sheetcast"  // Who has Cure Light Wounds memorized
	FixtureLootDetect FixtureID = "lootdetect" // Who has Detect Magic memorized
	FixtureEffects    FixtureID = "effects"    // spell effects on members 1 and 2
	FixtureInventory  FixtureID = "inventory"  // member 1 fully equipped, many items
	FixtureNoPicker   FixtureID = "nopicker"   // no one can pick a lock
	FixtureDual       FixtureID = "dual"       // member 1 may change class
	FixtureHurt       FixtureID = "hurt"       // member 1 wounded
	FixtureLearn      FixtureID = "learn"      // member 5 ready to rise
	FixturePaladin    FixtureID = "paladin"    // Who a paladin of level 3
	FixtureDrop       FixtureID = "drop"       // Who carries a dagger to drop
	FixtureUsable     FixtureID = "usable"     // Who carries a potion to drink
	FixtureLoot       FixtureID = "loot"       // a fight's items and coins to take
)

func (Fixture) action() {}
