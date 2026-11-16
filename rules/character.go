package rules

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
	NextExp   int64 // 0: it rises no further
	LevelExp  int64 // what its current level took
	HP, MaxHP int32
	AC, THAC0 int32 // AC: the front armour class
	RearAC    int32 // the rear armour class (attacks from behind)
	Move      int32
	Saves     map[Save]int32
	Skills    map[Skill]int32
	Items     []Carried
	Known     []SpellID            // can be memorized
	Memorized []MemorizedSpell     // what is in the head, with castings left
	Burden    Burden               // zero under Options.PartyInventory
	Slots     map[string]SlotCount // by level group, as the ruleset words it ("cleric 1")
	Effects   []string             // the spell effects on it, as the ruleset words them
	DualClass string               // a dual-classed character's class left and how it stands, as the ruleset words it ("": none)
	Health    Health               // how fit it is, behind its status word
	Figures   []Figure             // what a ruleset adds to the sheet (CombatRules.Sheet)
	CanCast   bool
	ItemCap   int32
}

// SlotCount is spell slots left and in all.
type SlotCount struct{ Left, Max int32 }

// Figure is a line a ruleset adds to the sheet, with its tooltip lines (game.Stat).
type Figure struct {
	Name, Value string
	Help        []string
}

// MemorizedSpell keeps memorized and remaining apart; slots come from class,
// level and cleric Wisdom, never from a storage size.
type MemorizedSpell struct {
	Spell   SpellID
	Left    int32
	Planned int32
}

// Burden is encumbrance: the one concept of the games' default rules.
type Burden struct{ Weight, Limit int32 }

// School groups a spell.
type School string

// SpellEffect is what the spell does, in the ruleset's terms; a behaviour
// variant between games is its own Effect, explicit.
type SpellEffect string

// SpellSource is where a casting comes from.
type SpellSource string

const (
	FromMemory SpellSource = "memorized"
	FromScroll SpellSource = "scroll"
	FromWand   SpellSource = "wand"
	FromItem   SpellSource = "item"
)

