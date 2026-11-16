package rules

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

// Doll slots: where a readied item shows on the paper doll. 0..11 are the
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
	DollArrows   // the Missiles slot: arrows, quarrels or (extra) a thrown weapon
	DollLauncher // extra: the Launcher slot, a bow, crossbow or sling
	DollSlots
)

// DollNames are the doll slots' names by slot (game.Item.Slot); the arrows
// (Missiles) slot is shared by arrows, bolts and missiles (a thrown weapon).
var DollNames = [DollSlots]string{DollWeapon: "weapon", DollShield: "shield", DollArmor: "armor", DollGauntlets: "gauntlets", DollHelm: "helm",
	DollBelt: "belt", DollRobe: "robe", DollCloak: "cloak", DollBoots: "boots", DollRing1: "ring", DollRing2: "ring2", DollArrows: "arrows", DollLauncher: "launcher"}

// UsingItem: Who would use Item out of the inventory (por's UseItem before
// its questions). Allowed lists the uses: the item's own spell (one UseItem
// with no Spell), or each spell a scroll lets Who read (UseItem with Spell,
// Name the spell's); one Offer with Not set (its refusal, as the rules word
// it) when the item must be readied first; none when it cannot be used now.
type UsingItem struct {
	Who     CharID
	Item    ItemHandle
	NoMagic bool // the area forbids magic: no item is used
}

func (UsingItem) situation() {}

// Handling: Who looks over Item in the gear menu (viewchar.pas UseItemMenu). Allowed lists what the gear menu offers
// beside Ready and Drop: a GiveItem to another member (To NoChar: the engine asks whom), a Halve, a JoinStack, each
// with its name; the ones the ruleset does not offer here are left out.
type Handling struct {
	Who  CharID
	Item ItemHandle
}

func (Handling) situation() {}

// Reloading: Who's Item in the shop's ammunition counter. Allowed lists
// what the shop sells for it (Reload, a Clip and a Fill with the price),
// none when no ammunition is sold for it here.
type Reloading struct {
	Who  CharID
	Item ItemHandle
}

func (Reloading) situation() {}

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

// ItemTableMissing: the base item table holds no record (the game data lacks it; Yes).
type ItemTableMissing struct{}

func (ItemTableMissing) check() {}

// StopUsingItem ends the use of an item out of the inventory (por UsingItem, DS:6DBF, cleared).
type StopUsingItem struct{}

func (StopUsingItem) action() {}

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
	Kind                ItemID // picks the icon: its base kind, the game's item type through its kind table (0: none), or its own id as one of the game's named items once identified
	Weight              int32
	Readied, Identified bool
	Usable              bool // it has a Use
	CannotUse           bool // the class may not ready it
	Tab, Slot           string
	Words               []string
	Hidden              string   // its real name while unidentified
	Look, LookPlain     []string // against the item in its slot; alone
	Sale                Purse    // what a shop pays for it, in the ruleset's own coins
	Scribing            bool     // someone was to scribe from the scroll
	Scribe              bool     // Who may scribe its first spell
	Scroll              []ScrollSpell
	Casts               int32 // the spell id its own use casts (por item +0x3D, decoded at import)
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
// Powers is the engine's to add. NoMagic: the area forbids magic (no items are used).
type InventoryFact struct{ NoMagic bool }

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

// DestroyItems takes every Item Whom carry (the party inventory too) out of
// the game, each readied one put away first; MagicalOnly spares those with
// no plus (the script's DestroyItems; ruling S9, for all games).
type DestroyItems struct {
	Whom        Party
	Item        ItemID
	MagicalOnly bool
}

func (DestroyItems) action() {}

// TakeItems is BR's TakeItem (eclvm.pas EclDestroyItems): the items of kind Item that Whom carry go,
// a readied one with them. OneEach: only the first of each member's, and a stack of joinable items
// loses one instead.
type TakeItems struct {
	Whom    Party
	Item    ItemID
	OneEach bool
}

// StashItems is BR's Stash (eclvm.pas EclTransferCharacterLists): every Whom's carried items go onto
// its second list in front of what is there (Away), or that list comes back onto the first.
type StashItems struct {
	Whom Party
	Away bool
}

func (TakeItems) action()  {}
func (StashItems) action() {}

// TreasureRecordsFact is the treasure's items as the game's 0x3F-byte file records
// (ItemRecords), last of the list first, so that a block read in file order
// by AddTreasureBlock comes back in it: what the vault writes back.
type TreasureRecordsFact struct{}

// ItemRecords is item records, one after the other.
type ItemRecords []byte

// DepositItem: Who leaves Item in the treasure (the vault's Deposit; the
// items menu's Drop there). It goes at the head, so it is last in the file.
// ErrNotNow when Who does not hold it; an item that may not be lost (readied) stays.
type DepositItem struct {
	Who  CharID
	Item ItemHandle
}

func (TreasureRecordsFact) fact() {}

func (ItemRecords) value() {}

func (DepositItem) action() {}
