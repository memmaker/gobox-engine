package importer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gobox/core"
)

// Folder is an imported adventure folder: where it is and its manifest.
type Folder struct {
	Dir string
	Manifest
}

// Read reads the adventure folder dir's manifest. It is valid when it has
// this Format, a ruleset and the folder's own name as its ID; nothing else
// in the folder (and no original) is looked at.
func Read(dir string) (Folder, error) {
	b, err := os.ReadFile(filepath.Join(dir, ManifestFile))
	if err != nil {
		return Folder{}, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Folder{}, fmt.Errorf("%s: %w", dir, err)
	}
	switch {
	case m.Format != Format:
		return Folder{}, fmt.Errorf("%s: manifest format %d, want %d (import it again)", dir, m.Format, Format)
	case m.ID != filepath.Base(dir):
		return Folder{}, fmt.Errorf("%s: manifest names %q", dir, m.ID)
	case m.Ruleset == "":
		return Folder{}, fmt.Errorf("%s: manifest names no ruleset", dir)
	}
	return Folder{Dir: dir, Manifest: m}, nil
}

// List is the valid adventure folders under root, by name; a folder without
// a valid manifest is not listed (skipped says why). No root: none.
func List(root string) (folders []Folder, skipped []error, err error) {
	ents, err := os.ReadDir(root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	for _, e := range ents {
		if !e.IsDir() {
			continue
		}
		f, err := Read(filepath.Join(root, e.Name()))
		if err != nil {
			skipped = append(skipped, err)
			continue
		}
		folders = append(folders, f)
	}
	return folders, skipped, nil
}

// CheckRecord reports whether the folder's character records are in layout
// l, the one the engine reads (an older import has none: re-import).
func (f Folder) CheckRecord(l core.RecordLayout) error {
	if f.Manifest.Record != l {
		return fmt.Errorf("adventure %s: character records are layout %d, the engine reads %d: re-import it", f.Dir, f.Manifest.Record, l)
	}
	return nil
}
