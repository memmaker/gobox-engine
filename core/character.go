package core

// Sheet is the rule facts of a character for the engine's screens: the part
// of game.Member a ruleset computes. Portrait and icon picks are the
// engine's and stay out.
type Sheet struct {
	Name      string
	Race      Race
	Classes   ClassSet
	Levels    map[Class]int32
	Alignment Alignment
	Sex       Sex
	Status    Status
	Age       int32
	Scores    map[Ability]int32
	Exp       int64
	NextExp   int64
	HP, MaxHP int32
	AC, THAC0 int32
	Move      int32
	Saves     map[Save]int32
	Skills    map[Skill]int32
	Items     []Carried
	Spells    []KnownSpell
	Slots     map[string]SlotCount // by level group, as the ruleset words it ("cleric 1")
	Effects   []string             // the spell effects on it, as the ruleset words them
	Figures   []Figure             // what a ruleset adds to the sheet (CombatRules.Sheet)
	CanCast   bool
	ItemCap   int32
}

// SlotCount is spell slots left and in all.
type SlotCount struct{ Left, Max int32 }

// Figure is a line a ruleset adds to the sheet (game.Stat).
type Figure struct {
	Name, Value string
	Help        []string
}

// KnownSpell is a spell known or memorized (game.SpellRef).
type KnownSpell struct {
	Spell     SpellID
	Name      string
	Level     int32
	Group     string
	Memorized bool
	Planned   int32
	Range     int32
	Targets   int32
	Area      int32
}

// Carried is an item on a character.
type Carried struct {
	Slot    ItemSlot
	Item    Item
	Readied bool
	Wrong   bool // the holder may not use it
}

// ItemKind is the kind of thing an item is.
type ItemKind string

const (
	KindWeapon ItemKind = "weapon"
	KindAmmo   ItemKind = "ammo"
	KindArmor  ItemKind = "armor"
	KindShield ItemKind = "shield"
	KindMisc   ItemKind = "misc" // rings, cloaks, potions: where it goes is Wear
)

// Dice is N d Sides + Bonus.
type Dice struct{ N, Sides, Bonus int32 }

// Item is one item as an adventure defines it, neutral. Today's records (por
// Item 0x3F bytes plus base table DS:5582, BR ITEMS) are decoded into this.
type Item struct {
	ID         ItemID
	NameWords  []string // the words known and unknown make the name (por Words, Hidden)
	Hidden     []bool
	Kind       ItemKind
	Wear       string // the slot it readies in, game.Item.Slot's names
	Hands      int32
	Range      int32
	Damage     Dice
	DamageBig  Dice
	Plus       int32
	Plus2      int32
	AC         int32
	Usable     []Class
	Charges    int32
	Count      int32
	Price      int32
	Cursed     bool
	Identified bool
	Casts      SpellID // a scroll, wand or potion's power, "" none
	Trait      Trait   // what it does while readied (por item affects), "" none
}

// Trait is a lasting thing a monster or item has that rules act on: racial
// attacks, gazes, poison. A ruleset says which it handles in its Catalog, and
// the one it does not fails at load.
type Trait string

// Attack row of a creature.
type CreatureAttack struct {
	Count int32
	Dice  Dice
}

// Creature is a monster or NPC as an adventure defines it (por charRecord +
// SPC + ITM blocks, eclbr.Record). A neutral sheet, in the numbers the module
// was written for; a ruleset converts them as it wants (por rules4e Role, Rank).
type Creature struct {
	Name      string
	Race      Race
	Classes   ClassSet
	Levels    map[Class]int32
	Alignment Alignment
	Sex       Sex
	Scores    map[Ability]int32
	Kind      int32 // por Kind: monster kind, bonuses against it
	HP, AC    int32
	THAC0     int32
	Move      int32
	Size      int32
	Attacks   []CreatureAttack
	Saves     map[Save]int32
	Skills    map[Skill]int32
	Spells    []SpellID
	Traits    []Trait
	Items     []Item
	Money     Purse
	ExpValue  int32
	ExpPerHP  int32
	Control   int32 // por Control: 0xB3 charmed, > 0x7F hired
	Icon      AssetRef
}

// Draft is a character being made: the sheet so far and what is still open.
// Allowed(Creating{...}) lists the choices.
type Draft struct {
	Sheet        Sheet
	Step         string            // por creation's step; "" any, as play.Draft
	Lows         map[Ability]int32 // the class's minimums (play.Draft.Minimums)
	MinHP, MaxHP int32
}

// Check is a question a script asks of the rules. Closed.
type Check interface{ check() }

type (
	// PartyStrength: por opPartyStrength.
	PartyStrength struct{ Party Party }
	// Census: por opCheckParty's low, high, average over a fact.
	Census struct {
		Party Party
		Fact  Fact
	}
	// HasCondition: por opCheckParty what=0, cotab and pod FindAffect, pod gameMagicCheck.
	HasCondition struct {
		Whom      []CharID
		Condition Condition
	}
	// HasClass: por opPartySurprise's ranger.
	HasClass struct {
		Party Party
		Class Class
	}
	// HasItem: por opFindItem, BR OpHasItem; the item by what a script numbers.
	HasItem struct {
		Whom []CharID
		Name ItemTerm
	}
	// SkillRoll: BR opSkill (skill, all, hardness).
	SkillRoll struct {
		Whom       []CharID
		Skill      Skill
		Difficulty int32
	}
)

// ItemTerm is what a script looks items up by: the item type in the original's
// numbering, named by the adventure. Open question in ARCHITECTURE.md.
type ItemTerm string

func (PartyStrength) check() {}
func (Census) check()        {}
func (HasCondition) check()  {}
func (HasClass) check()      {}
func (HasItem) check()       {}
func (SkillRoll) check()     {}

// Answer is a Check's result. Which fields mean anything follows from the
// Check: Low, High, Average for Census; Value for PartyStrength and SkillRoll;
// Yes for the Has ones; Who for who succeeded or was found.
type Answer struct {
	Yes            bool
	Value          int32
	Low, High, Avg int32
	Who            CharID
}

// CharacterFormat names a file format a ruleset reads.
type CharacterFormat string

// CharacterFile is one character outside the game: a .GUY, a CHRDATA, por's
// .cha with trailers (charimport.go, CHRDATAn.SAV).
type CharacterFile struct {
	Format CharacterFormat
	Bytes  []byte
}
