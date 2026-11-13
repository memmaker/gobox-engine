package present

import (
	"image"

	"gobox/front"
	"gobox/game"
	"gobox/game/play"
)

// derived is what the screens read of the game's State that is made from it,
// built once per Version (not every frame; the game is turn-based), with
// what camp keeps of the last exploring (kept, explored) carried over.
type derived struct {
	walls    *game.Walls       // the 3D view up (nil: none)
	kept     *game.Walls       // the one last explored: camp and a scene keep it
	automap  game.Automap      // the map of where the party is
	explored game.Automap      // the one last explored (camp)
	minimap  *front.MapSquares // automap's squares when it is the 16x16 map (nil: not)
	unknown  []game.ItemRef    // the party's items the shop's Id would tell something about
	border   windowBorder      // the window border's decoded pieces
}

// derived is the State's derived reads, made again when its Version changed.
func (r *frontState) derived() *derived {
	return r.der.update(r.g.Version(), func(old *derived) *derived {
		ps := r.g.State()
		d := &derived{}
		if old != nil {
			d.explored = old.explored
		}
		var oldKept *game.Walls
		if old != nil {
			oldKept = old.kept
		}
		d.walls, d.kept = wallsOf(ps, oldKept)
		d.unknown = unknownOf(ps)
		d.border = borderOf(r.Art())
		d.automap = automapOf(ps, d.explored)
		if _, ok := ps.(play.Exploring); ok {
			d.explored = d.automap
		}
		if am := d.automap; am.Size == image.Pt(16, 16) {
			d.minimap = Map(&front.MapSquares{}, am, play.CommonOf(ps).Where)
			if old != nil && old.minimap != nil && *d.minimap == *old.minimap { // the same map: the one uploaded
				d.minimap = old.minimap
			}
		}
		return d
	})
}
