// Package playpor starts Pool of Radiance's engine on an imported adventure
// folder (`gobox import por`) for cmd/gobox through the shared pool host
// (gobox/pool/playpool).
package playpor

import (
	"gobox/app"
	"gobox/ega"
	"gobox/game"
	"gobox/pool"
	"gobox/pool/playpool"
	"por/presentdata"
	"por/presentdata/blobberart"
	"por/rules4e"
)

// Play runs the game of s.Adventure until the window closes.
func Play(s app.Start) error {
	return playpool.Play(s, playpool.Game{
		Title: "Pool of Radiance", W: ega.W, H: ega.H, Record: pool.RecordLayout,
		Combat4e: rules4e.New(), // the Gameplay page's "4e Combat"; swapped in and out mid-fight
		Data:     Data,
	})
}

// Data is the engine's game.Data with POR's art handed in: the remastered
// area maps and the blobber's pictures.
func Data(eng *pool.Engine) game.Data {
	d := eng.Data()
	d.AreaMaps, d.Remaster = presentdata.AreaMaps, blobberart.FS
	return d
}
