package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// PrefsName is the settings file's name in Prefs.Dir.
const PrefsName = "SETTINGS.JSON"

// Prefs is the game's settings file: generic JSON key/value data kept
// across runs; the keys and their meaning are the game's.
type Prefs struct {
	Dir string // the folder holding PrefsName
}

// Path is the settings file.
func (p Prefs) Path() string { return filepath.Join(p.Dir, PrefsName) }

// Load fills each of dst (pointers) from the file; keys the file lacks
// keep dst's values, the defaults. No file: false, dst untouched.
func (p Prefs) Load(dst ...any) bool {
	b, err := os.ReadFile(p.Path())
	if err != nil {
		return false
	}
	for _, d := range dst {
		json.Unmarshal(b, d)
	}
	return true
}

// Save writes v as indented JSON, atomically: a temporary file beside the
// old one replaces it only once written whole.
func (p Prefs) Save(v any) error {
	b, err := json.MarshalIndent(v, "", " ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(p.Dir, 0o755); err != nil {
		return err
	}
	f, err := os.CreateTemp(p.Dir, "."+PrefsName+"-*")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(f.Name(), 0o644)
	}
	if err == nil {
		err = os.Rename(f.Name(), p.Path())
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}
