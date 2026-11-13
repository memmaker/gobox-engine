// Command idgen gives a game's spell and item numbers their dnd ids
// (ARCHITECTURE.md section 4). It reads the game's spec, one id a line:
//
//	spell 3 CureLightWounds
//	item 0x09 Dart
//
// and writes three files: rules1e's dnditems.go and dndspells.go, where an id
// keeps the number (or string) it was given once and a new one is added
// (an item at the next number), and the game's map from its numbers to the
// ids. It runs at `go generate` time only.
//
//	idgen -spec ids.txt -dir DIR -spells spellIDs -items itemIDs -itemtype itemType > ids.go
package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// noNumber is the number of an id the game has none for: "spell - StoneGaze".
const noNumber = -1

// entry is one line of the spec: the game's number and the dnd name.
type entry struct {
	num  int
	name string
}

func main() {
	spec := flag.String("spec", "", "the game's spec")
	dir := flag.String("dir", "", "the rules1e directory (dndspells.go and dnditems.go)")
	pkg := flag.String("pkg", "rules1e", "the game file's package")
	spellVar := flag.String("spells", "spellIDs", "the game's spell table: [n]core.SpellID by spell number")
	itemVar := flag.String("items", "itemIDs", "the game's item table: map[itemtype]core.ItemID")
	itemType := flag.String("itemtype", "itemType", "the game's base item type")
	flag.Parse()
	if *spec == "" || *dir == "" {
		log.Fatal("usage: idgen -spec ids.txt -dnd DIR [-pkg rules] > ids.go")
	}
	spells, items, err := readSpec(*spec)
	if err != nil {
		log.Fatal(err)
	}
	if err := writeSpells(filepath.Join(*dir, "dndspells.go"), spells); err != nil {
		log.Fatal(err)
	}
	if err := writeItems(filepath.Join(*dir, "dnditems.go"), items); err != nil {
		log.Fatal(err)
	}
	src, err := gameFile(*pkg, *spellVar, *itemVar, *itemType, spells, items)
	if err != nil {
		log.Fatal(err)
	}
	os.Stdout.Write(src)
}

func readSpec(path string) (spells, items []entry, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		l := strings.TrimSpace(sc.Text())
		if l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		fs := strings.Fields(l)
		if len(fs) != 3 {
			return nil, nil, fmt.Errorf("%s:%d: want KIND NUMBER NAME", path, n)
		}
		num := int64(noNumber)
		if fs[1] != "-" { // a power with no number of the game's (a gaze, a breath)
			num, err = strconv.ParseInt(fs[1], 0, 32)
		}
		if err != nil {
			return nil, nil, fmt.Errorf("%s:%d: %w", path, n, err)
		}
		e := entry{int(num), fs[2]}
		switch fs[0] {
		case "spell":
			spells = append(spells, e)
		case "item":
			items = append(items, e)
		default:
			return nil, nil, fmt.Errorf("%s:%d: kind %q: spell or item", path, n, fs[0])
		}
	}
	if err := sc.Err(); err != nil {
		return nil, nil, err
	}
	for _, l := range [][]entry{spells, items} {
		nums, names := map[int]bool{}, map[string]bool{}
		for _, e := range l {
			if e.num != noNumber && nums[e.num] || names[e.name] {
				return nil, nil, fmt.Errorf("%s: %d %s twice", path, e.num, e.name)
			}
			nums[e.num], names[e.name] = true, true
		}
	}
	return spells, items, nil
}

var (
	spellLine = regexp.MustCompile(`^\tSpell(\w+)\s+core\.SpellID = "([^"]+)"`)
	itemLine  = regexp.MustCompile(`^\tItemID(\w+)\s+core\.ItemID = (\d+)`)
)

