package pool

import (
	"bufio"
	"fmt"
	"gobox/core"
	"gobox/core/goldbox"
	"gobox/core/grid"
	"gobox/ecl/script"
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

// Label kinds, each with its colour on the maps (LabelInks, present.Data).
const (
	poiUser   = 0
	poiShop   = 1
	poiTemple = 2
	poiInn    = 3
	poiCivic  = 4
)

type mapLabel struct {
	text string
	kind byte
}

// MapLabel is a fixed label of a map square, the game's data (GameTables.MapLabels;
// POR: New Phlan's buildings): its text and kind, and the soundtrack song
// played there (0: the area's).
type MapLabel struct {
	Area int // areaKey: data set<<16 | GEO block
	X, Y int
	Text string
	Kind byte // poi*
	Song int
}

// PorMapLabels are the buildings of New Phlan (data set 3, GEO block 0;
// ECL3 block 0): the squares whose event id (cell byte & 0x7F, the
// script's [6E82]) leads to a shop, temple, inn, tavern or office in the
// block's step event (ECL 9B4C ONGOTO).
var PorMapLabels = func() []MapLabel {
	var m []MapLabel
	add := func(text string, kind byte, song int, sq ...[2]int) {
		for _, s := range sq {
			m = append(m, MapLabel{Area: 3 << 16, X: s[0], Y: s[1], Text: text, Kind: kind, Song: song})
		}
	}
	add("Boats", poiCivic, 0, [2]int{15, 1})                                                // 1
	add("Harbor master", poiCivic, 0, [2]int{11, 1})                                        // 2
	add("Bishop Braccio", poiTemple, songTemple, [2]int{10, 5})                             // 5
	add("Temple of Tyr", poiTemple, songTemple, [2]int{11, 3}, [2]int{8, 4}, [2]int{8, 6})  // 6
	add("Temple of Sune", poiTemple, songTemple, [2]int{1, 1}, [2]int{3, 1}, [2]int{1, 3})  // 7
	add("Inn", poiInn, songInn, [2]int{4, 12}, [2]int{6, 12}, [2]int{4, 13}, [2]int{6, 13}, // 9
		[2]int{0, 14}, [2]int{1, 14}, [2]int{2, 14})
	add("Training Hall", poiCivic, songTrainingHall, [2]int{5, 0}, [2]int{7, 0}, [2]int{9, 0})  // 10, 17
	add("Shop", poiShop, songShop, [2]int{15, 8}, [2]int{9, 10}, [2]int{12, 10}, [2]int{9, 11}, // 19
		[2]int{11, 11})
	add("Tavern", poiInn, songTavern, [2]int{10, 8}, [2]int{8, 9}, [2]int{14, 10}, [2]int{15, 10}, // 20
		[2]int{15, 12}, [2]int{15, 13}, [2]int{15, 14})
	add("Jeweler", poiShop, songShop, [2]int{8, 10})                                       // 21
	add("Arms and Armor", poiShop, songShop, [2]int{13, 8}, [2]int{8, 11}, [2]int{11, 12}, // 22
		[2]int{8, 13}, [2]int{9, 13})
	add("Silver Shop", poiShop, songShop, [2]int{11, 10}, [2]int{10, 13})                            // 23
	add("Temple of Tempus", poiTemple, songTemple, [2]int{0, 12}, [2]int{2, 12})                     // 24
	add("Gate to slums", poiCivic, 0, [2]int{0, 4})                                                  // 25
	add("City Hall", poiCivic, songCityHall, [2]int{4, 4}, [2]int{4, 3}, [2]int{3, 4}, [2]int{5, 4}) // 27, 26
	return m
}()

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
		if n, _ := fmt.Sscanf(line, "%d %d %d", &x, &y, &kind); n == 3 && cur != nil && grid.OnMap(x, y) && kind >= 0 && kind <= poiCivic {
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
// label; "" marks nothing).
func (e *Engine) Mark(x, y int, text string) {
	if text = strings.TrimSpace(text); text != "" && grid.OnMap(x, y) {
		e.userLabels()[y<<4|x] = mapLabel{goldbox.Cut(text, 0x20), poiUser}
	}
}

// Unmark removes the player's label from square x, y of the current map.
func (e *Engine) Unmark(x, y int) {
	if grid.OnMap(x, y) {
		delete(e.userLabels(), y<<4|x)
	}
}

// mapName is the current map's name in words, from the scripts' names of
// the GEO blocks (names/pictures.txt), or "Map".
func (e *Engine) mapName() string {
	vals, _ := script.ReadValues(os.DirFS(e.scriptDir()), "names/pictures.txt", nil)
	n := script.ValueName(vals, "Map", int(e.set), e.worldWord(core.WordGeo))
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
