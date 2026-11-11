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
type FixtureID int8

const (
	FixtureShop       FixtureID = iota + 1 // a shop's stock and unknown items for Id
	FixtureHall                            // experience and gold for training
	FixtureHallPoor                        // experience, no gold for the fee
	FixtureTemple                          // a wounded member and gold for healing
	FixtureAppraise                        // gems and jewelry to appraise
	FixtureMissile                         // missile weapons readied by everyone who may
	FixtureCastFight                       // Cure Light Wounds memorized by everyone
	FixtureAutoCombat                      // the computer fights for everyone
	FixtureWinFight                        // the placed foes at 1 hp
	FixtureLoseFight                       // the placed foes strong, the party at 1 hp
	FixtureCast                            // Cure Light Wounds memorized by the able and Who
	FixtureCastKO                          // as cast, Who unconscious
	FixtureSheetCast                       // Who has Cure Light Wounds memorized
	FixtureLootDetect                      // Who has Detect Magic memorized
	FixtureEffects                         // spell effects on members 1 and 2
	FixtureInventory                       // member 1 fully equipped, many items
	FixtureNoPicker                        // no one can pick a lock
	FixtureDual                            // member 1 may change class
	FixtureHurt                            // member 1 wounded
	FixtureLearn                           // member 5 ready to rise
	FixturePaladin                         // Who a paladin of level 3
	FixtureDrop                            // Who carries a dagger to drop
	FixtureUsable                          // Who carries a potion to drink
	FixtureLoot                            // a fight's items and coins to take
)

func (Fixture) action() {}

// ItemHandle is a placed item the ruleset holds, as CharID is a character:
// handed out by the ruleset, stored by the engine (core.ItemID is the base
// item it is made from).
type ItemHandle int32

// NoItem is no item.
const NoItem ItemHandle = 0
