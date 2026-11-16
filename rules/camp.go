package rules

// Camp: resting, healing, studying (what por's camp asks of the rules).

// HealMinutes asks the minutes of rest until everyone that resting can
// heal is whole (Answer.Value); 0 when nobody can be.
type HealMinutes struct{}

// PartyStudy asks the longest study time of the party in minutes
// (Answer.Value), the study hours of each member set on the way.
type PartyStudy struct{}

// StudyNeeded asks the longest study a member still has to do, in minutes
// (Answer.Value); it sets nothing, so it may be asked while a rest runs.
type StudyNeeded struct{}

// HealClerics asks how many members cast Cure Light Wounds (Answer.Value).
type HealClerics struct{}

// SpellsAllowed asks whether the area allows spells (Yes); NoMagic: it does not (the area's word).
type SpellsAllowed struct{ NoMagic bool }

// RestUnit asks the rest time left in clock unit Unit (1 minutes, 2 tens
// of minutes, 3 hours, 4 days) as Answer.Value.
type RestUnit struct{ Unit int32 }

// SkyDay asks whether the sky, inside or out, is a day sky (Yes). The sky is
// the area's words: In3D (a 3D area, else the overland), Sky, and SkyInside
// for a cell inside.
type SkyDay struct {
	Inside         bool
	In3D           bool
	Sky, SkyInside uint16
}

// SkyInk asks the colour of the 3D view's sky, as an EGA colour number in Answer.Value (Yes: the ruleset has one; the Pool's
// sky is a picture or a flat). Inside, In3D, Sky and SkyInside are SkyDay's.
type SkyInk struct {
	Inside         bool
	In3D           bool
	Sky, SkyInside uint16
}

// RestDone asks whether the rest is over before its next step (Yes): its
// time is up, a stop was asked or, UntilHealed, nobody is wounded.
type RestDone struct{ UntilHealed bool }

// CampOpen starts camp's rest state over: no healing steps, no rest time. A
// rest's encounter check comes every RestEvery steps (0: never) with
// RestChance percent; the rules keep them until camp closes.
type CampOpen struct{ RestEvery, RestChance uint16 }

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

// CastCures has the clerics cast Cure Light Wounds on the most wounded
// (none where NoMagic: the area forbids spells).
type CastCures struct{ NoMagic bool }

// SetRest sets the rest to Minutes (no rest time when 0).
type SetRest struct{ Minutes int32 }

// BeginRest starts a rest: study steps and effect ticks start over.
type BeginRest struct{}

// StopRest has the rest end at its next step.
type StopRest struct{}

func (HealMinutes) check() {}

func (PartyStudy) check()  {}
func (StudyNeeded) check() {}

func (HealClerics) check() {}

func (SpellsAllowed) check() {}

func (RestUnit) check() {}

func (SkyDay) check() {}

func (SkyInk) check() {}

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