// existing reads a dnd file: its header (what is before "const (") and its ids.
func existing(path string, re *regexp.Regexp) (header string, ids map[string]string, err error) {
	ids = map[string]string{}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", ids, nil
	}
	if err != nil {
		return "", nil, err
	}
	head, _, _ := strings.Cut(string(b), "const (")
	for _, l := range strings.Split(string(b), "\n") {
		if m := re.FindStringSubmatch(l); m != nil {
			ids[m[1]] = m[2]
		}
	}
	return head, ids, nil
}

// kebab is a new spell's string: CureLightWounds cure-light-wounds, ProtectFromEvilMU protect-from-evil-mu.
func kebab(name string) string {
	var b strings.Builder
	r := []rune(name)
	for i, c := range r {
		if i > 0 && unicode.IsUpper(c) && !unicode.IsUpper(r[i-1]) {
			b.WriteByte('-')
		}
		b.WriteRune(unicode.ToLower(c))
	}
	return b.String()
}

func writeSpells(path string, spells []entry) error {
	head, ids, err := existing(path, spellLine)
	if err != nil {
		return err
	}
	if head == "" {
		head = "package rules1e\n\nimport \"gobox/core\"\n\n"
	}
	for _, e := range spells {
		if _, ok := ids[e.name]; !ok {
			ids[e.name] = kebab(e.name)
		}
	}
	names := make([]string, 0, len(ids))
	for n := range ids {
		names = append(names, n)
	}
	sort.Strings(names)
	var b bytes.Buffer
	b.WriteString(head + "const (\n")
	for _, n := range names {
		fmt.Fprintf(&b, "\tSpell%s core.SpellID = %q\n", n, ids[n])
	}
	b.WriteString(")\n")
	return writeGo(path, b.Bytes())
}

func writeItems(path string, items []entry) error {
	head, ids, err := existing(path, itemLine)
	if err != nil {
		return err
	}
	if head == "" {
		head = "package rules1e\n\nimport \"gobox/core\"\n\n// The base items, numbered once (idgen): a number is never reused or changed.\n\n"
	}
	num := map[string]int{}
	last := 0
	for n, v := range ids {
		num[n], _ = strconv.Atoi(v)
		last = max(last, num[n])
	}
	for _, e := range items {
		if _, ok := num[e.name]; !ok {
			last++
			num[e.name] = last
		}
	}
	names := make([]string, 0, len(num))
	for n := range num {
		names = append(names, n)
	}
	sort.Slice(names, func(i, j int) bool { return num[names[i]] < num[names[j]] })
	var b bytes.Buffer
	b.WriteString(head + "const (\n")
	for _, n := range names {
		fmt.Fprintf(&b, "\tItemID%s core.ItemID = %d\n", n, num[n])
	}
	b.WriteString(")\n")
	return writeGo(path, b.Bytes())
}

func gameFile(pkg, spellVar, itemVar, itemType string, spells, items []entry) ([]byte, error) {
	n := 0
	for _, e := range spells {
		n = max(n, e.num+1)
	}
	var b bytes.Buffer
	fmt.Fprintf(&b, "// Code generated by idgen from ids.txt; DO NOT EDIT.\n\npackage %s\n\nimport (\n\t\"gobox/core\"\n)\n\n", pkg)
	fmt.Fprintf(&b, "// %s is the dnd id of each spell number (empty: none).\nvar %s = [%d]core.SpellID{\n", spellVar, spellVar, n)
	for _, e := range spells {
		if e.num != noNumber {
			fmt.Fprintf(&b, "\t%d: Spell%s,\n", e.num, e.name)
		}
	}
	fmt.Fprintf(&b, "}\n\n// %s is the dnd id of each base item type named.\nvar %s = map[%s]core.ItemID{\n", itemVar, itemVar, itemType)
	for _, e := range items {
		if e.num == noNumber {
			continue
		}
		fmt.Fprintf(&b, "\t0x%02X: ItemID%s,\n", e.num, e.name)
	}
	b.WriteString("}\n")
	return format.Source(b.Bytes())
}

func writeGo(path string, src []byte) error {
	out, err := format.Source(src)
	if err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return os.WriteFile(path, out, 0o644)
}
