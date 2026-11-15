package rules

// Phase is which screen or loop the game is in (por DS:49F3). The engine owns it
// and hands each change to the ruleset (EnterPhase), which keeps what it was last
// handed: some of its rules differ in a fight.
type Phase uint8

const (
	PhaseMenu       Phase = iota // the party menu
	PhaseShop                    // a shop or temple
	PhaseCamp                    // camp
	PhaseExploring               // exploring
	PhaseAreaStart               // an area starting
	PhaseFight                   // a fight (its setup too)
	PhaseAfterFight              // after a fight, its results
	PhaseFinale                  // the finale's pictures: nothing sets it here, a save of the original may hold it
)

// EnterPhase hands the ruleset the phase the game is in now.
type EnterPhase struct{ Phase Phase }

// MemberCount asks the party's count in the game's record (Answer.Value): the
// members the record counts. PartySize is the list's length, which includes the
// extras a fight adds.
type MemberCount struct{}

// SetMemberCount stores the record's count of the party.
type SetMemberCount struct{ N int32 }

// ActFor makes the party's member Member (an index) the one the rules act for.
type ActFor struct{ Member int32 }

// ActingFor asks the member the rules act for (Answer.Value, an index).
type ActingFor struct{}

// PartyEffect asks entry Index of the party's affect table (Answer.Value; Yes:
// the entry exists).
type PartyEffect struct{ Index int32 }

// OpenHallForAll opens the training hall with nothing asked of a member: the
// game won, the hall teaches every class (adventure.EffectOpenHall).
type OpenHallForAll struct{}

func (EnterPhase) action()     {}
func (MemberCount) check()     {}
func (SetMemberCount) action() {}
func (ActFor) action()         {}
func (ActingFor) check()       {}
func (PartyEffect) check()     {}
func (OpenHallForAll) action() {}
