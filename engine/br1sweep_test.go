package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"gobox/importer"
)

// The BR1 sweep (br1sweeprun_test.go has the runs, br1sweepreport_test.go the report of what was not reached): every
// script block entered in each of its entry states (br1sweepstate_test.go), its map explored by a breadth-first walk
// over every square a step reaches (in all four facings where the script reads the facing), every question's every reply
// taken at least once from its place in the script (a work list of runs: the tokens that lead to the question and the
// reply), fights played to a win on Quick. A panic ends its run and is a finding; findings are grouped and compared with
// br1SweepKnown. A lost fight is noted, not a finding. GOBOX_SWEEP=heavy goes on until the work list is empty.

// br1SweepKnown are the findings still open: Pattern is matched against "message @ top frame".
// The test fails on a finding no pattern matches and on a pattern no finding matches.
var br1SweepKnown = []struct{ Pattern, Why string }{
	{`^sweep: a script ran # instructions on one action without asking \(block #\)`, "open, unexplained: block 50's step entry, instr 259..264 (area50.txt `goto [L_96B2, L_973A, L_978C, L_97CE][NewGameBytes]` with NewGameBytes = 63 & Cell past the table, so it falls through to `if 1 == Search`, `EntryBytes_20 -= 1`, `goto L_88C3` back to the table) loops for ever once the party stands on such a square after the first fight; no other exit is visible in the text. Replay: the finding's line"},
	{`^LOG: script: area \d+: T_\w+\[\d+\] is past the field's \d+ entries`, "none: the entry skips the story, so the index variable (Word4C61 ...) holds 0 and the table is read past its end as the original does (engine log, TableFaults)"},
}

const (
	sweepWorkCap   = 20000           // the work one block's runs may take by default: tokens and, for each fight, sweepFightWork more
	sweepHeavyWork = 1 << 30         // the heavy sweep's: the work list, or its time
	sweepHeavyTime = 8 * time.Minute // what a block gets in the heavy sweep
	sweepFightWork = 40
)

// sweepWorkers is how many blocks are swept at once (GOBOX_SWEEP_WORKERS; the cores, less two).
func sweepWorkers() int {
	if n, err := strconv.Atoi(os.Getenv("GOBOX_SWEEP_WORKERS")); err == nil && n > 0 {
		return n
	}
	return max(1, min(8, runtime.NumCPU()-2))
}

// sweepDigits makes findings that differ in a number only one group.
var sweepDigits = regexp.MustCompile(`\d+`)

// sweepKnownMatch is whether a finding's key is listed.
func sweepKnownMatch(known []struct{ Pattern, Why string }, k string) bool {
	return slices.ContainsFunc(known, func(kn struct{ Pattern, Why string }) bool { return regexp.MustCompile(kn.Pattern).MatchString(k) })
}

func TestBR1Sweep(t *testing.T) { sweepGameRun(t, gameBR1, br1Blocks, br1SweepKnown) }

