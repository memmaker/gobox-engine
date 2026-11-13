package importer

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"gobox/asset"
	"gobox/asset/orig"
)

// fontFiles are the originals that may hold a game's text font
// (asset.FontBlock): por and cotab 8X8D1, sotsb 8X8D5, pod 8X8D0.
var fontFiles = []string{"8X8D*.DAX"}

// poolFont adds g's text font, block asset.FontBlock of whichever of the
// originals in src holds it, to root's font pool (asset.AddFont), named
// after g's title; originals without one add none.
func poolFont(g Game, src, root string) error {
	ents, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range ents {
		if e.IsDir() || !matches(e.Name(), fontFiles) {
			continue
		}
		f, err := orig.Open(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if b := f.BlockShared(asset.FontBlock); len(b) >= 64*8 {
			return asset.AddFont(root, g.ID, fontName(g.Title), asset.F8(b))
		}
	}
	return nil
}

// fontName is title as a pool font's name: letters and digits, the rest "_".
func fontName(title string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, title)
}
