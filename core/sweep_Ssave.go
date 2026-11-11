package core

// Additions of sweep Ssave (fold E-R6): what por's save, import and native
// script code asks of the rules. Records of the original's files stay out of
// core: the ruleset reads them and answers with typed values.

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
	// ImportOriginal reads an original save image (Files: its file name to its
	// bytes) without touching the ruleset's state; the Outcome holds
	// OriginalImported, the canonical form of it. Error: not a save.
	ImportOriginal struct{ Files map[string][]byte }
)

func (SpendKnock) action()     {}
func (SetPurseGold) action()   {}
func (ForgetParty) action()    {}
func (RecordView) action()     {}
func (TakeSavedWalls) action() {}
func (ImportOriginal) action() {}
func (PickKind) action()       {}

type (
	// KnockCast is who spent the Knock.
	KnockCast struct{ Who CharID }
	// SavedWalls are the wall slots of a loaded save (block, slot).
	SavedWalls struct{ A, B [3]uint16 }
	// Picked is the creation picks PickKind asked.
	Picked struct{ Picks map[FieldID]int }
	// OriginalImported is an original save read: the ruleset's canonical part
	// (Rules, for Load), what the engine keeps (positions, the party's file
	// names as stored) and the record words by index, for the script fields
	// that live in the records (the engine knows each record's address base).
	OriginalImported struct {
		Rules                    []byte
		Set                      byte
		X, Y, Facing, Pos3, Pos4 byte
		ViewMode, BeginFlag      byte
		AreaKind                 uint16
		Party                    []string
		Area, Game, Aux          []uint16
	}
)

func (Picked) effect()           {}
func (KnockCast) effect()        {}
func (SavedWalls) effect()       {}
func (OriginalImported) effect() {}
