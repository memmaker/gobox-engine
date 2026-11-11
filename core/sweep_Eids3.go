package core

// The inventory as the engine shows it (fold Ids3: moved from rules1e).

// InvItem is one item of the inventory.
type InvItem struct {
	Index   int      // its place in the member's items (its number in play)
	Name    string   // as the list shows it
	Type    byte     // base item type
	Words   []string // the item's name words, known or not (for its icon)
	Count   int
	Readied bool     // it is readied (shown with present.Data.ReadyMark, for a game without a doll)
	Slot    bool     // it has a doll slot (it can be readied)
	Wrong   bool     // the member's class cannot ready it
	Use     bool     // it has a use (a scroll, a spell of its own)
	Look    []string // what it does (the shop's Look text)
	Price   Money    // what selling it pays (a ware: what it costs; loot: none)
	Scribe  bool     // a scroll someone was going to scribe from (the front end asks before it goes)
	Kind    InvTab   // its party inventory tab
	Rules   []string // extra: how it attacks in combat (the bar's tooltip), set on the Turn's Weapon and Missile
}

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
