package presentdata

import (
	"gobox/engine"
	"gobox/game"
	"gobox/games/por/presentdata/blobberart"
)

// Data is the engine's game.Data with POR's art handed in: the remastered
// area maps and the blobber's pictures.
func Data(eng *engine.Engine) game.Data {
	d := eng.Data()
	d.AreaMaps, d.Remaster = AreaMaps, blobberart.FS
	return d
}
