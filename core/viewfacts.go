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
	FixtureShop     FixtureID = "shop"     // a shop's stock and unknown items for Id
	FixtureHall     FixtureID = "hall"     // experience and gold for training
	FixtureHallPoor FixtureID = "hallpoor" // experience, no gold for the fee
	FixtureTemple   FixtureID = "temple"   // a wounded member and gold for healing
	FixtureAppraise FixtureID = "appraise" // gems and jewelry to appraise
)

func (Fixture) action() {}
