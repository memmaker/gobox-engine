package importpor

import (
	"errors"
	"gobox/dax"
	"gobox/ecl/script"
	"io/fs"
	"os"
	"path/filepath"
	"gobox/games/por"
	"gobox/games/por/importsrc"
)

// ScriptData is the checked-in script names and layout, the decompiler's
// input (por.Assets' scripts/, its README.md), the same wherever the import runs.
func ScriptData() fs.FS {
	sub, err := fs.Sub(por.Assets, "scripts")
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
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dst, b, 0o644)
	})
}

// ImportScripts decompiles the game's event scripts (ECL*.DAX in dos) into
// dir, the scripts folder with its names/ and layout/, with POR's special
// cases (scriptFixes). It is the import path: nothing at run time
// decompiles (script.Library only reads the texts).
func ImportScripts(dos, dir string) error {
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	_, err = script.Import(script.V1, importsrc.ImportSource(l), os.DirFS(dir), dir, script.FoldGates, scriptFixes)
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
	l, err := dax.NewLoader(dos)
	if err != nil {
		return err
	}
	return script.ImportIfStale(script.V1, importsrc.ImportSource(l), os.DirFS(dir), dir, script.FoldGates, scriptFixes)
}
