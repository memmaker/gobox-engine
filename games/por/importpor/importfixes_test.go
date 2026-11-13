package importpor

import (
	"path/filepath"
	"testing"

	"gobox/asset/orig"
	"gobox/ecl/script"
	"gobox/games/por/importsrc"
)

// TestImportAppliesFixes: the game data decompiles and every special case (scriptFixes) applies.
func TestImportAppliesFixes(t *testing.T) {
	l, err := orig.NewLoader("../../../data/por/sources/dos")
	if err != nil {
		t.Skip(err)
	}
	dir := t.TempDir()
	if _, err := script.Import(script.V1, importsrc.ImportSource(l), ScriptData(), dir, script.FoldGates, scriptFixes); err != nil {
		t.Fatal(err)
	}
	if files, _ := filepath.Glob(filepath.Join(dir, "area*.txt")); len(files) == 0 {
		t.Fatal("no scripts")
	}
}
