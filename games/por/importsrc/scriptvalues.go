package importsrc

import (
	"fmt"
	"gobox/rules1e"
	"slices"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"gobox/asset/orig"
	"gobox/ecl/script"
)

// DataValues names set's monsters (MON<set>CHA records) and item types
// (the items of ITEM<set>), for names/values.txt. A name two numbers
// share gets the number on.
func DataValues(l *orig.Loader, set int) []script.Value {
	var out []script.Value
	add := func(kind string, names map[int]string) {
		count := map[string]int{}
		for _, n := range names {
			count[n]++
		}
		var ids []int
		for id := range names {
			ids = append(ids, id)
		}
		sort.Ints(ids)
		for _, id := range ids {
			n := names[id]
			if count[n] > 1 {
				n += strconv.Itoa(id)
			}
			out = append(out, script.Value{Kind: kind, Set: set, N: id, Name: n})
		}
	}
	mons := map[int]string{}
	for id := range 256 {
		if r := l.BlockShared(fmt.Sprintf("MON%dCHA", set), id); len(r) > 0 {
			c := &rules1e.Character{}
			c.Sheet = rules1e.SheetOf(r)
			if n := camelName(c.Name()); n != "" {
				mons[id] = n
			}
		}
	}
	add("Monster", mons)
	items := map[int]string{} // the shortest name of an item of the type
	for id := range 0x80 {
		for _, it := range rules1e.ReadItems(l.BlockShared(fmt.Sprintf("ITEM%d", set), id)) {
			n := camelName(itemFullName(&it))
			if old, ok := items[int(it.Type)]; n != "" && (!ok || len(n) < len(old)) {
				items[int(it.Type)] = n
			}
		}
	}
	add("Item", items)
	for id := range 0x80 { // the stocks, an ITEM block's items by name
		var s []string
		for _, it := range rules1e.ReadItems(l.BlockShared(fmt.Sprintf("ITEM%d", set), id)) {
			n := camelName(itemFullName(&it))
			if it.Count > 1 {
				n += "*" + strconv.Itoa(int(it.Count))
			}
			s = append(s, n)
		}
		if len(s) > 0 && !slices.Contains(s, "") {
			out = append(out, script.Value{Kind: "Stock", Set: set, N: id, Name: strings.Join(s, ",")})
		}
	}
	return out
}

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

// camelName is a game name as one script word: "KOBOLD LEADER" ->
// KoboldLeader.
func camelName(s string) string {
	var b strings.Builder
	up := true
	for _, r := range strings.ToLower(s) {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) && b.Len() > 0:
			if up {
				r = unicode.ToUpper(r)
			}
			b.WriteRune(r)
			up = false
		default:
			up = true
		}
	}
	return b.String()
}

// ImportSource is the game data the script import reads.
func ImportSource(l *orig.Loader) script.Source {
	return script.Source{
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
