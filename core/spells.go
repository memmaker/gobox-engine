package core

// Spells: aims, the spell menu, books, study and Knock.

// AreaFact is the preview of an area aimed at At (Spell: the spell's,
// "" a maneuver's burst of Area squares); the fighter is the caster. AreaPreview.
type AreaFact struct {
	Spell SpellID
	At    BoardPos
	Area  int
}

// AreaPreview: the squares of those it hits (Hit, in order), the squares
// it covers (Area), and allies it leaves out (Skipped).
type AreaPreview struct{ Hit, Area, Skipped Squares }

// SpellAimFact is how the fighter's Spell aims (SpellAim).
type SpellAimFact struct{ Spell SpellID }

// SpellAimKind is how a spell is aimed in a fight.
type SpellAimKind int32

const (
	SpellAimNone    SpellAimKind = iota // cast without an aim
	SpellAimOne                         // one target, no bare square
	SpellAimArea                        // one square, Area squares round it
	SpellAimTargets                     // up to Targets different targets
)

// SpellAim: how the spell aims, Range its squares, FoesOnly offers foes
// only, Now: cast this turn (not begun for the next round).
type SpellAim struct {
	Name          string
	Kind          SpellAimKind
	Area, Targets int
	Range         int
	FoesOnly, Now bool
}

// SpellMenuFact is the fighter's spell list of Kind (por buildSpellList's
// kinds: 0 memorized, 1 the book, 2 the scroll in hand (Scroll), 3/6 all
// scrolls carried (6: still to scribe), 4 the learnable, 5 to memorize);
// SpellMenu. Listing a scroll reveals it to a reader (as the original).
type SpellMenuFact struct {
	Kind   int
	Scroll ItemHandle
}

// SpellLine is a line of the spell menu: a header (Header: its group's
// name, Planned/Slots its level's plan and slots, -1 none) or a spell
// (Number, Name, Memo: to be memorized, Count: on the plan, Item: the
// scroll it is on).
type SpellLine struct {
	Header         bool
	Text           string
	Number         int
	Memo           bool
	Count          int
	Planned, Slots int
	Item           ItemHandle
}

// SpellMenu is spell lines in order.
type SpellMenu []SpellLine

func (AreaFact) fact() {}

func (SpellAimFact) fact() {}

func (SpellMenuFact) fact() {}

func (AreaPreview) value() {}

func (SpellAim) value() {}

func (SpellMenu) value() {}

// StudyFact is what Who could study in camp (Study).
type StudyFact struct{}

// Study counts the spells of Who's book and those on the scrolls it carries.
type Study struct{ Book, Scrolls int32 }

func (Study) value() {}

func (StudyFact) fact() {}

// SpellNumber asks the spell's number in the original's list (Value; Yes: it has one).
type SpellNumber struct{ Spell SpellID }

func (SpellNumber) check() {}

// SpendKnock casts Knock the native way: the first okay member who can
// spends the spell. The Outcome holds KnockCast, none when nobody could.
type SpendKnock struct{}

func (SpendKnock) action() {}

// KnockCast is who spent the Knock.
type KnockCast struct{ Who CharID }

func (KnockCast) effect() {}

// LearnableFact is the spells Who may learn in the hall (SpellCards).
type LearnableFact struct{}

// BookFact is Who's spell book (Book).
type BookFact struct{}

func (LearnableFact) fact() {}

func (BookFact) fact() {}

// ScrollSpell is one spell on a scroll.
type ScrollSpell struct {
	Spell    SpellID
	Number   int32 // as the game numbers it
	Scribing bool
}

// Book is a character's spells for the roster cards: the book's, those in
// the head, the slots a class has per level (class 0 cleric, 1 mage; level 1..3
// at index 0..2) and the level of each class column.
type Book struct {
	Known, Memorized []BookSpell
	Slots            [2][3]int32
	Levels           []int32
}

// BookSpell is a spell of a book as the engine numbers it (Number, the game's
// own spell numbers) with its class and level, how many are planned and
// whether it is still to be memorized (Memorized lists only).
type BookSpell struct {
	Spell        SpellID
	Number       int32
	Name         string
	Class, Level int32
	Planned      int32
}

// SpellCard is a spell to pick: its name and level.
type SpellCard struct {
	Spell  SpellID
	Number int32 // as the game numbers it
	Name   string
	Level  int32
}

// SpellCards are spells in an order.
type SpellCards []SpellCard

func (SpellCards) value() {}

func (Book) value() {}

// CastAim is how a spell out of a fight finds its target before it is cast.
type CastAim int32

const (
	CastAimSelf       CastAim = iota // no question (the caster, the party, or none)
	CastAimCombatOnly                // a combat-only spell or item: "Lose it?" / "Use it?"
	CastAimMember                    // a member it goes on: "Cast Spell on whom"
)

// CastCardFact is spell number Number as a cast out of a fight sees it
// (CastCard); Who is NoChar. Number is the ruleset's spell number (SpellLine.Number).
type CastCardFact struct{ Number int }

// CastCard: the spell's Name ("" for an item's power) and its Aim.
type CastCard struct {
	Name string
	Aim  CastAim
}

// OpenCast: Who opens camp's Cast (por CastMenu's start); refused
// (ErrNotNow) when Who cannot cast here, as the rules word it to the party.
type OpenCast struct{ Who CharID }

// CampCast: Who casts memorized spell Spell out of a fight; Sure answers
// a combat-only spell's "Lose it?", Target is the member it goes on (NoChar: none).
type CampCast struct {
	Who, Target CharID
	Spell       SpellID
	Sure        bool
}

func (CastCardFact) fact() {}
func (CastCard) value()    {}
func (OpenCast) action()   {}
func (CampCast) action()   {}
