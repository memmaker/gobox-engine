// poolpic.go: portraits the game names by DAX ref (or pool name) are drawn
// from the shared pool (gobox/sprites), pixel-exact copies of the decoded DAX.
// Combat icons stay on DAX: games recolour them by palette index.

package present

import (
	"gobox/importer/pictures"
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
		originsOnce.Do(func() { origins = pictures.Origins() })
		f := origins[gameID][pc.File]
		if amigaShown() { // the Amiga set only (<game>-amiga, one archive per kind): never the DOS art
			a, n := origins[gameID+"-amiga"], strings.TrimSuffix(strings.ToUpper(pc.File), ".DAX")
			if f = a[n+".DAX"]; f[strconv.Itoa(pc.Block)] == "" && f[strconv.Itoa(pc.Block)+"."+strconv.Itoa(pc.Frame)] == "" { // a set's own archive (HEAD5), else the shared one
				f = a[strings.TrimRight(n, "0123456789")+".DAX"]
			}
		}
		if name = f[strconv.Itoa(pc.Block)+"."+strconv.Itoa(pc.Frame)]; name == "" {
			name = f[strconv.Itoa(pc.Block)]
		}
	}
	if n := strings.TrimPrefix(name, "amiga_"); !strings.HasPrefix(n, "head_") && !strings.HasPrefix(n, "body_") {
		return nil
	}
	return sprites.Image(series, "portraits", name)
}
