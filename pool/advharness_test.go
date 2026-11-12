package pool

import (
	"os"
	"testing"

	"gobox/char"
	"gobox/game/play"
)

// The adventure harness: any game's imported folder (gobox import <game>
// <originals>) started headless, as quickStarted does for POR. Each game has
// its own env var holding the folder; a test skips when it is unset.
// Until pool runs these games the smoke tests fail when their var is set
// (sotsb, cotab: script: no game field QuestSlums; pod: LoadSymbolSet 202 exits).
const (
	advEnvSOTSB = "POOL_ADV_SOTSB"
	advEnvPOD   = "POOL_ADV_POD"
	advEnvCOTAB = "POOL_ADV_COTAB"
)

// advEngine is an engine on the imported folder named by env: the game's
// ruleset tables (no title steps: the Start sequence is the presenter's), its saves under a temp dir, a fixed seed.
func advEngine(t *testing.T, env string) *Engine {
	t.Helper()
	adv := os.Getenv(env)
	if adv == "" {
		t.Skip(env + " unset: an imported adventure folder")
	}
	tables, err := ReadGameTables(adv)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(char.Rules1e, tables, adv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	seed := uint32(7)
	e.FixedSeed = &seed
	return e
}

// advStarted runs the engine to its first input wait: the title done, the
// party menu up (scripted input then goes through Do).
func advStarted(t *testing.T, env string) *Engine {
	t.Helper()
	e := advEngine(t, env)
	e.Open("")
	e.titleDone()
	e.Advance()
	return e
}

// smoke: starts, reaches the first input wait (the party menu, an Assembling
// state), and stops there (Quit would os.Exit the test binary; dropping the engine is the quit).
func smoke(t *testing.T, env string) {
	e := advStarted(t, env)
	if _, ok := e.State().(play.Assembling); !ok {
		t.Fatalf("first wait is %T, want the party menu (play.Assembling)", e.State())
	}
}

func TestAdvSmokeSOTSB(t *testing.T) { smoke(t, advEnvSOTSB) }
func TestAdvSmokePOD(t *testing.T)   { smoke(t, advEnvPOD) }
func TestAdvSmokeCOTAB(t *testing.T) { smoke(t, advEnvCOTAB) }
