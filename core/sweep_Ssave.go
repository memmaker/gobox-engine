package core

// Additions of sweep Ssave (fold E-R6): what por's save, import and native
// script code asks of the rules. Records of the original's files stay out of
// core and out of the rules: gobox/char reads them (ReadOriginalSave).

type (
	// SpendKnock casts Knock the native way: the first okay member who can
	// spends the spell. The Outcome holds KnockCast, none when nobody could.
	SpendKnock struct{}
	// SetPurseGold sets the party's gold pieces (the tavern game's wager).
	SetPurseGold struct{ Gold uint64 }
	// ForgetParty drops the party and the pool of characters (a load fills
	// them again).
	ForgetParty struct{}
	// RecordView is the engine's view state the next Save records with the
	// area: the running area and the wall slots (a block, a slot; 0xFFFF none).
	RecordView struct {
		AreaKind       uint16
		WallsA, WallsB [3]uint16
	}
	// TakeSavedWalls reads the wall slots a Load brought and forgets them; the
	// Outcome holds SavedWalls.
	TakeSavedWalls struct{}
	// PickKind asks the creation picks (the field of Template.Picks to the
	// option's place) of a kind named as the game's data names it (Race,
	// Gender, Class, Align); the Outcome holds Picked. Error: a name the
	// ruleset does not know.
	PickKind struct{ Race, Gender, Class, Align string }
)

func (SpendKnock) action()     {}
func (SetPurseGold) action()   {}
func (ForgetParty) action()    {}
func (RecordView) action()     {}
func (TakeSavedWalls) action() {}
func (PickKind) action()       {}

type (
	// KnockCast is who spent the Knock.
	KnockCast struct{ Who CharID }
	// SavedWalls are the wall slots of a loaded save (block, slot).
	SavedWalls struct{ A, B [3]uint16 }
	// Picked is the creation picks PickKind asked.
	Picked struct{ Picks map[FieldID]int }
)

func (Picked) effect()     {}
func (KnockCast) effect()  {}
func (SavedWalls) effect() {}
