package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
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
	case ecl.OpHMenu, ecl.OpVMenu, ecl.OpMenuKeys, ecl.OpYesNo:
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
	var srcs [][]int // per region: the instructions that lead into it
	regionOf := map[int]int{}
	for i := 0; i < len(s.Code); {
		if hit[i] {
			i++
			continue
		}
		j := i + 1
		for j < len(s.Code) && !hit[j] && !target[j] {
			j++
		}
		var sources []int
		sources = append(sources, from[i]...)
		if i > 0 && !sweepTerminal(s.Code[i-1].Op) {
			sources = append(sources, i-1)
		}
		if i > 1 && sweepIsCondition(s.Code[i-2].Op) { // an If skips the command after it when it does not hold: the one after that is where it goes on
			sources = append(sources, i-2)
		}
		for _, ev := range s.Events {
			if ev == i {
				sources = append(sources, -1) // an entry point
			}
		}
		for k := i; k < j; k++ {
			regionOf[k] = len(out)
		}
		out = append(out, sweepRegion{Block: block, From: i, To: j - 1})
		srcs = append(srcs, sources)
		i = j
	}
	done := make([]bool, len(out))
	var resolve func(n int, busy map[int]bool)
	resolve = func(n int, busy map[int]bool) {
		if done[n] {
			return
		}
		busy[n] = true
		r := &out[n]
		sources := srcs[n]
		var reached []int
		for _, src := range sources {
			if src == -1 || hit[src] {
				reached = append(reached, src)
			}
		}
		switch {
		case len(sources) == 0:
			r.Class, r.Reason = "no jump leads here", "no instruction jumps or falls into it"
		case len(reached) == 0: // every way in starts in another region: it is behind that one's reason
			root := -1
			for _, src := range sources {
				if m := regionOf[src]; !busy[m] {
					root = m
					break
				}
			}
			if root < 0 {
				r.Class, r.Reason = "behind unreached code", "its ways in start in itself (a loop never entered)"
				break
			}
			resolve(root, busy)
			r.Class, r.Reason = out[root].Class, out[root].Reason+fmt.Sprintf(" [nested, via instr %d]", out[root].From)
		case reached[0] == -1:
			r.Class, r.Reason = "entry point never run", "an entry of the script that no run called"
		default:
			src := reached[0]
			op, text := s.Code[src].Op, lineText(src)
			switch {
			case sweepIsCondition(op):
				if wd := sweepQuestWord.FindString(text); wd != "" {
					r.Class, r.Reason = "behind a quest word", text+" ("+sweepWordNote(t, w, wd, text)+")"
				} else {
					r.Class, r.Reason = "behind a condition", text
				}
			case op == ecl.OpCombat:
				r.Class, r.Reason = "behind a fight not won", text
			case sweepIsMenu(op):
				r.Class, r.Reason = "behind a reply not taken", text
			case op == ecl.OpOnGoto || op == ecl.OpOnGosub: // goto [labels][index]: the index never took the value that leads here
				r.Class, r.Reason = "behind a computed jump", text
			default:
				r.Class, r.Reason = "after a command that goes on elsewhere", text
			}
		}
		done[n] = true
	}
	for n := range out {
		resolve(n, map[int]bool{})
	}
	return out
}

// sweepReport logs the coverage table, the unreached regions by class and the fights.
func sweepReport(t *testing.T, w *sweepWorld, elapsed time.Duration) {
	e := br1EngineIn(t, w.adv)
	run, total := 0, 0
	t.Logf("sweep: %v, %d runs, %d dropped from the work list, wedged walks by block %v", elapsed.Round(time.Second), sumRuns(w), w.dropped, w.wedged)
	t.Logf("%-6s %8s %8s %5s %6s %6s %6s", "block", "run", "total", "%", "runs", "left", "secs")
	classes := map[string]int{}
	byClass := map[string][]sweepRegion{}
	texts := map[int][]string{} // the scripts' lines, by block
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
		t.Logf("%-6d %8d %8d %4d%% %6d %6d %6.0f", block, n, len(s.Code), 100*n/max(len(s.Code), 1), w.runs[block], len(w.queue[block]), w.took[block].Seconds())
		texts[block] = strings.Split(string(b), "\n")
		for _, r := range sweepRegions(t, w, s, block, texts[block]) {
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
		top := 6 // the heavy sweep lists the most
		if w.heavy {
			top = 60
		}
		for _, r := range rs[:min(len(rs), top)] {
			t.Logf("    block %d instr %d..%d: %s", r.Block, r.From, r.To, r.Reason)
		}
	}
	lostN := 0
	for pos, n := range w.lost {
		lostN += n
		t.Logf("lost fight at %v: %d times, won %d", pos, n, w.won[pos])
	}
	overs := make([]sweepPos, 0, len(w.over))
	for pos := range w.over {
		overs = append(overs, pos)
	}
	sort.Slice(overs, func(i, j int) bool { return w.over[overs[i]] > w.over[overs[j]] })
	for _, pos := range overs[:min(len(overs), 12)] {
		line := ""
		if code, ok := sweepCodeAt(e, pos); ok && code.Line >= 1 && code.Line <= len(texts[pos.Block]) {
			line = strings.TrimSpace(texts[pos.Block][code.Line-1])
		}
		t.Logf("game over (the party beaten, or a script ended the game) after script place %v: %d runs: %s", pos, w.over[pos], line)
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

// sweepWordNote says where the story sets the word a condition compares: the constants it is compared with, and for each
// the blocks whose scripts set it to that constant ("never set to it by any script": only the engine or a computation can).
func sweepWordNote(t testing.TB, w *sweepWorld, word, text string) string {
	st := sweepStoryOf(t, w.adv)
	var parts []string
	seen := map[int64]bool{}
	note := func(v int64) {
		if seen[v] {
			return
		}
		seen[v] = true
		var by []int
		for _, b := range br1Blocks {
			if slices.Contains(st.writes[b][word], v) {
				by = append(by, b)
			}
		}
		if len(by) == 0 {
			parts = append(parts, fmt.Sprintf("%s vs %d: no script sets it to %d", word, v, v))
		} else {
			parts = append(parts, fmt.Sprintf("%s vs %d: set to it by blocks %v", word, v, by))
		}
	}
	for _, m := range sweepCmpLeft.FindAllStringSubmatch(text, -1) {
		if m[1] == word {
			v, _ := strconv.ParseInt(m[3], 10, 64)
			note(v)
		}
	}
	for _, m := range sweepCmpRight.FindAllStringSubmatch(text, -1) {
		if m[3] == word {
			v, _ := strconv.ParseInt(m[1], 10, 64)
			note(v)
		}
	}
	if len(parts) == 0 {
		return "word " + word
	}
	return strings.Join(parts, "; ")
}

// sweepCodeAt is the instruction at a script place.
func sweepCodeAt(e *Engine, pos sweepPos) (ecl.Instr, bool) {
	s, err := e.parseScriptFile(pos.Block)
	if err != nil || pos.Instr >= len(s.Code) {
		return ecl.Instr{}, false
	}
	return s.Code[pos.Instr], true
}
