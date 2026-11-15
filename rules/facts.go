package rules

// Fact is one thing about a character that a script can read or write. The
// set is closed and small because the scripts' use is: see the per-game table
// in ARCHITECTURE.md. Facts compare with ==, so Catalog can list them.
type Fact interface{ fact() }

type (
	AbilityFact     struct{ Ability Ability } // por 0x18 con; cotab cvIntelligence
	SkillFact       struct{ Skill Skill }     // por 0xA5..0xAC; BR 76+s
	SaveFact        struct{ Save Save }       // por 0x9B
	LevelFact       struct{ Class Class }     // the level the class counts at: its own plus the class left behind while that one is active (cotab cvMageLevel, sotsb $C9..$D0)
	OwnLevelFact    struct{ Class Class }     // the level in the class itself, a class left behind not counted (pod classLevels[])
	FormerLevelFact struct{ Class Class }     // the level the class had before a dual-class change (pod oldLevels[])
	OwnScoreFact    struct{ Ability Ability } // the score before items and spells (pod stats[].base); AbilityFact is the one in use
	CoinFact        struct{ Coin Coin }       // the party's one purse, whoever is selected. por 0xBB..0xC3, also computed (area 9)
	KnownFact       struct{ Skill Skill }     // has the skill at all: BR 76 reads 128 or 0
	MemoFact        struct{ Slot int }        // por 0x1F..0x70: Knock in area 14
	PlainFact       int                       // the rest, one each
	FlagFact        struct{ Slot int }        // BR 161+Slot: a script's flag byte kept on the character (rulesxxvc ScriptFlags)
	HitFact         struct{ Target CharID }   // a's chance in percent to hit Target with one blow of the readied attack (por hover text)
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
	FactHP          // BR 281; pod hpCur (a write stores the byte)
	FactMaxHP       // BR 69; sotsb cvMaxHP; pod hpMax
	FactMaxLevel    // por 0xA0
	FactExpValue    // por 0xF7: what killing it gives
	FactExpPerHP    // por 0xF9
	FactReaction    // por 0x2CF: charisma as a percent
	FactPaladin     // cotab cvPaladinFlag
	FactPartyPlace  // por 0x2B1
	FactExceptional // exceptional strength: por Str00, a sheet word only
	FactHarmed      // a sheet word only: the harmful effects on it, ", " between (por's after-fight list)
	FactSideOnly    // por opAddNPC: 268 bit 0 written alone (SideParty or SideMonsters), the auto flag kept
	FactLevelReady  // por (extra): it may train a level (Yes)
	FactCompanion   // BR 76 (isCompanion): a party member the player does not command (Yes)
)

func (AbilityFact) fact()     {}
func (SkillFact) fact()       {}
func (SaveFact) fact()        {}
func (LevelFact) fact()       {}
func (OwnLevelFact) fact()    {}
func (FormerLevelFact) fact() {}
func (OwnScoreFact) fact()    {}
func (CoinFact) fact()        {}
func (PiecesFact) fact()      {}
func (MemoFact) fact()        {}
func (KnownFact) fact()       {}
func (PlainFact) fact()       {}
func (FlagFact) fact()        {}
func (HitFact) fact()         {}

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

// PiecesFact is the count of unappraised gems (Jewel false) or jewelry the party
// holds: items, not money, until appraised and sold. A write adds or removes them.
type PiecesFact struct{ Jewel bool }

// Coin is a kind of money: gold and copper in the Gold Box games (the import
// rewrites the scripts' other coins, script.GoldCoins), credits in Buck Rogers.
type Coin int8

// The coins of every ruleset: one closed set, so no two rulesets' coins collide
// (dnd and xxvc name them).
const (
	CoinCopper Coin = iota + 1
	CoinGold
	CoinCredits
)

// Copper is money: the rules' one unit. Gold-and-copper is presentation's, and
// a script's CoinFact is a view that converts.
type Copper int64
