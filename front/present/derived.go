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
	version  uint64
	ok       bool
	walls    *game.Walls       // the 3D view up (nil: none)
	kept     *game.Walls       // the one last explored: camp and a scene keep it
	automap  game.Automap      // the map of where the party is
	explored game.Automap      // the one last explored (camp)
	minimap  *front.MapSquares // automap's squares when it is the 16x16 map (nil: not)
	border   windowBorder      // the window border's decoded pieces
}

// derived is the State's derived reads, made again when its Version changed.
func (r *rules) derived() *derived {
	if v := r.g.Version(); !r.der.ok || r.der.version != v {
		ps, old := r.g.State(), r.der
		d := derived{version: v, ok: true, explored: old.explored}
		d.walls, d.kept = wallsOf(ps, old.kept)
		d.border = borderOf(r.Art())
		d.automap = automapOf(ps, old.explored)
		if _, ok := ps.(play.Exploring); ok {
			d.explored = d.automap
		}
		if am := d.automap; am.Size == image.Pt(16, 16) {
			d.minimap = Map(&front.MapSquares{}, am, play.CommonOf(ps).Where)
			if old.minimap != nil && *d.minimap == *old.minimap { // the same map: the one uploaded
				d.minimap = old.minimap
			}
		}
		r.der = d
	}
	return &r.der
}
