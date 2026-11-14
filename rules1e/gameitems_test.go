package rules1e

import (
	"encoding/json"
	"gobox/asset"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"testing"
)

// itemIconRec is the pool's items.rec as ItemID -> cell (internal_name is the id).
func itemIconRec(t *testing.T) map[int]int {
	_, rec, err := asset.ItemIcons("pool")
	if err != nil {
		t.Fatal(err)
	}
	m := map[int]int{}
	for _, r := range regexp.MustCompile(`internal_name: (\d+)\nname: .*\nicon: (-?\d+)`).FindAllStringSubmatch(rec, -1) {
		id, _ := strconv.Atoi(r[1])
		m[id], _ = strconv.Atoi(r[2])
	}
	return m
}

// TestItemIDsUnique: no two base item kinds share a dnd id (dnditems.go).
func TestItemIDsUnique(t *testing.T) {
	b, err := os.ReadFile("dnditems.go")
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]string{}
	for _, m := range regexp.MustCompile(`(?m)^\tItemID(\w+)\s+rules\.ItemID = (\d+)`).FindAllStringSubmatch(string(b), -1) {
		if o, dup := seen[m[2]]; dup {
			t.Errorf("ItemID %s: %s and %s", m[2], o, m[1])
		}
		seen[m[2]] = m[1]
	}
}

// TestGameItems: every item of each game's data (ITEM blocks, PoD's item
// library) has a base kind with an icon cell, a name of its own game's
// words and an inventory tab.
func TestGameItems(t *testing.T) {
	icons := itemIconRec(t)
	for _, g := range []string{"por", "cotab", "sotsb", "pod"} {
		dir := "../data/" + g + "/data/"
		files, _ := filepath.Glob(dir + "ITEM[0-9]*.json")
		if len(files) == 0 {
			t.Logf("no imported %s", g)
			continue
		}
		tb := POR
		var words []string
		var library []Item
		if _, err := os.Stat(dir + "rules.json"); err == nil { // POR: its own tables
			l, err := ReadLayout(dir + LayoutFile)
			if err != nil {
				t.Fatal(g, err)
			}
			p, err := POR.LoadRules(dir+"rules.json", l)
			if err != nil {
				t.Fatal(g, err)
			}
			tb = *p
			b, _ := os.ReadFile(dir + "rules.json")
			var raw struct {
				Words   []string `json:"itemWords"`
				Library []Item   `json:"itemLibrary"`
			}
			if err := json.Unmarshal(b, &raw); err != nil {
				t.Fatal(g, err)
			}
			words, library = raw.Words, raw.Library
		}
		items := library
		for _, f := range files {
			b, err := os.ReadFile(f)
			if err != nil {
				t.Fatal(err)
			}
			var blocks map[string][]Item
			if json.Unmarshal(b, &blocks) != nil {
				var one []Item
				if err := json.Unmarshal(b, &one); err != nil {
					t.Fatal(f, err)
				}
				blocks = map[string][]Item{"": one}
			}
			for _, l := range blocks {
				items = append(items, l...)
			}
		}
		r := &RuleState{Tables: &tb}
		bad := map[ItemType]string{}
		for i := range items {
			it := &items[i]
			k := r.kind(it.Type)
			if _, ok := icons[int(k)]; k == 0 || !ok || icons[int(k)] < 0 {
				bad[it.Type] = "no kind with an icon (kind " + strconv.Itoa(int(k)) + ")"
				continue
			}
			r.ItemName(nil, it, false)
			if it.Name() == "" {
				bad[it.Type] = "no name"
			}
			for _, w := range it.Words {
				if w != 0 && words != nil && (int(w) >= len(words) || r.itemWord(w) != words[w]) {
					bad[it.Type] = "word " + strconv.Itoa(int(w)) + " not the game's"
				}
			}
			if tab, _ := r.ItemTab(&Character{}, it); tab == "" {
				bad[it.Type] = "no tab"
			}
		}
		for ty, why := range bad {
			t.Errorf("%s item type %d: %s", g, ty, why)
		}
	}
}
