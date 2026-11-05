package classic

import (
	"gobox/front"
	"gobox/game"
)

// Flags is v with the plain status flags s.Flags holds. v where s has none.
func Flags(v front.View, s game.State) front.View {
	f := s.Flags
	if f == nil {
		return v
	}
	v.Exploring, v.Sweeping, v.Interact, v.CanCast = f.Exploring, f.Sweeping, f.Interact, f.CanCast
	v.Adventure, v.Day, v.Inside, v.Ordering = f.Adventure, f.Day, f.Inside, f.Ordering
	v.Exploring = v.Exploring && !overMenu(s)
	v.Credits, v.CreditPage, v.Full, v.IconDrop = f.Credits, f.CreditPage, f.Full, f.IconDrop
	return v
}
