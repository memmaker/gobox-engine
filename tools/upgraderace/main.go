// Command upgraderace rewrites the imported monster files (data/<game>/data/MON<n>.json
// and the por sources/assets/data copies) from the old format, where a
// monster's Record.Race is the game's race number, to the new one, where it
// is the race's name and Record.MonsterRace keeps the number of a monster
// that names no player race. One-time; the files are generated, so
// re-importing the game gives the same result.
//
//	go run ./tools/upgraderace <layout> <file>...
//
// layout is the game's character record size: 285 (por), 422 (cotab), 439
// (sotsb) or 510 (pod). A file already in the new format is left alone.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"gobox/rules"
	"gobox/rules1e"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: upgraderace <layout 285|422|439|510> <MON*.json>...")
		os.Exit(2)
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "layout:", err)
		os.Exit(2)
	}
	for _, path := range os.Args[2:] {
		if err := upgradeFile(rules.RecordLayout(n), path); err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
	}
}

func upgradeFile(l rules.RecordLayout, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out, changed, err := upgrade(l, b)
	if err != nil || !changed {
		return err
	}
	return os.WriteFile(path, out, 0o644)
}

// upgrade is a monster file (monster id -> {Record, Affects, Items}) with
// every Record's race number turned into the race; false where none was.
func upgrade(l rules.RecordLayout, b []byte) ([]byte, bool, error) {
	var file map[string]map[string]json.RawMessage
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, false, err
	}
	changed := false
	for id, m := range file {
		var rec map[string]json.RawMessage
		if err := json.Unmarshal(m["Record"], &rec); err != nil {
			return nil, false, fmt.Errorf("monster %s: %w", id, err)
		}
		var number byte
		if err := json.Unmarshal(rec["Race"], &number); err != nil {
			continue // not a number: already the race's name
		}
		race, monster := rules1e.RaceOfRecord(l, number)
		rec["Race"], _ = json.Marshal(race) // a string: cannot fail
		rec["MonsterRace"], _ = json.Marshal(monster)
		var err error
		if m["Record"], err = json.Marshal(rec); err != nil {
			return nil, false, err
		}
		changed = true
	}
	if !changed {
		return b, false, nil
	}
	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetIndent("", "\t")
	if err := enc.Encode(file); err != nil {
		return nil, false, err
	}
	return out.Bytes(), true, nil
}
