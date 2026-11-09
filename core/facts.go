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
	CoinFact    struct{ Coin Coin }       // por 0xBB..0xC3: a view of the one purse
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
	FactStatus     // por 0x100 reads Active, writes down it; BR 256
	FactSide       // 268
	FactControl    // por 0xB8 (charmed, hired)
	FactMove       // por 0x11B
	FactHP         // BR 281
	FactMaxHP      // BR 69; sotsb cvMaxHP
	FactMaxLevel   // por 0xA0
	FactExpValue   // por 0xF7: what killing it gives
	FactExpPerHP   // por 0xF9
	FactReaction   // por 0x2CF: charisma as a percent
	FactPaladin    // cotab cvPaladinFlag
	FactPartyPlace // por 0x2B1
)

func (AbilityFact) fact() {}
func (SkillFact) fact()   {}
func (SaveFact) fact()    {}
func (LevelFact) fact()   {}
func (CoinFact) fact()    {}
func (MemoFact) fact()    {}
func (PlainFact) fact()   {}

// Value is what a Fact holds.
type Value interface{ value() }

type (
	Number int32 // BR's operands are longints
	Text   string
	Memo   struct { // a memory slot: the spell and whether it is still to learn
		Spell   SpellID
		Pending bool
	}
)

func (Number) value()    {}
func (Text) value()      {}
func (Memo) value()      {}
func (Race) value()      {}
func (ClassSet) value()  {}
func (Sex) value()       {}
func (Alignment) value() {}
func (Status) value()    {}
func (Side) value()      {}

// Coin is a kind of money. Rulesets hold one purse (gobox/rules.Purse); old
// scripts see per-member coins as a view of it.
type Coin string

const (
	Copper   Coin = "copper"
	Silver   Coin = "silver"
	Electrum Coin = "electrum"
	Gold     Coin = "gold"
	Platinum Coin = "platinum"
	Gems     Coin = "gems"
	Jewelry  Coin = "jewelry"
	Credits  Coin = "credits" // BR
)

// Purse is the party's money by kind.
type Purse map[Coin]uint64

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
type CharBinding struct {
	Addr   ScriptAddr
	Fact   Fact
	Values []RawValue
}

// WorldWord is a script variable the rules depend on. The engine keeps script
// memory; it passes these on in FightTerms and ShopTerms.
type WorldWord int

const (
	WordMorale      WorldWord = iota + 1 // por game.Morale
	WordSurprise                         // por game.Surprise
	WordMonsterHit                       // por game.MonsterHit
	WordPartyHit                         // por game.PartyHit
	WordPartyMove                        // por game.PartyMove
	WordNoMagic                          // por area.NoMagic
	WordNoLoot                           // por game.NoLoot
	WordDifficulty                       // pod gameDifficulty
	WordPriceFactor                      // cotab gameWordPriceFactor
	WordHideExp                          // sotsb HideExperience
	WordTrainable                        // por game.PartyExp: classes a hall teaches
)

// WorldBinding names the address of a WorldWord in this adventure's scripts.
type WorldBinding struct {
	Addr ScriptAddr
	Word WorldWord
}
