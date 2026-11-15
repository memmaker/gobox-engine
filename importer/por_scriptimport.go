package importer

import (
	"bytes"
	"errors"
	"gobox/asset/orig"
	"gobox/ecl"
	"io/fs"
	"os"
	"path/filepath"
)

// ScriptData is the checked-in script names and layout, the decompiler's
// input (PorAssets' scripts/, its README.md), the same wherever the import runs.
func ScriptData() fs.FS {
	sub, err := fs.Sub(PorAssets, "por/scripts")
	if err != nil {
		panic(err) // a constant path in the embedded tree
	}
	return sub
}

// writeScriptData writes ScriptData into the scripts folder dir, over what is there.
func writeScriptData(dir string) error {
	data := ScriptData()
	return fs.WalkDir(data, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := fs.ReadFile(data, p)
		if err != nil {
			return err
		}
		dst := filepath.Join(dir, filepath.FromSlash(p))
		if have, err := os.ReadFile(dst); err == nil && bytes.Equal(have, b) {
			return nil // as it is: not rewritten, so a reader beside this one (another test binary) never sees it cut short
		}
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o644)
	})
}

// ImportScripts decompiles the game's event scripts (ECL*.DAX in dos) into
// dir, the scripts folder with its names/ and layout/, with POR's special
// cases (scriptFixes). It is the import path: nothing at run time
// decompiles (ecl.Library only reads the texts).
func ImportScripts(dos, dir string) error {
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	_, err = ecl.Import(ecl.V1, ImportSource(l), os.DirFS(dir), dir, ecl.FoldGates, porScriptFixes)
	return err
}

// ImportIfStale runs ImportScripts when dir's scripts are missing or stale:
// the setup's step, and a test binary's before it loads an area. Without
// game files there is nothing to import (the tests skip).
func ImportIfStale(dos, dir string) error {
	if _, err := os.Stat(dos); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err := writeScriptData(dir); err != nil {
		return err
	}
	l, err := orig.NewLoader(dos)
	if err != nil {
		return err
	}
	return ecl.ImportIfStale(ecl.V1, ImportSource(l), os.DirFS(dir), dir, ecl.FoldGates, porScriptFixes)
}
