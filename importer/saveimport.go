package importer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobox/adventure"
	"gobox/ecl"
	"gobox/rules"
	"gobox/rules1e"
)

// The original's saves become canonical saves here, once, at import: the
// engine reads only adventure.Save.

// origSaves matches an original lettered save, SAVGAMA.DAT..SAVGAMJ.DAT (ovr18 SaveGame).
const origSaves = "[Ss][Aa][Vv][Gg][Aa][Mm][A-Ja-j].[Dd][Aa][Tt]"

// importSaves turns each original save the game's import copied into adv's
// SavesDir into the slot "GAME <letter>" (SlotsDir): its canonical save and
// copies of its members' files. The original leaves the folder.
func importSaves(adv string) error {
	dir := filepath.Join(adv, adventure.SavesDir)
	found, err := filepath.Glob(filepath.Join(dir, origSaves))
	if err != nil || len(found) == 0 {
		return err
	}
	t, err := adventure.ReadGameTables(adv)
	if err != nil {
		return err
	}
	scripts := os.DirFS(filepath.Join(adv, scriptsDir))
	members, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, path := range found {
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		g, err := ImportSave(b, t.OriginalSave, scripts)
		if err != nil {
			return fmt.Errorf("%s: %w", filepath.Base(path), err)
		}
		letter := strings.ToUpper(filepath.Base(path)[len("SAVGAM") : len("SAVGAM")+1])
		slot := filepath.Join(dir, adventure.SlotsDir, "GAME "+letter)
		if err := os.MkdirAll(slot, 0o755); err != nil {
			return err
		}
		for _, m := range members {
			base, _, _ := strings.Cut(m.Name(), ".")
			if m.IsDir() || !slices.ContainsFunc(g.Party, func(p string) bool { return strings.EqualFold(p, base) }) {
				continue
			}
			mb, err := os.ReadFile(filepath.Join(dir, m.Name()))
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(slot, m.Name()), mb, 0o644); err != nil {
				return err
			}
		}
		if err := adventure.WriteSave(filepath.Join(slot, adventure.GameFile), g); err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// ImportSave turns an original save image b laid out as s into the canonical
// save: rules1e reads the image; the fields' cells come out of its record
// words by their layout addresses (scripts' game fields, and the locals of
// the area it was saved in).
func ImportSave(b []byte, s rules1e.SaveLayout, scripts fs.FS) (adventure.Save, error) {
	o, err := rules1e.ReadOriginalSave(b, s)
	if err != nil {
		return adventure.Save{}, err
	}
	g := adventure.Save{
		Set: rules.Set(o.Set), Fields: map[string][]int64{}, Locals: map[string][]int64{},
		X: o.X, Y: o.Y, Facing: o.Facing, Pos3: o.Pos3, Pos4: o.Pos4,
		ViewMode: o.ViewMode, BeginFlag: o.BeginFlag,
	}
	fields, err := ecl.GameFields(scripts)
	if err != nil {
		return g, err
	}
	for _, fl := range fields {
		if c := decodeCells(&o, fl); c != nil && saved(fl) {
			g.Fields[fl.Name] = c
		}
	}
	l, err := ecl.ReadLayout(scripts, int(o.AreaKind))
	if err != nil {
		return g, err
	}
	if l == nil {
		return g, fmt.Errorf("area %d: no layout", o.AreaKind)
	}
	if g.Rules, err = rulesOf(&o, s, l); err != nil {
		return g, err
	}
	for _, fl := range l.Fields {
		if c := decodeCells(&o, fl); c != nil && fl.Owner == ecl.OwnerLocal && saved(fl) {
			g.Locals[fl.Name] = c
		}
	}
	for _, n := range o.Party {
		g.Party = append(g.Party, adventure.MakeFileName(rules.RemoveSubstr(n, ".")))
	}
	return g, nil
}

// rulesOf is the ruleset's part of o, whose script fields are l: its
// records when they are rules1e's, else made from the fields by name.
func rulesOf(o *rules1e.ImportedSave, s rules1e.SaveLayout, l *ecl.Layout) ([]byte, error) {
	if s.RulesRecords {
		return o.Rules, nil
	}
	return rules1e.RulesFromFields(canonFields(o, l), o.Purse)
}

// originalRules is the ruleset's part of the original save b laid out as s
// whose script fields are l (a game's test against its real saves).
func originalRules(b []byte, s rules1e.SaveLayout, l *ecl.Layout) ([]byte, error) {
	o, err := rules1e.ReadOriginalSave(b, s)
	if err != nil {
		return nil, err
	}
	return rulesOf(&o, s, l)
}

// saved: field f keeps cells of its own in a save. The engine also binds
// some without cells; it drops those on load.
func saved(f ecl.Field) bool {
	switch f.Owner {
	case ecl.OwnerGame, ecl.OwnerArea, ecl.OwnerMember, ecl.OwnerLocal:
		return !f.Replaces
	}
	return false
}

// decodeCells are field fl's cells in o's records (a text up to its NUL,
// a text buffer with named elements all its Len); nil when all are zero.
func decodeCells(o *rules1e.ImportedSave, fl ecl.Field) []int64 {
	var c []int64
	n := fl.Len
	text := fl.Width == ecl.WidthText && fl.Elements == nil // a text buffer with named elements keeps all its cells
	if text {
		n = 0x100
	}
	for i := range n {
		p := o.Cell(fl.From + i)
		if p == nil || text && *p == 0 {
			break
		}
		c = append(c, int64(*p))
	}
	if !slices.ContainsFunc(c, func(v int64) bool { return v != 0 }) {
		return nil
	}
	return c
}

// canonFields are the canonical values (adventure.Canonical) of l's game and
// area fields in o, by name: what rules1e.RulesFromFields reads.
func canonFields(o *rules1e.ImportedSave, l *ecl.Layout) map[string][]int64 {
	out := map[string][]int64{}
	for _, fl := range l.Fields {
		if fl.Owner != ecl.OwnerGame && fl.Owner != ecl.OwnerArea {
			continue
		}
		c := decodeCells(o, fl)
		for k := range c {
			c[k] = adventure.Canonical(fl.Encoding, c[k], rules.Set(o.Set))
		}
		if c != nil {
			out[fl.Name] = c
		}
	}
	return out
}
