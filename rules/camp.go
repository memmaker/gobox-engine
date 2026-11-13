package rules

// Camp: resting, healing, studying (what por's camp asks of the rules).

// HealMinutes asks the minutes of rest until everyone that resting can
// heal is whole (Answer.Value); 0 when nobody can be.
type HealMinutes struct{}

// PartyStudy asks the longest study time of the party in minutes
// (Answer.Value), the study hours of each member set on the way.
type PartyStudy struct{}

// HealClerics asks how many members cast Cure Light Wounds (Answer.Value).
type HealClerics struct{}

// SpellsAllowed asks whether the area allows spells (Yes).
type SpellsAllowed struct{}

// RestUnit asks the rest time left in clock unit Unit (1 minutes, 2 tens
// of minutes, 3 hours, 4 days) as Answer.Value.
type RestUnit struct{ Unit int32 }

// SkyDay asks whether the sky, inside or out, is a day sky (Yes).
type SkyDay struct{ Inside bool }

// RestDone asks whether the rest is over before its next step (Yes): its
// time is up, a stop was asked or, UntilHealed, nobody is wounded.
type RestDone struct{ UntilHealed bool }

// CampOpen starts camp's rest state over: no healing steps, no rest time.
type CampOpen struct{}

// ClearStudy has nobody memorize or scribe any more.
type ClearStudy struct{}

// CampClose ends camp: the plans hold (the next rest learns them) and
// nobody is the caster.
type CampClose struct{}

// PlanParty takes up every member's planned spells to be memorized.
type PlanParty struct{}

// ForgetCures has the clerics forget their memorized Cure Light Wounds
// and take up their planned spells.
type ForgetCures struct{}

// MemorizeCures has the clerics memorize Cure Light Wounds in every
// first-level slot.
type MemorizeCures struct{}

// CastCures has the clerics cast Cure Light Wounds on the most wounded.
type CastCures struct{}

// SetRest sets the rest to Minutes (no rest time when 0).
type SetRest struct{ Minutes int32 }

// BeginRest starts a rest: study steps and effect ticks start over.
type BeginRest struct{}

// StopRest has the rest end at its next step.
type StopRest struct{}

func (HealMinutes) check() {}

func (PartyStudy) check() {}

func (HealClerics) check() {}

func (SpellsAllowed) check() {}

func (RestUnit) check() {}

func (SkyDay) check() {}

func (RestDone) check() {}

func (CampOpen) action() {}

func (ClearStudy) action() {}

func (CampClose) action() {}

func (PlanParty) action() {}

func (ForgetCures) action() {}

func (MemorizeCures) action() {}

func (CastCures) action() {}

func (SetRest) action() {}

func (BeginRest) action() {}

func (StopRest) action() {}
