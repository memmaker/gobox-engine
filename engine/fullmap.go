package engine

import (
	"bufio"
	"fmt"
	"gobox/adventure"
	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"log"
	"os"
	"sort"
	"strings"
)

// The full-screen map (extra): M on the command bar
// or a click on the minimap. It shows the minimap's area at 10 pixels a
// square with a letter on every labelled square and a legend beside it. A
// cursor (arrows, click) picks a square; Walk goes there (walk.go), Label
// names it, Remove clears it.
// The game's fixed labels (GameTables.MapLabels; POR: New Phlan's buildings) are always shown;
// the player's labels go over them and are kept with the save (labelsFile).

type mapLabel struct {
	text string
	kind byte
}

// labelsFile holds a slot's map labels (the player's, by map).
const labelsFile = "LABELS.TXT"

// builtinLabels are map k's fixed labels (GameTables.MapLabels), by square.
func (e *Engine) builtinLabels(k int) map[int]mapLabel {
	m := map[int]mapLabel{}
	for _, l := range e.game().MapLabels {
		if l.Area == k {
			m[l.Y<<4|l.X] = mapLabel{l.Text, l.Kind}
		}
	}
	return m
}

// userLabels are the player's labels of the current map (ext.labels).
func (e *Engine) userLabels() map[int]mapLabel {
	k := e.areaKey()
	if e.ext.labels == nil {
		e.ext.labels = map[int]map[int]mapLabel{}
	}
	if e.ext.labels[k] == nil {
		e.ext.labels[k] = map[int]mapLabel{}
	}
	return e.ext.labels[k]
}

// readLabels reads a slot's labelsFile: "map KEY" (areaKey) starts a map,
// then "X Y KIND TEXT" per label.
func (e *Engine) readLabels(path string) {
	e.ext.labels = nil
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	e.ext.labels = map[int]map[int]mapLabel{}
	var cur map[int]mapLabel
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var k, x, y, kind int
		line := sc.Text()
		if _, err := fmt.Sscanf(line, "map %d", &k); err == nil {
			cur = map[int]mapLabel{}
			e.ext.labels[k] = cur
			continue
		}
		var text string
		if n, _ := fmt.Sscanf(line, "%d %d %d", &x, &y, &kind); n == 3 && cur != nil && rules.OnMap(x, y) && kind >= 0 && kind <= adventure.PoiCivic {
			if f := strings.SplitN(line, " ", 4); len(f) == 4 {
				text = f[3]
			}
			cur[y<<4|x] = mapLabel{text, byte(kind)}
		}
	}
}

// writeLabels writes the player's labels to path, but those that are built in.
func (e *Engine) writeLabels(path string) error {
	var b strings.Builder
	keys := make([]int, 0, len(e.ext.labels))
	for k := range e.ext.labels {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "map %d\n", k)
		m, built := e.ext.labels[k], e.builtinLabels(k)
		sq := make([]int, 0, len(m))
		for s, l := range m {
			if b, ok := built[s]; !ok || b != l {
				sq = append(sq, s)
			}
		}
		sort.Ints(sq)
		for _, s := range sq {
			fmt.Fprintf(&b, "%d %d %d %s\n", s&15, s>>4, m[s].kind, m[s].text)
		}
	}
	return e.writeFile(path, []byte(b.String()))
}

// Mark labels square x, y of the current map with text (the player's
// label); nothing typed or a square off the map is refused.
func (e *Engine) Mark(x, y int, text string) error {
	if text = strings.TrimSpace(text); text == "" || !rules.OnMap(x, y) {
		return &game.Refusal{} // nothing to mark: refused, no toast
	}
	e.userLabels()[y<<4|x] = mapLabel{rules.Cut(text, 0x20), adventure.PoiUser}
	return nil
}

// Unmark removes the player's label from square x, y of the current map;
// a square without one is refused.
func (e *Engine) Unmark(x, y int) error {
	m := e.userLabels()
	if _, ok := m[y<<4|x]; !ok || !rules.OnMap(x, y) {
		return refuseIf(true, "no mark there")
	}
	delete(m, y<<4|x)
	return nil
}

// mapName is the current map's name in words, from the scripts' names of
// the GEO blocks (names/pictures.txt), or "Map".
func (e *Engine) mapName() string {
	vals, err := ecl.ReadValues(os.DirFS(e.scriptDir()), "names/pictures.txt", nil)
	if err != nil {
		log.Printf("map names: %v", err)
	}
	n := ecl.ValueName(vals, "Map", int(e.set), e.worldWord(rules.WordGeo))
	if n == "" {
		return "Map"
	}
	var b strings.Builder
	for i, c := range n { // "ValjevoCastle3" -> "Valjevo Castle 3"
		if i > 0 && (c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' && !(n[i-1] >= '0' && n[i-1] <= '9')) {
			b.WriteByte(' ')
		}
		b.WriteRune(c)
	}
	return b.String()
}
