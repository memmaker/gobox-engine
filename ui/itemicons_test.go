package ui

import (
	"bufio"
	"gobox/asset"
	"strconv"
	"strings"
	"testing"
)

// TestItemIconRecs: each series' items.rec is in remapper's classic format
// (every record has internal_name, name and icon), its ids are unique and
// every icon is a cell of its atlas (-1: no picture yet). The pool's parse
// as rules.ItemIDs (readIcons).
func TestItemIconRecs(t *testing.T) {
	for _, s := range []string{"pool", "buck-rogers"} {
		atlas, rec, err := asset.ItemIcons(s)
		if err != nil {
			t.Fatal(s, err)
		}
		b := decodePNG(atlas).Bounds()
		cell := b.Dx() / atlasCols
		cells := atlasCols * (b.Dy() / cell)
		seen := map[string]bool{}
		check := func(r map[string]string) {
			if len(r) == 0 {
				return
			}
			id := r["internal_name"]
			if id == "" || r["name"] == "" || r["icon"] == "" {
				t.Errorf("%s: record %v lacks internal_name, name or icon", s, r)
			}
			if seen[id] {
				t.Errorf("%s: %q twice", s, id)
			}
			seen[id] = true
			if n, err := strconv.Atoi(r["icon"]); err != nil || n < -1 || n >= cells {
				t.Errorf("%s: %q icon %q not a cell of %d", s, id, r["icon"], cells)
			}
		}
		r := map[string]string{}
		sc := bufio.NewScanner(strings.NewReader(rec))
		for sc.Scan() {
			l := sc.Text()
			if strings.TrimSpace(l) == "" {
				check(r)
				r = map[string]string{}
				continue
			}
			if k, v, ok := strings.Cut(l, ":"); ok && !strings.HasPrefix(k, "#") {
				r[k] = strings.TrimSpace(v)
			}
		}
		check(r)
		if s == "pool" {
			if _, err := readIcons(rec); err != nil {
				t.Error(err)
			}
		}
	}
}
