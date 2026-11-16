package rules

// Looks in flight: what the ruleset says a spell or a thrown or shot item is,
// so the engine picks its sprite, sound and frames from the kind alone.

// FlightKind is what a missile is in flight.
type FlightKind int

const (
	FlightDefault      FlightKind = iota // a spell or item with no art of its own
	FlightGaze                           // a gaze
	FlightSpit                           // spat acid
	FlightBolt                           // a breath or a lightning bolt
	FlightPointed                        // a dart, javelin, quarrel, spear or arrow
	FlightTumbling                       // a hand axe, club or hammer
	FlightFlask                          // holy water or oil
	FlightSlingStone                     // a sling's stone
	FlightHeatBolt                       // a heat gun's or plasma thrower's shot (cbtacts.pas AnimateMissile: icon 13)
	FlightLaser                          // a laser pistol's or rifle's beam (icon 16)
	FlightCrossbowBolt                   // a D.R. X-bow's quarrel (icon 19)
	FlightGunBolt                        // a bolt gun's or needle gun's shot (icon 22)
	FlightRocket                         // a rocket pistol's, rifle's or launcher's rocket (icon 25)
	FlightWave                           // a microwave gun's or sonic stunner's wave (icon 30)
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
