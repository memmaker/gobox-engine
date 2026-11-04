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
	v.Exploring, v.Sweeping, v.Interact, v.CanCast, v.Open = f.Exploring, f.Sweeping, f.Interact, f.CanCast, f.Open
	v.Adventure, v.Day, v.Inside, v.Ordering = f.Adventure, f.Day, f.Inside, f.Ordering
	v.Credits, v.Full, v.IconDrop = f.Credits, f.Full, f.IconDrop
	return v
}
