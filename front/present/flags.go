package present

import (
	"gobox/front"
	"gobox/game"
)

// Flags is v with the plain status flags f; over: a menu is up over the 3D
// view. v where f is nil.
func Flags(v front.View, f *game.Flags, over bool) front.View {
	if f == nil {
		return v
	}
	v.Exploring, v.Sweeping, v.Interact, v.CanCast = f.Exploring, f.Sweeping, f.Interact, f.CanCast
	v.Adventure, v.Day, v.Inside, v.Ordering = f.Adventure, f.Day, f.Inside, f.Ordering
	v.Exploring = v.Exploring && !over
	v.Credits, v.CreditPage, v.CreditsPic, v.Full, v.IconDrop = f.Credits, f.CreditPage, f.CreditsPic, f.Full, f.IconDrop
	return v
}
