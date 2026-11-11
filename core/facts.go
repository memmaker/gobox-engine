package core

// Fact is one thing about a character that a script can read or write. The
// set is closed and small because the scripts' use is: see the per-game table
// in ARCHITECTURE.md. Facts compare with ==, so Catalog can list them.
type Fact interface{ fact() }

type (
	AbilityFact struct{ Ability Ability } // por 0x18 con; cotab cvIntelligence
	SkillFact   struct{ Skill Skill }     // por 0xA5..0xAC; BR 76+s
	SaveFact    struct{ Save Save }       // por 0x9B
	LevelFact   struct{ Class Class }     // cotab cvMageLevel
	CoinFact    struct{ Coin Coin }       // the party's one purse, whoever is selected. por 0xBB..0xC3, also computed (area 9)
	KnownFact   struct{ Skill Skill }     // has the skill at all: BR 76 reads 128 or 0
	MemoFact    struct{ Slot int }        // por 0x1F..0x70: Knock in area 14
	PlainFact   int                       // the rest, one each
)

// The plain facts.
const (
	FactName PlainFact = iota + 1
	FactRace
	FactClasses
	FactSex
	FactAlignment
	FactActive      // 256: a pure read; NoChar reads inactive
	FactStatus      // por 0x100; writing it only stores: consequences are Actions
	FactSide        // 268 bit 0
	FactSideLock    // BR 268 bit 7, r[215]; meaning to confirm at migration
	FactControl     // por 0xB8 (charmed, hired)
	FactMove        // por 0x11B
	FactHP          // BR 281
	FactMaxHP       // BR 69; sotsb cvMaxHP
	FactMaxLevel    // por 0xA0
	FactExpValue    // por 0xF7: what killing it gives
	FactExpPerHP    // por 0xF9
	FactReaction    // por 0x2CF: charisma as a percent
	FactPaladin     // cotab cvPaladinFlag
	FactPartyPlace  // por 0x2B1
	FactExceptional // exceptional strength: por Str00, a sheet word only
)

func (AbilityFact) fact() {}
func (SkillFact) fact()   {}
func (SaveFact) fact()    {}
func (LevelFact) fact()   {}
func (CoinFact) fact()    {}
func (MemoFact) fact()    {}
func (KnownFact) fact()   {}
func (PlainFact) fact()   {}

// Value is what a Fact holds.
type Value interface{ value() }

type (
	Number int32 // BR's operands are longints
	Text   string
	Yes    bool
	Memo   struct { // a memory slot: the spell and whether it is still to learn
		Spell   SpellID
		Pending bool
	}
)

func (Number) value()    {}
func (Text) value()      {}
func (Yes) value()       {}
func (Memo) value()      {}
func (Race) value()      {}
func (ClassSet) value()  {}
func (Sex) value()       {}
func (Alignment) value() {}
func (Status) value()    {}
func (Side) value()      {}

// Coin is a kind of money. Old scripts see per-member coins as a view of the
// party's one purse.
type Coin string

// Copper is money: the rules' one unit. Gold-and-copper is presentation's, and
// a script's CoinFact is a view that converts.
type Copper int64

// ScriptAddr is a variable's address in a script's own memory map.
type ScriptAddr uint16

// RawValue is one script number and what it means: 0x87 written to Status is
// StatusStoned. A number with no row is an error at the moment it is used.
type RawValue struct {
	Raw   int32
	Value Value
}

// CharBinding says that a script's address reads and writes a Fact of the
// selected character. Without Values the number is the value.
//
// Bits instead splits one address over several facts: BR 268 is FactSide at
// bit 0, KnownFact at bit 6, FactSideLock at bit 7. The script machine has no
// character memory: direct or computed (GetTable, SaveTable), an address in the
// window with no binding is a run-time error, never a byte.
type CharBinding struct {
	Addr   ScriptAddr
	Fact   Fact
	Values []RawValue
	Bits   []BitBinding
}

// BitBinding is one fact in a mask of a packed address.
type BitBinding struct {
	Mask int32
	Fact Fact
}

// WorldWord is a script variable the rules depend on. The engine keeps script
// memory; it passes these on in FightTerms.
type WorldWord int

const (
	WordMorale     WorldWord = iota + 1 // por game.Morale
	WordSurprise                        // por game.Surprise
	WordMonsterHit                      // por game.MonsterHit
	WordPartyHit                        // por game.PartyHit
	WordPartyMove                       // por game.PartyMove
	WordNoMagic                         // por area.NoMagic
	WordNoLoot                          // por game.NoLoot
	WordNoExpAward                      // sotsb noXpAward; pod offset 30: this Combat() only hands out loot
	WordTrainable                       // por game.PartyExp: classes a hall teaches
)

// WorldBinding names the address of a WorldWord in this adventure's scripts.
type WorldBinding struct {
	Addr ScriptAddr
	Word WorldWord
}
