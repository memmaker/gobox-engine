package ui

import (
	"image"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gobox/asset"
)

// poolImage is pool picture name of kind, its Amiga twin while the Amiga art
// shows and it has one.
func poolPicture(kind, name string) image.Image {
	if amigaShown() { // the members' pool ids stay the DOS names: only the drawing swaps them for their Amiga twins
		if t := amigaTwin(name); t != "" {
			if im := asset.SpriteImage(Series, kind, t); im != nil {
				return im
			}
		}
	}
	return asset.SpriteImage(Series, kind, name)
}

var (
	twinsOnce sync.Once
	twins     map[string]string
)

// amigaTwin is the amiga_ pool name drawn where a game's DOS file and block
// of pool name dos show on the Amiga (any game's: the pool is shared) ("": none): the same block in the set's
// own Amiga archive (HEAD5), else in the one archive of its kind (HEAD), else
// for an action-pose CHEAD (block 128 on) the one head the Amiga draws for
// both poses.
func amigaTwin(dos string) string {
	twinsOnce.Do(func() {
		twins = map[string]string{}
		o := iconOrigins()
		var games []string // every game with an Amiga set: the pool is shared, a pick may be another game's picture
		for g := range o {
			if _, ok := o[asset.PlatformID(g, asset.PlatformAmiga)]; ok {
				games = append(games, g)
			}
		}
		sort.Strings(games) // the same twin every run when several games share a DOS picture
		for _, g := range games {
			d, a := o[g], o[asset.PlatformID(g, asset.PlatformAmiga)]
			var files []string
			for f := range d {
				files = append(files, f)
			}
			sort.Strings(files) // the same twin every run when several blocks share a DOS picture
			for _, f := range files {
				base := strings.TrimRight(strings.TrimSuffix(f, ".DAX"), "0123456789") + ".DAX"
				refs := make([]string, 0, len(d[f]))
				for r := range d[f] {
					refs = append(refs, r)
				}
				sort.Strings(refs)
				for _, r := range refs {
					n := d[f][r]
					if _, done := twins[n]; done {
						continue
					}
					t := a[f][r]
					if t == "" {
						t = a[base][r]
					}
					if b, err := strconv.Atoi(r); t == "" && err == nil && base == "CHEAD.DAX" && b >= 128 {
						t = a[base][strconv.Itoa(b-128)]
					}
					if t != "" {
						twins[n] = t
					}
				}
			}
		}
	})
	return twins[dos]
}
