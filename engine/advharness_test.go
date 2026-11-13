package engine

import (
	"os"
	"slices"
	"testing"

	"gobox/adventure"
	"gobox/game"
	"gobox/importer"
	"gobox/rules1e"
)

// The adventure harness: any game's imported folder (gobox import <game>
// <originals>) started headless, as quickStarted does for POR. Each game has
// its own env var holding the folder; a test skips when it is unset.
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
	f, err := importer.Read(adv)
	if err != nil {
		t.Fatal(err)
	}
	tables, err := adventure.ReadGameTables(adv)
	if err != nil {
		t.Fatal(err)
	}
	e, err := New(rules1e.Rules1e, tables, adv, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := e.SetStart(f.Manifest.Start); err != nil { // the adventure's Start sequence
		t.Fatal(err)
	}
	if err := e.SetEnd(f.Manifest.End); err != nil { // and its End, as playpool hands both
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

// quitEnds: Quit at the party menu ends the run with game.Ended (the engine
// never ends the process).
func quitEnds(t *testing.T, e *Engine) {
	t.Helper()
	out, err := e.doNow(game.Quit{})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.ContainsFunc(out.Events, func(ev game.Event) bool { _, ok := ev.(game.Ended); return ok }) {
		t.Fatalf("Quit gave %v, want game.Ended", out.Events)
	}
}

// smoke: starts, reaches the first input wait (the party menu, an Assembling
// state), and quits.
func smoke(t *testing.T, env string) {
	e := advStarted(t, env)
	if _, ok := e.build().(game.Assembling); !ok {
		t.Fatalf("first wait is %T, want the party menu (game.Assembling)", e.build())
	}
	quitEnds(t, e)
}

func TestAdvSmokeSOTSB(t *testing.T) { smoke(t, advEnvSOTSB) }
func TestAdvSmokePOD(t *testing.T)   { smoke(t, advEnvPOD) }
func TestAdvSmokeCOTAB(t *testing.T) { smoke(t, advEnvCOTAB) }

// TestAdvSmokePOR: the same on POR's game files (it skips without them).
func TestAdvSmokePOR(t *testing.T) {
	e := testEngine(t)
	e.Open("")
	e.titleDone()
	e.Advance()
	if _, ok := e.build().(game.Assembling); !ok {
		t.Fatalf("first wait is %T, want the party menu (game.Assembling)", e.build())
	}
	quitEnds(t, e)
}

// TestAdvBeginSOTSB: Quick Start and Begin take sotsb's party into the
// adventure (the first wait after the party menu is not the party menu and
// no error), as sotsb's own journey tests started (rules/journey_test.go).
func TestAdvBeginSOTSB(t *testing.T) { advBegin(t, advEnvSOTSB) }

// TestAdvBeginPOD: the same for pod (its own engine's journey tests, pod rules/, are gone).
func TestAdvBeginPOD(t *testing.T) { advBegin(t, advEnvPOD) }

// TestAdvBeginCOTAB: the same for cotab (its own engine, cotab rules/, is gone).
func TestAdvBeginCOTAB(t *testing.T) { advBegin(t, advEnvCOTAB) }

// advBegin: Quick Start and Begin leave the party menu without an error.
func advBegin(t *testing.T, env string) {
	e := advStarted(t, env)
	for _, a := range []game.Action{game.QuickStart{}, game.Begin{}} {
		if _, err := e.doNow(a); err != nil {
			t.Fatalf("%T: %v", a, err)
		}
	}
	if _, ok := e.build().(game.Assembling); ok {
		t.Fatalf("after Begin still at the party menu")
	}
	t.Logf("first wait after Begin: %T", e.build())
}
