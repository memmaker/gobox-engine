package core

// Items: the inventory as the engine shows it, item cards, identifying and valuing.

// Money is an amount the front end shows (as "12g 50c").
type Money struct{ Gold, Copper int }

// InvTab sorts the party inventory into tabs (extra).
type InvTab int

const (
	TabMisc InvTab = iota
	TabWeapon
	TabArmor
)

// Doll slots: where a readied item shows on the paper doll. 0..12 are the
// game's own slots (rules1e Character.Slots); bracers take the armour's (readySlot).
const (
	DollWeapon = iota
	DollShield
	DollArmor
	DollGauntlets
	DollHelm
	DollBelt
	DollRobe
	DollCloak
	DollBoots
	DollRing1
	DollRing2
	DollArrows // arrows or quarrels beside the bow or crossbow in hand
	DollSlots
)

// DollNames are the doll slots' names by slot (game.Item.Slot); the arrows
// slot is shared by arrows and bolts.
var DollNames = [DollSlots]string{DollWeapon: "weapon", DollShield: "shield", DollArmor: "armor", DollGauntlets: "gauntlets", DollHelm: "helm",
	DollBelt: "belt", DollRobe: "robe", DollCloak: "cloak", DollBoots: "boots", DollRing1: "ring", DollRing2: "ring2", DollArrows: "arrows"}

// UsingItem: Who would use Item out of the inventory (por's UseItem before
// its questions). Allowed lists the uses: the item's own spell (one UseItem
// with no Spell), or each spell a scroll lets Who read (UseItem with Spell,
// Name the spell's); one Offer with Not set (its refusal, as the rules word
// it) when the item must be readied first; none when it cannot be used now.
type UsingItem struct {
	Who  CharID
	Item ItemHandle
}

func (UsingItem) situation() {}

// Fit asks whether Who is fit to act (Yes; por's Okay, [6C00]).
type Fit struct{ Who CharID }

func (Fit) check() {}

// The name words of a valued gem and jewel (por's item word 2).
const (
	GemNameWord   = 0x65
	JewelNameWord = 0xD6
)

// Valuate takes a gem (or a jewel) from the purse, rolls its value and holds
// the valued item (ovr22:19B2); refused when the purse has none.
type Valuate struct{ Jewel bool }

// KeepValued keeps the valued item for Who when Keep and there is room,
// else sells it; nothing is held after.
type KeepValued struct {
	Who  CharID
	Keep bool
}

// ValuedItem asks the valued item held: Yes it is a jewel, Value its
// ItemHandle (NoItem none), High its value in gold pieces.
type ValuedItem struct{}

func (Valuate) action() {}

func (KeepValued) action() {}

func (ValuedItem) check() {}

// PoolCap: the party inventory's room (Value).
type PoolCap struct{}

// PoolCount: the items in the party inventory (Value).
type PoolCount struct{}

func (PoolCount) check() {}

// CarryCount: what Who can reach, the party inventory after its own (Value).
type CarryCount struct{ Who CharID }

func (CarryCount) check() {}

func (PoolCap) check() {}

// ListedFact is the items Who sees listed as the inventory orders them
// (Handles); Goods: the shop's wares, named for no one.
type ListedFact struct{ Goods bool }

// IdentifiableFact is the items the shop's Id offers Who (Handles).
type IdentifiableFact struct{}

// ItemCardFact is Item as Who sees it (ItemCard); Who may be NoChar. Who is
// the holder when the card shows a held item.
type ItemCardFact struct{ Item ItemHandle }

// DamageFact is the dice Who's blow rolls, as the sheet words them ("1d8+1": Text).
type DamageFact struct{}

func (ListedFact) fact() {}

func (IdentifiableFact) fact() {}

func (ItemCardFact) fact() {}

func (DamageFact) fact() {}

// Handles are items in an order.
type Handles []ItemHandle

// ItemCard is an item as a character sees it: the one card for the item
// list, the inventory screen and the weapon in hand.
type ItemCard struct {
	Item                ItemHandle // the item
	Rules               []string   // extra: how it attacks in combat (the bar's tooltip), set on the Turn's Weapon and Missile
	Name                string
	Type, Count, Price  int32
	Weight              int32
	Readied, Identified bool
	Usable              bool // it has a Use
	CannotUse           bool // the class may not ready it
	Tab, Slot           string
	Words               []string
	Hidden              string   // its real name while unidentified
	Look, LookPlain     []string // against the item in its slot; alone
	Sale                Copper
	Scribing            bool // someone was to scribe from the scroll
	Scribe              bool // Who may scribe its first spell
	Scroll              []ScrollSpell
	Casts               int32 // the spell number its own use casts (por item +0x3D)
}

func (Handles) value() {}

func (ItemCard) value() {}

// IdentifyItem pays the shop's price for Item's hidden words to show (the
// Outcome has an ItemIdentified; none: the party could not pay).
type IdentifyItem struct {
	Who  CharID
	Item ItemHandle
}

func (IdentifyItem) action() {}

// ItemIdentified is the shop's Id done: Revealed is false when the item had
// nothing new to tell.
type ItemIdentified struct {
	Item     ItemHandle
	Revealed bool
}

func (ItemIdentified) effect() {}

// InventoryFact is the member's inventory screen (Fact: an InventoryView);
// Powers is the engine's to add.
type InventoryFact struct{}

// InventoryView is the inventory screen's rule part: the readied items by
// doll slot (nil: empty), the party inventory in its list order, and what
// can be done here.
type InventoryView struct {
	Member int // the member's place in the party
	Cap    int // the party inventory's slots
	Doll   []*ItemCard
	Pool   []ItemCard
	Use    bool // items can be used here
	Sell   bool // a shop buys them
	Switch bool // another member can be shown (not in combat)
}

func (InventoryFact) fact() {}

func (InventoryView) value() {}

// ApplyItems: Who's readied items are worked into its figures again (AC,
// THAC0, damage, the item count; por ApplyItems, 00BA:0BB8).
type ApplyItems struct{ Who CharID }

func (ApplyItems) action() {}

// HasRoom asks whether Who has room for Item (Yes; por's Overloaded,
// ovr20:275B, negated) with Who's items as last applied (ApplyItems).
type HasRoom struct {
	Who  CharID
	Item ItemHandle
}

func (HasRoom) check() {}

// ItemByNumber is the ruleset's id of the original's item type N
// (ItemID; the Fact errs when the ruleset has none).
type ItemByNumber struct{ N int }

func (ItemByNumber) fact() {}
func (ItemID) value()      {}
