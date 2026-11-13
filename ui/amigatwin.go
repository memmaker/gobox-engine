package ui

import (
	"image"
	"sort"
	"strconv"
	"strings"
	"sync"

	"gobox/sprites"
)

// AmigaArt reports whether the Amiga art is showing (F12); nil: never.
// The front sets it; the members' pool ids stay the DOS names, only the
// drawing swaps them for their Amiga twins (amigaTwin).
var AmigaArt func() bool

// poolImage is pool picture name of kind, its Amiga twin while the Amiga art
// shows and it has one.
func poolPicture(kind, name string) image.Image {
	if AmigaArt != nil && AmigaArt() {
		if t := amigaTwin(name); t != "" {
			if im := sprites.Image(Series, kind, t); im != nil {
				return im
			}
		}
	}
	return sprites.Image(Series, kind, name)
}

var (
	twinsOnce sync.Once
	twins     map[string]string
)

// amigaTwin is the amiga_ pool name drawn where the game's DOS file and block
// of pool name dos show on the Amiga ("": none): the same block in the set's
// own Amiga archive (HEAD5), else in the one archive of its kind (HEAD), else
// for an action-pose CHEAD (block 128 on) the one head the Amiga draws for
// both poses.
func amigaTwin(dos string) string {
	twinsOnce.Do(func() {
		twins = map[string]string{}
		o := iconOrigins()
		d, a := o[GameID], o[GameID+"-amiga"]
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
	})
	return twins[dos]
}
