package core

// Additions of sweep Scamp: what por's camp (camp_engine.go) asks of the rules.
// Time of day is IsDay and WordClockMinutes; camp up is WordCampOn.

type (
	// HealMinutes asks the minutes of rest until everyone that resting can
	// heal is whole (Answer.Value); 0 when nobody can be.
	HealMinutes struct{}
	// PartyStudy asks the longest study time of the party in minutes
	// (Answer.Value), the study hours of each member set on the way.
	PartyStudy struct{}
	// HealClerics asks how many members cast Cure Light Wounds (Answer.Value).
	HealClerics struct{}
	// SpellsAllowed asks whether the area allows spells (Yes).
	SpellsAllowed struct{}
	// RestUnit asks the rest time left in clock unit Unit (1 minutes, 2 tens
	// of minutes, 3 hours, 4 days) as Answer.Value.
	RestUnit struct{ Unit int32 }
	// SkyDay asks whether the sky, inside or out, is a day sky (Yes).
	SkyDay struct{ Inside bool }
	// RestDone asks whether the rest is over before its next step (Yes): its
	// time is up, a stop was asked or, UntilHealed, nobody is wounded.
	RestDone struct{ UntilHealed bool }
)

type (
	// CampOpen starts camp's rest state over: no healing steps, no rest time.
	CampOpen struct{}
	// ClearStudy has nobody memorize or scribe any more.
	ClearStudy struct{}
	// CampClose ends camp: the plans hold (the next rest learns them) and
	// nobody is the caster.
	CampClose struct{}
	// PlanParty takes up every member's planned spells to be memorized.
	PlanParty struct{}
	// ForgetCures has the clerics forget their memorized Cure Light Wounds
	// and take up their planned spells.
	ForgetCures struct{}
	// MemorizeCures has the clerics memorize Cure Light Wounds in every
	// first-level slot.
	MemorizeCures struct{}
	// CastCures has the clerics cast Cure Light Wounds on the most wounded.
	CastCures struct{}
	// SetRest sets the rest to Minutes (no rest time when 0).
	SetRest struct{ Minutes int32 }
	// BeginRest starts a rest: study steps and effect ticks start over.
	BeginRest struct{}
	// StopRest has the rest end at its next step.
	StopRest struct{}
)

func (HealMinutes) check()   {}
func (PartyStudy) check()    {}
func (HealClerics) check()   {}
func (SpellsAllowed) check() {}
func (RestUnit) check()      {}
func (SkyDay) check()        {}
func (RestDone) check()      {}

func (CampOpen) action()      {}
func (ClearStudy) action()    {}
func (CampClose) action()     {}
func (PlanParty) action()     {}
func (ForgetCures) action()   {}
func (MemorizeCures) action() {}
func (CastCures) action()     {}
func (SetRest) action()       {}
func (BeginRest) action()     {}
func (StopRest) action()      {}
