package pool_test

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

// TestMain runs the import a test's areas need (por-go's cmd/importstale),
// as the setup does: nothing at load decompiles. It runs the game's own
// command, so the engine's module does not import the game.
func TestMain(m *testing.M) {
	cmd := exec.Command("go", "run", "./cmd/importstale")
	cmd.Dir = "../../por-go"
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(m.Run())
}
