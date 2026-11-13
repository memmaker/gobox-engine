package importer

import (
	"fmt"
	"gobox/rules1e"
	"strconv"
	"strings"

	"gobox/asset/orig"
	"gobox/ecl"
)

// TreasureItems lists the items of set's ITEM block, for a comment.
func TreasureItems(l *orig.Loader, set, block int) string {
	var s []string
	for _, it := range rules1e.ReadItems(l.BlockShared(fmt.Sprintf("ITEM%d", set), block)) {
		n := itemFullName(&it)
		if it.Count > 1 {
			n = strconv.Itoa(int(it.Count)) + " " + n
		}
		s = append(s, n)
	}
	if len(s) > 6 {
		s = append(s[:6], fmt.Sprintf("and %d more", len(s)-6))
	}
	return strings.Join(s, ", ")
}

// itemFullName is an item's words, all known.
func itemFullName(it *rules1e.Item) string {
	var w []string
	for i := 2; i >= 0; i-- {
		if it.Words[i] != 0 {
			w = append(w, strings.TrimSpace(rules1e.ItemWords[it.Words[i]]))
		}
	}
	return strings.Join(w, " ")
}

// ImportSource is the game data the script import reads.
func ImportSource(l *orig.Loader) ecl.Source {
	return ecl.Source{
		Block: func(disk, block int) []byte {
			name := fmt.Sprintf("ECL%d", disk)
			if !l.Exists(name + ".DAX") {
				return nil
			}
			return l.BlockShared(name, block)
		},
		Treasure: func(set, block int) string { return TreasureItems(l, set, block) },
	}
}
