package core

// OpenTurn: the Fight's Actor's turn menu is shown (por combatMenu's
// opening, before its offer): Resolve runs what the menu runs first, then
// Allowed(Fight) lists the offer without changing anything.
type OpenTurn struct{ Actor CharID }

func (OpenTurn) action() {}
