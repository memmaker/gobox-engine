package ui

import (
	"image"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gobox/asset"
)

// The pool's entry IDs are art-neutral: an entry is one portrait or icon
// part with its DOS picture and/or its Amiga twin. Its ID is the DOS sprite
// name where a DOS picture exists (so every save stored before Amiga art
// keeps its ID), else the Amiga-only sprite's own name. A stored amiga_ name
// that has a DOS twin (a pick made under F12 before the IDs) resolves to that
// twin's ID (poolEntryID), so which art showed at the pick never matters.

// amigaPrefix starts the pool name of every Amiga-art sprite.
const amigaPrefix = "amiga_"

// poolPicture is the picture of pool entry id (or a legacy amiga_ name) in
// the art shown (F12).
func poolPicture(kind, id string) image.Image { return entryPicture(kind, id, amigaShown()) }

// entryPicture is entry id's picture under the Amiga art (amiga) or the DOS
// art, falling back to the other art's picture when the entry has none in
// the one asked (a DOS-only or Amiga-only entry shows under both arts).
func entryPicture(kind, id string, amiga bool) image.Image {
	id = poolEntryID(id)
	dos, am := id, amigaTwin(id)
	if strings.HasPrefix(id, amigaPrefix) {
		dos, am = "", id
	}
	first, other := dos, am
	if amiga {
		first, other = am, dos
	}
	for _, n := range []string{first, other} {
		if n == "" {
			continue
		}
		if im := asset.SpriteImage(Series, kind, n); im != nil {
			return im
		}
	}
	return nil
}

// poolEntryID is the entry ID of a stored pool name: an amiga_ name with a
// DOS twin is that twin's ID (the alphabetically first when several DOS
// pictures share one Amiga twin), anything else is its own ID.
func poolEntryID(name string) string {
	amigaTwin("") // builds twins and dosOf
	if d := dosOf[name]; d != "" {
		return d
	}
	return name
}

// poolEntries lists the entry IDs of kind whose DOS sprite names start with
// prefix, sorted: every DOS picture's, then every Amiga picture without a
// DOS twin. The list depends on the sprites present, not their order.
func poolEntries(kind, prefix string) []string {
	var ids []string
	for _, n := range asset.SpriteNames(Series, kind) {
		if strings.HasPrefix(n, prefix) || strings.HasPrefix(n, amigaPrefix+prefix) {
			ids = append(ids, poolEntryID(n))
		}
	}
	slices.Sort(ids)
	return slices.Compact(ids)
}

var (
	twinsOnce sync.Once
	twins     map[string]string // DOS name -> its Amiga twin
	dosOf     map[string]string // Amiga twin -> the first DOS name it twins
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
		dosOf = map[string]string{}
		for d, a := range twins {
			if o, ok := dosOf[a]; !ok || d < o {
				dosOf[a] = d
			}
		}
	})
	return twins[dos]
}
