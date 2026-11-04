// poolpic.go: portraits the game names by DAX ref (or pool name) are drawn
// from the shared pool (gobox/sprites), pixel-exact copies of the decoded DAX.
// Combat icons stay on DAX: games recolour them by palette index.

package present

import (
	"image"
	"strconv"
	"strings"
	"sync"

	"gobox/front"
	"gobox/sprites"
)

var (
	gameID      string // game.Data.ID, the origins.json key
	originsOnce sync.Once
	origins     map[string]map[string]map[string]string
)

// poolImage is the pool portrait pc names (Name, else File, Block, Frame via origins.json), nil if none.
func poolImage(pc front.Picture) image.Image {
	name := pc.Name
	if name == "" && pc.File != "" {
		originsOnce.Do(func() { origins = sprites.Origins() })
		f := origins[gameID][pc.File]
		if name = f[strconv.Itoa(pc.Block)+"."+strconv.Itoa(pc.Frame)]; name == "" {
			name = f[strconv.Itoa(pc.Block)]
		}
	}
	if !strings.HasPrefix(name, "head_") && !strings.HasPrefix(name, "body_") {
		return nil
	}
	return sprites.Image(series, "portraits", name)
}
