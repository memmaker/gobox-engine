package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
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
	// Seen once, gone after the EndFight fix moved the walks: block 50, "a script ran N instructions on one action"
	// (opOnGotoGosub in a step entry; replay 50 0 0 6 12138934 at commit "rulesxxvc: a move's reach"). Unverified, unexplained.
	{`^LOG: script: area \d+: T_\w+\[\d+\] is past the field's \d+ entries`, "none: the entry skips the story, so the index variable (Word4C61 ...) holds 0 and the table is read past its end as the original does (engine log, TableFaults)"},
}

const (
	sweepRunsPerBlock = 60                // runs a block gets by default
	sweepDeadline     = 100 * time.Second // the default run's time
	sweepHeavyRuns    = 100000
	sweepHeavyTime    = 30 * time.Minute
)

// sweepDigits makes findings that differ in a number only one group.
var sweepDigits = regexp.MustCompile(`\d+`)

// sweepKnownMatch is whether a finding's key is listed.
func sweepKnownMatch(k string) bool {
	return slices.ContainsFunc(br1SweepKnown, func(kn struct{ Pattern, Why string }) bool { return regexp.MustCompile(kn.Pattern).MatchString(k) })
}

func TestBR1Sweep(t *testing.T) {
	t.Setenv("GOBOX_NOSOUND", "1")
	dos := filepath.Join("..", "data", "br1", "sources", "dos")
	if _, err := os.Stat(filepath.Join(dos, "START.EXE")); err != nil {
		t.Skip("no BR1 originals:", err)
	}
	adv := t.TempDir()
	if err := importer.ImportBREngine("br1", importer.BR1EXE, dos, adv); err != nil {
		t.Fatal(err)
	}
	perBlock, limit := sweepRunsPerBlock, sweepDeadline
	if os.Getenv("GOBOX_SWEEP") == "heavy" {
		perBlock, limit = sweepHeavyRuns, sweepHeavyTime
	}
	var logged strings.Builder // what the engine logs (a script that did not resolve a field...) is a finding too
	log.SetFlags(0)
	log.SetOutput(&logged)
	defer log.SetOutput(os.Stderr)
	w := newSweepWorld(adv, os.Getenv("GOBOX_SWEEP") == "heavy")
	start := time.Now()
	for _, block := range br1Blocks {
		for s := range sweepStatesOf(t, adv, block) {
			w.push(sweepItem{Entry: sweepEntry{Block: block, State: s}})
		}
	}
	for progress := true; progress && time.Since(start) < limit; {
		progress = false
		for _, block := range br1Blocks {
			if w.runs[block] >= perBlock || len(w.queue[block]) == 0 {
				continue
			}
			it := w.queue[block][0]
			w.queue[block] = w.queue[block][1:]
			w.runs[block]++
			if f := w.run(t, it); f != nil {
				w.add(f)
			}
			progress = true
		}
	}
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
		if !sweepKnownMatch(k) {
			t.Errorf("new finding (x%d): %s", len(w.groups[k]), k)
		}
	}
	for _, kn := range br1SweepKnown {
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
	if err := importer.ImportBREngine("br1", importer.BR1EXE, filepath.Join("..", "data", "br1", "sources", "dos"), adv); err != nil {
		t.Fatal(err)
	}
	w := newSweepWorld(adv, true)
	it := sweepItem{Entry: en, Prefix: tk}
	if f := w.run(t, it); f != nil {
		t.Fatalf("%s\ntail %v\n%s", f.Key, f.Tail, f.Frame)
	}
	for k := range w.groups {
		t.Fatalf("%s", k)
	}
}
