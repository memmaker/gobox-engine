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
	sweepParty     = 4 // roster characters added through the party menu before Begin (the party menu's limit): a stronger party outlives more fights
	sweepStatesMax = 8 // states a block is entered in
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
)

// sweepLoadEntry is the text of a script's load entry.
func sweepLoadEntry(text string) string {
	_, rest, ok := strings.Cut(text, "on load {")
	if !ok {
		return ""
	}
	if i := strings.Index(rest, "\non "); i >= 0 {
		rest = rest[:i]
	}
	return rest
}

// sweepStatesOf derives the entry states of a block: the first is the default (no cell set).
func sweepStatesOf(t testing.TB, adv string, block int) [][]sweepCell {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(adv, "scripts", "area"+strconv.Itoa(block)+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	var blocks []int64 // ScriptBlock values, as compared and next to them
	words := map[string][]int64{}
	add := func(name string, v int64) {
		if name == "ScriptBlock" {
			for _, d := range []int64{-1, 0, 1} {
				if w := v + d; w > 0 && w < 256 && !slices.Contains(blocks, w) {
					blocks = append(blocks, w)
				}
			}
		} else if v != 0 && !slices.Contains(words[name], v) { // 0 is the default state
			words[name] = append(words[name], v)
		}
	}
	for _, line := range strings.Split(sweepLoadEntry(string(b)), "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "if ") {
			continue
		}
		for _, m := range sweepCmpLeft.FindAllStringSubmatch(line, -1) {
			v, _ := strconv.ParseInt(m[3], 10, 64)
			add(m[1], v)
		}
		for _, m := range sweepCmpRight.FindAllStringSubmatch(line, -1) {
			v, _ := strconv.ParseInt(m[1], 10, 64)
			add(m[3], v)
		}
		for _, m := range sweepMasked.FindAllStringSubmatch(line, -1) {
			v, _ := strconv.ParseInt(m[2], 10, 64)
			add(m[1], v)
		}
	}
	slices.Sort(blocks)
	names := make([]string, 0, len(words))
	for n := range words {
		names = append(names, n)
	}
	sort.Strings(names)
	states := [][]sweepCell{nil}
	for _, v := range blocks {
		states = append(states, []sweepCell{{"ScriptBlock", v}})
	}
	var story []sweepCell
	for _, n := range names {
		slices.Sort(words[n])
		for _, v := range words[n] {
			states = append(states, []sweepCell{{n, v}})
		}
		story = append(story, sweepCell{n, words[n][len(words[n])-1]})
	}
	if len(story) > 1 {
		states = append(states, story) // the story told as far as the entry asks
	}
	if len(states) > sweepStatesMax {
		thin := make([][]sweepCell, sweepStatesMax)
		for i := range thin {
			thin[i] = states[i*len(states)/sweepStatesMax]
		}
		states = thin
	}
	return states
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

// sweepCoverageTotal logs the distinct instructions run over all blocks' (the table's last row).
func sweepCoverageTotal(t testing.TB, e *Engine, hit map[int]map[int]bool) {
	run, total := 0, 0
	for _, block := range br1Blocks {
		s, err := e.parseScriptFile(block)
		if err != nil {
			t.Fatal(block, err)
		}
		run, total = run+len(hit[block]), total+len(s.Code)
	}
	t.Logf("%-6s %8d %8d  (%d%%)", "all", run, total, 100*run/total)
}