// Spell is a registry row, keyed by a stable id. What the original tables
// numbered by game ends here.
type Spell struct {
	ID      SpellID
	Name    string
	School  School
	Level   int32
	Classes []Class
	Effect  SpellEffect
	Range   int32
	Targets int32
	Area    int32
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

// Item is one item as a ruleset states it. Today's records (por Item 0x3F bytes
// plus base table DS:5582, BR ITEMS) become ruleset tables at import.
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

// ItemRef is an item an adventure places: the ruleset supplies everything else
// about ID (Rules.Item). Plus and Charges are the placed one's own.
type ItemRef struct {
	ID      ItemID
	Plus    int32
	Charges int32
	Count   int32
	Cursed  bool
}

// Draft is a character being made: the sheet so far and what is still open.
// Allowed(Creating{...}) lists the choices.
type Draft struct {
	Sheet        Sheet
	Step         string            // por creation's step; "" any, as play.Draft
	Lows         map[Ability]int32 // the class's minimums (play.Draft.Minimums)
	MinHP, MaxHP int32
	// Picks is the option picked per FieldChoice field; Points the points
	// spent per option of a FieldSkills field and Left the points that field
	// still has to spend (CreationRules).
	Picks      map[FieldID]int
	Points     map[FieldID][]int32
	Left       map[FieldID]int32
	Name       string
	Head, Body PicID // portrait pool pictures
	Portrait   PNG   // a dropped PNG portrait, encoded (nil: the pool head and body)
	Icon       Icon
	// Of is the character being modified (NoChar: a new one); Make updates it.
	Of CharID
	// The roll's values the sheet does not show, so Make keeps the rolled
	// character: Purse the starting money, HitRoll the hit-dice roll before
	// adjustments, Ident the ruleset's identity roll (POR RandomID).
	Purse   Copper
	HitRoll int32
	Ident   uint32
	// HitRolls are the starting training's hit-dice results in the order the
	// ruleset rolled them (per level, per class: cotab, sotsb, pod train to
	// the starting experience), so Make replays the training without dice.
	HitRolls []int32
	// Buy is the bought bases RollPointBuy takes (gobox/creation's PointBuy).
	Buy map[Ability]int32
}

// Pics is d's pictures by their field names (play.SetChoice's fields).
func (d *Draft) Pics() map[string]*PicID {
	return map[string]*PicID{"portraithead": &d.Head, "portraitbody": &d.Body, "iconhead": &d.Icon.Head, "iconbody": &d.Icon.Body}
}

// TrapDefinition is a trap as data: no universal skip of the stoned, the damage
// kind and target eligibility decide. Replaces TrapSkipStoned, TrapRollEach.
type (
	TrapTarget string
	TrapDamage string
	TrapSave   string
)

const (
	TrapOne      TrapTarget = "one"
	TrapAll      TrapTarget = "all"
	TrapOnePer   TrapDamage = "roll-once"
	TrapRollEach TrapDamage = "roll-each"
	SaveNegates  TrapSave   = "negates"
	SaveHalves   TrapSave   = "halves"
	SaveNone     TrapSave   = ""
)

type TrapDefinition struct {
	TargetMode TrapTarget
	DamageMode TrapDamage
	Damage     Dice
	Kind       string // what kind of damage: decides who it can touch
	SaveType   Save
	SaveEffect TrapSave
}

// AttackRequest and AttackResult are one attack roll, for every ruleset: this
// replaces por's Hits hook (rules.go:230). Natural is the die, kept for 1, 20,
// and CriticalHits.
type AttackRequest struct {
	From, To CharID
	Bonus    int32
	Need     int32
}

type AttackResult struct {
	Natural  int32
	Hit      bool
	Critical bool
}

// Check is a question a script asks of the rules. Closed.
type Check interface{ check() }

type (
	// PartyStrength: por, cotab EclPartyStrength, a number scripts divide into
	// monster counts or compare. Scale fixed for the dnd world (ARCHITECTURE.md).
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
	// HasAffect: a script's FindAffect (cotab, sotsb, pod) and CheckParty
	// what=0 (por): the affect by the number the module's scripts write.
	HasAffect struct {
		Whom   []CharID
		Affect ScriptAffect
	}
	// HasClass: por opPartySurprise's ranger.
	HasClass struct {
		Party Party
		Class Class
	}
	// HasItem: por opFindItem, BR OpHasItem; the item by what a script numbers.
	HasItem struct {
		Whom []CharID
		Item ItemID
	}
	// HasMemorized: an okay member with Spell memorized (por lockedDoor's
	// Knock, FindCaster); Who is the first in marching order.
	HasMemorized struct {
		Party Party
		Spell SpellID
	}
	// SkillRoll: BR opSkill (skill, all, hardness), eclvm.pas EclSkillCheck. The skill is named, or
	// (Skill empty) the script's number; All: Whom is the whole party (the report says "party").
	SkillRoll struct {
		Whom     []CharID
		Skill    Skill
		Script   ScriptSkill
		All      bool
		Hardness Hardness
	}
)

func (PartyStrength) check() {}
func (Census) check()        {}
func (HasCondition) check()  {}
func (HasClass) check()      {}
func (HasAffect) check()     {}
func (HasItem) check()       {}
func (HasMemorized) check()  {}
func (SkillRoll) check()     {}

// Answer is a Check's result. Which fields mean anything follows from the
// Check: Low, High, Average for Census; Value for PartyStrength; Grade for
// SkillRoll (Text: the report the original shows, "<name> succeeds at <skill>"); Yes for the Has
// ones; Who for who succeeded or was found.
type Answer struct {
	Text           string
	Yes            bool
	Value          int32
	Grade          Grade
	Low, High, Avg int32
	Who            CharID
}

// CharacterFile is one character in the engine's canonical format, between
// adventures on the same ruleset. Originals (.GUY, CHRDATA) come in once through
// the save importer (ARCHITECTURE.md); nothing writes them.
type CharacterFile []byte

// ScriptAffect is an affect as a module's scripts number it; the ruleset
// reads it through the game's own affect numbering.
type ScriptAffect byte

// Health is how fit a character is: what its status word stands for.
type Health int

const (
	Hale     Health = iota // nothing ails it
	Ailing                 // something ails it, and it can act
	Disabled               // it cannot act at all (down, held, helpless ...)
)
