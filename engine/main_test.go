package engine_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// TestMain needs the imports a test's areas need (por-go's cmd/importstale,
// run by the gate before these tests): nothing at load decompiles, and the
// engine's module does not import the game. With the game files present and
// an import missing, it fails with a clear message; it cannot tell a stale
// import (por's the por import decides that), so the gate runs the import first.
func TestMain(m *testing.M) {
	const por = "../data/por" // the imported adventure; its originals in sources/
	if _, err := os.Stat(filepath.Join(por, "sources", "dos", "TITLE.DAX")); err == nil {
		for _, f := range []string{"data/key.txt", "data/portraits.json", "scripts/area1.txt"} {
			if _, err := os.Stat(filepath.Join(por, f)); err != nil {
				fmt.Fprintf(os.Stderr, "missing %s: run por's import (go run ./cmd/importstale in games/por) before these tests\n", f)
				os.Exit(1)
			}
		}
	}
	os.Exit(m.Run())
}