// sweepGameRun sweeps the blocks of game g (its imported folder, its known findings).
func sweepGameRun(t *testing.T, g brGame, blockList []int, known []struct{ Pattern, Why string }) {
	t.Setenv("GOBOX_NOSOUND", "1")
	if _, err := os.Stat(filepath.Join(g.dos(), "START.EXE")); err != nil {
		t.Skip("no "+g.Name+" originals:", err)
	}
	adv := t.TempDir()
	if err := importer.ImportBR(g.ID, g.EXE, g.dos(), adv); err != nil {
		t.Fatal(err)
	}
	heavy := os.Getenv("GOBOX_SWEEP") == "heavy"
	defer debug.SetGCPercent(debug.SetGCPercent(400)) // the runs allocate a great deal that dies at once: fewer collections
	var logged strings.Builder                        // what the engine logs (a script that did not resolve a field...) is a finding too
	log.SetFlags(0)
	log.SetOutput(&logged)
	defer log.SetOutput(os.Stderr)
	sweepStoryOf(t, adv, blockList) // read once, before the blocks go in parallel
	// Each entry block is a world of its own (its work list, its replies taken, its map explored), run one run after the
	// other, so a block's runs are the same on every sweep; the blocks run side by side and the worlds are merged for the report.
	workers := sweepWorkers()
	sem := make(chan struct{}, workers)
	ws := make([]*sweepWorld, len(blockList))
	var wg sync.WaitGroup
	start := time.Now()
	for i, block := range blockList {
		ws[i] = newSweepWorld(adv, blockList, heavy)
		wg.Add(1)
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			ws[i].runBlock(t, block)
		}()
	}
	wg.Wait()
	w := sweepMerge(adv, blockList, heavy, ws)
	elapsed := time.Since(start)
	for k, n := range w.notPorted {
		for range n {
			w.add(&sweepFinding{Key: "REFUSED: " + k + " @ -"})
		}
	}
	for _, l := range strings.Split(logged.String(), "\n") {
		if l != "" {
			w.add(&sweepFinding{Key: "LOG: " + l + " @ -"})
		}
	}
	sweepReport(t, w, elapsed)

	keys := make([]string, 0, len(w.groups))
	for k := range w.groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(w.groups[keys[i]]) > len(w.groups[keys[j]]) || len(w.groups[keys[i]]) == len(w.groups[keys[j]]) && keys[i] < keys[j]
	})
	for _, k := range keys {
		g := w.groups[k]
		blocks := map[int]bool{}
		for _, f := range g {
			blocks[f.Item.Entry.Block] = true
		}
		first := g[0]
		t.Logf("x%d blocks %v: %s\n  replay: GOBOX_SWEEP_REPLAY=%q\n  tail: %v\n  frames:\n%s", len(g), sortedInts(blocks), k, sweepReplayString(first), first.Tail, first.Frame)
	}
	for _, k := range keys {
		if !sweepKnownMatch(known, k) {
			t.Errorf("new finding (x%d): %s", len(w.groups[k]), k)
		}
	}
	for _, kn := range known {
		if !slices.ContainsFunc(keys, func(k string) bool { return regexp.MustCompile(kn.Pattern).MatchString(k) }) {
			t.Errorf("listed finding no longer occurs: %s (%s)", kn.Pattern, kn.Why)
		}
	}
	if w.diverged > 0 {
		t.Errorf("%d replays diverged from their runs (the engine is not deterministic here)", w.diverged)
	}
}

func sortedInts(m map[int]bool) []int {
	var out []int
	for k := range m {
		out = append(out, k)
	}
	slices.Sort(out)
	return out
}

// sweepReplayString is a finding's run as GOBOX_SWEEP_REPLAY takes it: the entry, " : ", the tokens.
func sweepReplayString(f *sweepFinding) string {
	return f.Item.Entry.String() + " : " + sweepToksString(f.Toks)
}

// TestBR1SweepReplay replays one recorded run: GOBOX_SWEEP_REPLAY="block/state/x,y/facing : tokens" (a finding's replay line).
// It fails with the panic the run recorded, so a fix is proven by the same line passing.
func TestBR1SweepReplay(t *testing.T) {
	spec := os.Getenv("GOBOX_SWEEP_REPLAY")
	if spec == "" {
		t.Skip("GOBOX_SWEEP_REPLAY not set")
	}
	entry, toks, _ := strings.Cut(spec, " : ")
	var en sweepEntry
	if _, err := fmt.Sscanf(entry, "%d/%d/%d,%d/%d", &en.Block, &en.State, &en.X, &en.Y, &en.Facing); err != nil {
		t.Fatalf("entry %q: %v", entry, err)
	}
	tk, err := parseSweepToks(toks)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOBOX_NOSOUND", "1")
	adv := t.TempDir()
	g, blocks := gameBR1, br1Blocks
	if os.Getenv("GOBOX_SWEEP_GAME") == gameBR2.ID { // GOBOX_SWEEP_GAME=br2 replays a BR2 finding
		g, blocks = gameBR2, br2Blocks
	}
	if err := importer.ImportBR(g.ID, g.EXE, g.dos(), adv); err != nil {
		t.Fatal(err)
	}
	sweepStoryOf(t, adv, blocks)
	w := newSweepWorld(adv, blocks, true)
	it := sweepItem{Entry: en, Prefix: tk}
	if f := w.run(t, it); f != nil {
		t.Fatalf("%s\ntail %v\n%s", f.Key, f.Tail, f.Frame)
	}
	for k := range w.groups {
		t.Fatalf("%s", k)
	}
}
