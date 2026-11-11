package core

// OpenTurn: the Fight's Actor's turn menu is shown (por combatMenu's
// opening, before its offer): Resolve runs what the menu runs first, then
// Allowed(Fight) lists the offer without changing anything.
type OpenTurn struct{ Actor CharID }

func (OpenTurn) action() {}

// ApplyItems: Who's readied items are worked into its figures again (AC,
// THAC0, damage, the item count; por ApplyItems, 00BA:0BB8).
type ApplyItems struct{ Who CharID }

func (ApplyItems) action() {}

// MemoSlots is how many MemoFact slots a character has (0..MemoSlots-1).
const MemoSlots = 21
