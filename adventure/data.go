package adventure

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// Blocks are an imported folder's numbered binary records (a map's GEO
// blocks): dir/<name>.bin holds them one after another, dir/<name>.txt a line
// "block <id> <offset> <length>" each, after '#' comment lines.

// WriteBlocks writes blocks as dir/name.bin and .txt, with comment as the
// first line; no blocks write nothing.
func WriteBlocks(dir, name, comment string, blocks map[int][]byte) error {
	if len(blocks) == 0 {
		return nil
	}
	var bin bytes.Buffer
	var meta strings.Builder
	fmt.Fprintf(&meta, "# %s\n", comment)
	ids := make([]int, 0, len(blocks))
	for id := range blocks {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		fmt.Fprintf(&meta, "block %d %d %d\n", id, bin.Len(), len(blocks[id]))
		bin.Write(blocks[id])
	}
	if err := os.WriteFile(filepath.Join(dir, name+".bin"), bin.Bytes(), 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".txt"), []byte(meta.String()), 0o644)
}

// ReadBlocks reads dir/name.bin and .txt; nil and no error: there are none.
func ReadBlocks(dir, name string) (map[int][]byte, error) {
	bin, err := os.ReadFile(filepath.Join(dir, name+".bin"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	meta, err2 := os.ReadFile(filepath.Join(dir, name+".txt"))
	if err = errors.Join(err, err2); err != nil {
		return nil, err
	}
	out := map[int][]byte{}
	sc := bufio.NewScanner(bytes.NewReader(meta))
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var id, off, n int
		if _, err := fmt.Sscanf(line, "block %d %d %d", &id, &off, &n); err != nil {
			return nil, fmt.Errorf("%s.txt: %q: %w", name, line, err)
		}
		if off < 0 || n < 0 || off+n > len(bin) {
			return nil, fmt.Errorf("%s.txt: %q: past the end of %s.bin", name, line, name)
		}
		out[id] = bin[off : off+n : off+n]
	}
	return out, nil
}

// WriteJSON writes v as dir/name.json, indented.
func WriteJSON(dir, name string, v any) error {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name+".json"), b, 0o644)
}

// ReadJSON reads dir/name.json into v; false and no error: there is none, v
// is left as it is.
func ReadJSON(dir, name string, v any) (bool, error) {
	b, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal(b, v)
}

// Monster is a monster's records: MON<set>CHA, SPC (affects) and ITM (items)
// blocks of one id. MON<n>.json holds it with the record as the rules' sheet
// (orig.ImportData's sheet), so every game's monsters read the same.
type Monster[C, A, I any] struct {
	Record  C
	Affects []A `json:",omitempty"`
	Items   []I `json:",omitempty"`
}

// ItemTableFile is the base item table in an import's data dir (JSON).
const ItemTableFile = "ITEMS"

// RuleTablesFile is the ruleset's own tables in an import's data dir (JSON; XXVc: xxvc.Tables), rules.SourceRuleTables.
const RuleTablesFile = "tables"

// RulesFile is the ruleset's tables in an imported game's data folder.
const RulesFile = "rules.json"
