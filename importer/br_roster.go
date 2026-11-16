package importer

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gobox/adventure"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// The shipped characters are the roster a new game starts with. In the original they are the CHRDATAn.SAV files the
// installation puts into the save folder (buckrogers-go rules/rules.go seedSaves); the engine's roster is charlist.txt
// and one <NAME>.cha per character (engine/roster.go), and a new save folder is seeded from the adventure's saves/
// (cmd/gobox homeSaves). So saves/ holds the shipped characters in the engine's format, not the originals' bytes.
const (
	rosterListFile = "charlist.txt" // the engine's roster list (engine.WriteCharList: a name and CRLF a line)
	rosterExt      = ".cha"         // the engine's roster file of a character (engine.rosterFile)
)

// writeRoster writes saves/charlist.txt and saves/<NAME>.cha for each character in data/characters (written by
// importBRSaves), in the originals' file number order (CHRDATA2 before CHRDATA10), one per name (see below), and removes the originals' CHRDATA*
// copies from saves/. The files carry no roster number: the engine numbers them when it first reads the list.
func writeRoster(dst string) error {
	chars := filepath.Join(dst, adventure.DataDir, CharactersDir)
	ents, err := os.ReadDir(chars)
	if err != nil {
		return err
	}
	var files []string
	for _, e := range ents {
		if stem, ok := strings.CutSuffix(e.Name(), ".json"); ok && !e.IsDir() {
			files = append(files, stem)
		}
	}
	slices.SortFunc(files, func(a, b string) int { // CHRDATA2 before CHRDATA10: the shorter number first
		if len(a) != len(b) {
			return len(a) - len(b)
		}
		return strings.Compare(a, b)
	})
	saves := filepath.Join(dst, SavesDir)
	if err := os.MkdirAll(saves, 0o755); err != nil {
		return err
	}
	var list strings.Builder
	named := map[string]string{} // .cha base -> the name that took it
	for _, stem := range files {
		rec, err := os.ReadFile(filepath.Join(chars, stem+".json"))
		if err != nil {
			return err
		}
		r := rulesxxvc.New()
		file, err := r.Pack(rules.CharacterFiles{Record: rec})
		if err != nil {
			return fmt.Errorf("%s: %w", stem, err)
		}
		id, err := r.Import(file)
		if err != nil {
			return fmt.Errorf("%s: %w", stem, err)
		}
		name := r.Sheet(id).Name
		base := adventure.MakeFileName(rules.RemoveSubstr(name, "."))
		if _, dup := named[base]; dup {
			continue // ponytail: the engine's roster is keyed by name (one .cha per name); CHRDATA has 23 premade characters of 7 names (several rolls each). The first of a name is listed; the others need a roster keyed by something else. The Pascal never reads CHRDATA (the roster is its *.who files), so the evidence for the original's list is silent.
		}
		named[base] = name
		if err := os.WriteFile(filepath.Join(saves, base+rosterExt), file, 0o644); err != nil {
			return err
		}
		list.WriteString(name + "\r\n")
	}
	if err := os.WriteFile(filepath.Join(saves, rosterListFile), []byte(list.String()), 0o644); err != nil {
		return err
	}
	old, err := filepath.Glob(filepath.Join(saves, "CHRDATA*"))
	if err != nil {
		return err
	}
	for _, p := range old {
		if err := os.Remove(p); err != nil {
			return err
		}
	}
	return nil
}
