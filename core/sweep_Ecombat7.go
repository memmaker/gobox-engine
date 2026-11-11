package core

// Additions of sweep Ecombat7 (fold E-R6): what por's combat screens ask of
// the fight (aims, the board, the area preview, the spell menu).

// Square is one square of the fight's ground as the battle view draws it.
type Square struct {
	Tile   int
	Wall   bool // nobody walks onto it
	Height int  // blocking height
}

// Step is one step of a walk: where it ends and whether a foe gets a free
// attack when the walker leaves it.
type Step struct {
	At   BoardPos
	Free bool
}

// MayActFact: the fighter may act against Target (por MayActAgainst); Yes.
type MayActFact struct{ Target CharID }

// RangeToFact is the fighter's distance to Target in squares (por RangeTo); Number.
type RangeToFact struct{ Target CharID }

// SizeFact is the fighter's footprint size on the board (por CombatSize); Number.
type SizeFact struct{}

// FightCardFact is the fighter's turn record (FightCard).
type FightCardFact struct{}

// FightCard is a fighter's turn record: Delay > 0 while it acts this
// round, Moves its half-squares left, Casting the spell number it began
// casting last round (0: none).
type FightCard struct {
	Delay, Moves, Casting int
	Auto                  bool // the computer runs it
}

// AimFact is what an aim of the fighter offers (por playerAim's opening):
// Range 0xFF takes the readied weapon's; FoesOnly offers foes only. AimOffer.
type AimFact struct {
	Range, ShowRange int
	FoesOnly         bool
}

// AimOffer: the resolved Range, the fighters in sight (Near, the order
// they were found in), and those the aim offers (Targets); Hits[i] is the
// chance in percent to hit Targets[i], -1 when not shown.
type AimOffer struct {
	Range   int
	Near    []CharID
	Targets []CharID
	Hits    []int
}

// AimCheckFact is aiming at At from Near (an AimOffer's) (por AimCheck); AimCheck.
type AimCheckFact struct {
	Near             []CharID
	At               BoardPos
	Range, ShowRange int
	Ground           bool // a bare square may be taken
}

// AimCheck: who stands at the square (NoChar: nobody, or a body's owner),
// why it cannot be taken ("" it can), and whether a fighter stands on it.
type AimCheck struct {
	Target   CharID
	Why      string
	Occupied bool
}

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

// BoardUp: the fight's board is laid (Yes).
type BoardUp struct{}

// GroundFact is the fight's ground (Ground); Who is NoChar.
type GroundFact struct{}

// Ground is the fight's 50x25 squares and the squares a footprint of each
// size covers (offsets from its top left).
type Ground struct {
	Squares    [25][50]Square
	Footprints [5]Squares
}

// UnitsFact is the fighters on the board in unit order (Units); Who is NoChar.
type UnitsFact struct{}

// Unit is one fighter on the board.
type Unit struct {
	ID              CharID
	Number          int // unit number, 1-based
	At              BoardPos
	Size, Facing    int
	Party, Helpless bool
	Active          bool // not down
	Icon            int  // its combat icon slot
	Name            string
	HP, MaxHP, AC   int // AC as the sheet shows it
	Damage          string
	Squares         Squares
}

// Units are fighters in unit order.
type Units []Unit

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

// ClearCombat: the fighter's turn record is cleared (por ClearCombatInfo);
// Outcome.TurnOver is ClearCombatInfo's answer.
type ClearCombat struct{ Who CharID }

// CastingTaken: the spell the fighter began goes off now (por: the turn
// record's spell cleared).
type CastingTaken struct{ Who CharID }

// ClassNameFact is the fighter's class as the party card words it; Text.
type ClassNameFact struct{}

func (MayActFact) fact()    {}
func (RangeToFact) fact()   {}
func (SizeFact) fact()      {}
func (FightCardFact) fact() {}
func (AimFact) fact()       {}
func (AimCheckFact) fact()  {}
func (AreaFact) fact()      {}
func (SpellAimFact) fact()  {}
func (GroundFact) fact()    {}
func (UnitsFact) fact()     {}
func (SpellMenuFact) fact() {}
func (ClassNameFact) fact() {}
func (BoardUp) check()      {}
func (FightCard) value()    {}
func (AimOffer) value()     {}
func (AimCheck) value()     {}
func (AreaPreview) value()  {}
func (SpellAim) value()     {}
func (Ground) value()       {}
func (Units) value()        {}
func (SpellMenu) value()    {}
func (ClearCombat) action() {}
func (CastingTaken) action() {}
