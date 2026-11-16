package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The sweep's entry states and guided replies (see br1sweep_test.go). Entering a block directly leaves the story
// words and ScriptBlock at their defaults, so the branches a block's load entry takes by them never ran. The states
// here are derived from the load entry's own text: every ScriptBlock value and quest word value it compares with,
// one state each (and one with every quest word set), so each distinct branch is entered; nothing is told by hand.

const (
	sweepParty       = 4  // roster characters added through the party menu before Begin (the party menu's limit): a stronger party outlives more fights
	sweepStatesMax   = 8  // states a block is entered in (the default sweep)
	sweepHeavyStates = 40 // and in the heavy sweep
)

// sweepCell is a layout cell a state sets before the block is entered.
type sweepCell struct {
	Name  string
	Value int64
}

var (
	sweepCmpLeft  = regexp.MustCompile(`\b(ScriptBlock|Word[0-9A-F]{4})\s*(==|!=|>=|<=|<|>)\s*(\d+)\b`)
	sweepCmpRight = regexp.MustCompile(`\b(\d+)\s*(==|!=|>=|<=|<|>)\s*(ScriptBlock|Word[0-9A-F]{4})\b`)
	sweepMasked   = regexp.MustCompile(`\b(Word[0-9A-F]{4})\s*&\s*(\d+)\b`)
	sweepWrite    = regexp.MustCompile(`\b(ScriptBlock|Word[0-9A-F]{4})\s*=\s*(\d+)\b`) // a constant written: ` == ` and the others do not match
	sweepNewECL   = regexp.MustCompile(`\bNewECL\((\d+)\)`)
)

// sweepStory is what the scripts' text says about the story: per block the persistent words it compares with a constant (in any
// entry or routine) and the words it sets to a constant, and the blocks that call NewECL for it. Quest words are the words
// the scripts keep across blocks (Word4C80 ...: the original's game record); ScriptBlock is the block left.
type sweepStory struct {
	tests  map[int]map[string][]int64
	writes map[int]map[string][]int64
	preds  map[int][]int // block -> the blocks whose scripts call NewECL for it
}

var sweepStories = map[string]*sweepStory{}

// sweepStoryOf reads the story of the blocks of the folder adv (cached by folder; blocks nil: it must be read already).
func sweepStoryOf(t testing.TB, adv string, blocks []int) *sweepStory {
	t.Helper()
	if st := sweepStories[adv]; st != nil {
		return st
	}
	if blocks == nil {
		t.Fatal("sweep: the story of " + adv + " is not read")
	}
	st := &sweepStory{tests: map[int]map[string][]int64{}, writes: map[int]map[string][]int64{}, preds: map[int][]int{}}
	add := func(m map[int]map[string][]int64, block int, name string, v int64) {
		if m[block] == nil {
			m[block] = map[string][]int64{}
		}
		if !slices.Contains(m[block][name], v) {
			m[block][name] = append(m[block][name], v)
		}
	}
	for _, block := range blocks {
		b, err := os.ReadFile(filepath.Join(adv, "scripts", "area"+strconv.Itoa(block)+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(b), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "#") {
				continue
			}
			for _, m := range sweepCmpLeft.FindAllStringSubmatch(line, -1) {
				v, _ := strconv.ParseInt(m[3], 10, 64)
				add(st.tests, block, m[1], v)
			}
			for _, m := range sweepCmpRight.FindAllStringSubmatch(line, -1) {
				v, _ := strconv.ParseInt(m[1], 10, 64)
				add(st.tests, block, m[3], v)
			}
			for _, m := range sweepMasked.FindAllStringSubmatch(line, -1) {
				v, _ := strconv.ParseInt(m[2], 10, 64)
				add(st.tests, block, m[1], v)
			}
			for _, m := range sweepWrite.FindAllStringSubmatch(line, -1) {
				v, _ := strconv.ParseInt(m[2], 10, 64)
				add(st.writes, block, m[1], v)
			}
			for _, m := range sweepNewECL.FindAllStringSubmatch(line, -1) {
				to, _ := strconv.Atoi(m[1])
				if !slices.Contains(st.preds[to], block) {
					st.preds[to] = append(st.preds[to], block)
				}
			}
		}
	}
	sweepStories[adv] = st
	return st
}

// sweepStatesOf derives the entry states of a block, the first the default (no cell set), at most max:
//   - as after each block that calls NewECL for it: ScriptBlock that block, and the words that block sets to a constant and
//     this one compares with, at those constants (the story as that block leaves it);
//   - each ScriptBlock value the block compares with, and next to it;
//   - each persistent word at each constant the block compares with (anywhere in it, every entry and routine);
//   - every word at its largest constant: the story told as far as the block asks.
func sweepStatesOf(t testing.TB, adv string, block, max int) [][]sweepCell {
	t.Helper()
	st := sweepStoryOf(t, adv, nil)
	chain := [][]sweepCell{}
	tested := st.tests[block]
	for _, from := range st.preds[block] {
		cells := []sweepCell{{"ScriptBlock", int64(from)}}
		names := make([]string, 0, len(st.writes[from]))
		for n := range st.writes[from] {
			names = append(names, n)
		}
		sort.Strings(names)
		for _, n := range names {
			if n == "ScriptBlock" || len(tested[n]) == 0 {
				continue
			}
			vs := slices.Clone(st.writes[from][n])
			slices.Sort(vs)
			cells = append(cells, sweepCell{n, vs[len(vs)-1]}) // the last value the story gives it there
		}
		chain = append(chain, cells)
	}
	states := [][]sweepCell{}
	var blocks []int64 // ScriptBlock values, as compared and next to them
	for _, v := range tested["ScriptBlock"] {
		for _, d := range []int64{-1, 0, 1} {
			if w := v + d; w > 0 && w < 256 && !slices.Contains(blocks, w) {
				blocks = append(blocks, w)
			}
		}
	}
	slices.Sort(blocks)
	for _, v := range blocks {
		states = append(states, []sweepCell{{"ScriptBlock", v}})
	}
	names := make([]string, 0, len(tested))
	for n := range tested {
		if n != "ScriptBlock" {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	var story []sweepCell
	for _, n := range names {
		vs := slices.Clone(tested[n])
		slices.Sort(vs)
		for _, v := range vs {
			if v != 0 { // 0 is the default state
				states = append(states, []sweepCell{{n, v}})
			}
		}
		story = append(story, sweepCell{n, vs[len(vs)-1]})
	}
	if len(story) > 1 {
		states = append(states, story) // the story told as far as the block asks
	}
	chain = chain[:min(len(chain), max/2)] // the story's own entries first, then the rest thinned evenly
	room := max - 1 - len(chain)
	if len(states) > room {
		thin := make([][]sweepCell, room)
		for i := range thin {
			thin[i] = states[i*len(states)/room]
		}
		states = thin
	}
	return append(append([][]sweepCell{nil}, chain...), states...)
}

// sweepApply sets a state's cells on the engine (before the block is entered).
func sweepApply(t testing.TB, e *Engine, cells []sweepCell) {
	t.Helper()
	for _, c := range cells {
		r, ok := e.layoutField(c.Name)
		if !ok {
			t.Fatalf("no field %s in the layout", c.Name)
		}
		e.host.Set(r, c.Value)
	}
}
