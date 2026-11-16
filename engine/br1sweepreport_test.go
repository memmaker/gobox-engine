package engine

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"gobox/ecl"
)

// The sweep's report: per block the instructions run of the total, and the positions never reached with the reason
// the sweep can tell. A region is a stretch of never-run instructions between two jump targets; its reason is read from
// what leads into it (the jumps to it, the instruction before it): a condition (a quest word named when the condition
// compares one), a fight, a reply of a menu, nothing at all, or code that is itself never reached.

// sweepTerminal is whether an instruction never falls through to the next.
func sweepTerminal(op byte) bool {
	switch op {
	case ecl.OpExit, ecl.OpGoto, ecl.OpReturn, ecl.OpNewECL, ecl.OpClearExit:
		return true
	}
	return false
}

func sweepIsCondition(op byte) bool {
	switch op {
	case ecl.OpIfEq, ecl.OpIfNe, ecl.OpIfLt, ecl.OpIfGt, ecl.OpIfLe, ecl.OpIfGe:
		return true
	}
	return false
}

func sweepIsMenu(op byte) bool {
	switch op {
	case ecl.OpHMenu, ecl.OpVMenu, ecl.OpMenuKeys, ecl.OpYesNo, ecl.OpOnGoto, ecl.OpOnGosub:
		return true
	}
	return false
}

var sweepQuestWord = regexp.MustCompile(`\bWord[0-9A-F]{4}\b`)

// sweepRegion is a stretch of instructions never run.
type sweepRegion struct {
	Block, From, To int
	Reason, Class   string
}

// sweepRegions are the stretches of block's never-run instructions with the reason each is not reached.
func sweepRegions(t testing.TB, w *sweepWorld, s *ecl.Script, block int, lines []string) (out []sweepRegion) {
	hit := w.hit[block]
	target := map[int]bool{}
	from := map[int][]int{} // jump target -> the instructions that jump there
	for i, in := range s.Code {
		for _, a := range in.Args {
			if a.Label {
				target[a.N] = true
				from[a.N] = append(from[a.N], i)
			}
		}
	}
	for _, ev := range s.Events {
		target[ev] = true
	}
	lineText := func(i int) string {
		if n := s.Code[i].Line; n >= 1 && n <= len(lines) {
			return strings.TrimSpace(lines[n-1])
		}
		return ""
	}
	for i := 0; i < len(s.Code); {
		if hit[i] {
			i++
			continue
		}
		j := i + 1
		for j < len(s.Code) && !hit[j] && !target[j] {
			j++
		}
		r := sweepRegion{Block: block, From: i, To: j - 1}
		var sources []int
		sources = append(sources, from[i]...)
		if i > 0 && !sweepTerminal(s.Code[i-1].Op) {
			sources = append(sources, i-1)
		}
		isEntry := false
		for _, ev := range s.Events {
			isEntry = isEntry || ev == i
		}
		switch reached := []int{}; {
		case len(sources) == 0 && !isEntry:
			r.Class, r.Reason = "no jump leads here", "no instruction jumps or falls into it"
		default:
			for _, src := range sources {
				if hit[src] {
					reached = append(reached, src)
				}
			}
			if len(reached) == 0 {
				r.Class, r.Reason = "behind unreached code", "every way in starts at an instruction never run"
				break
			}
			src := reached[0]
			op, text := s.Code[src].Op, lineText(src)
			switch {
			case sweepIsCondition(op):
				if wd := sweepQuestWord.FindString(text); wd != "" {
					r.Class, r.Reason = "behind a quest word", "if "+text+" (word "+wd+")"
				} else {
					r.Class, r.Reason = "behind a condition", text
				}
			case op == ecl.OpCombat:
				r.Class, r.Reason = "behind a fight not won", text
			case sweepIsMenu(op):
				r.Class, r.Reason = "behind a reply not taken", text
			default:
				r.Class, r.Reason = "after "+sweepOpName(op), text
			}
		}
		out = append(out, r)
		i = j
	}
	return out
}

func sweepOpName(op byte) string { return strconv.Itoa(int(op)) }

// sweepReport logs the coverage table, the unreached regions by class and the fights.
func sweepReport(t *testing.T, w *sweepWorld, elapsed time.Duration) {
	e := br1EngineIn(t, w.adv)
	run, total := 0, 0
	t.Logf("sweep: %v, %d runs, %d dropped from the work list, wedged walks by block %v", elapsed.Round(time.Second), sumRuns(w), w.dropped, w.wedged)
	t.Logf("%-6s %8s %8s %5s %6s %6s", "block", "run", "total", "%", "runs", "left")
	classes := map[string]int{}
	byClass := map[string][]sweepRegion{}
	for _, block := range br1Blocks {
		s, err := e.parseScriptFile(block)
		if err != nil {
			t.Fatal(block, err)
		}
		b, err := os.ReadFile(filepath.Join(w.adv, "scripts", "area"+strconv.Itoa(block)+".txt"))
		if err != nil {
			t.Fatal(err)
		}
		n := len(w.hit[block])
		run, total = run+n, total+len(s.Code)
		t.Logf("%-6d %8d %8d %4d%% %6d %6d", block, n, len(s.Code), 100*n/max(len(s.Code), 1), w.runs[block], len(w.queue[block]))
		for _, r := range sweepRegions(t, w, s, block, strings.Split(string(b), "\n")) {
			classes[r.Class] += r.To - r.From + 1
			byClass[r.Class] = append(byClass[r.Class], r)
		}
	}
	t.Logf("%-6s %8d %8d %4d%%", "all", run, total, 100*run/max(total, 1))
	names := make([]string, 0, len(classes))
	for c := range classes {
		names = append(names, c)
	}
	sort.Slice(names, func(i, j int) bool { return classes[names[i]] > classes[names[j]] })
	for _, c := range names {
		t.Logf("unreached, %s: %d instructions in %d regions", c, classes[c], len(byClass[c]))
		rs := byClass[c]
		sort.Slice(rs, func(i, j int) bool { return rs[i].To-rs[i].From > rs[j].To-rs[j].From })
		for _, r := range rs[:min(len(rs), 6)] {
			t.Logf("    block %d instr %d..%d: %s", r.Block, r.From, r.To, r.Reason)
		}
	}
	lostN := 0
	for pos, n := range w.lost {
		lostN += n
		t.Logf("lost fight at %v: %d times, won %d", pos, n, w.won[pos])
	}
	never := 0
	for pos, n := range w.offered {
		for i := range n {
			if !w.taken[pos][i] {
				never++
			}
		}
	}
	t.Logf("fights won at %d places, lost at %d; replies offered and never taken: %d; states met: %v", len(w.won), len(w.lost), never, w.seen)
}

func sumRuns(w *sweepWorld) (n int) {
	for _, r := range w.runs {
		n += r
	}
	return n
}
