package core

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
