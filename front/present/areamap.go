package present

import (
	"gobox/front"
	"gobox/gpu"
)

// drawAreaMap draws the overland's whole map over the window: whole times
// its pixels, at least the page's scale and as much more as the map needs
// to cover the window, centred on the party as far as the map's edges let
// it (no bars), the party's marker on its square.
func (p *Presenter) drawAreaMap(dst *gpu.Image, m *front.AreaMap) {
	p.area.Draw(dst, m.Img, m.Tile, m.Party, m.Marker, scale(dst))
}
