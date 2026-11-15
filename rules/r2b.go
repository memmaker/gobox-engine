package rules

// Looks in flight: what the ruleset says a spell or a thrown or shot item is,
// so the engine picks its sprite, sound and frames from the kind alone.

// FlightKind is what a missile is in flight.
type FlightKind int

const (
	FlightDefault    FlightKind = iota // a spell or item with no art of its own
	FlightGaze                         // a gaze
	FlightSpit                         // spat acid
	FlightBolt                         // a breath or a lightning bolt
	FlightPointed                      // a dart, javelin, quarrel, spear or arrow
	FlightTumbling                     // a hand axe, club or hammer
	FlightFlask                        // holy water or oil
	FlightSlingStone                   // a sling's stone
)

// FlightFact asks how a missile looks: Spell when it is a spell or power
// (Item is then ignored), else Item. The Fact's own id is NoChar.
type FlightFact struct {
	Spell SpellID
	Item  ItemID
}

// Flight is FlightFact's answer.
type Flight struct{ Kind FlightKind }

func (FlightFact) fact() {}

func (Flight) value() {}

// BeastKindFact asks the bestiary kind of a monster name ("Goblin" for
// "GOBLIN LEADER"). The Fact's own id is NoChar.
type BeastKindFact struct{ Monster string }

// BeastKind is BeastKindFact's answer; Name "": the monster has no kind.
type BeastKind struct{ Name string }

func (BeastKindFact) fact() {}

func (BeastKind) value() {}
