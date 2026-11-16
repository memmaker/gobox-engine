package engine

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gobox/ecl"
	"gobox/game"
	"gobox/importer"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// The BR1 sweep: every script block entered at several squares, each walked by a bounded random sequence of the
// actions a front could send (chosen from what the state allows, by a seeded rules.RNG). A panic ends its walk and
// is a finding; findings are grouped and compared with br1SweepKnown. GOBOX_SWEEP=heavy runs ten times the walks.

// br1SweepKnown are the findings still open: Pattern is matched against "message @ top frame".
// The test fails on a finding no pattern matches and on a pattern no finding matches.
var br1SweepKnown = []struct{ Pattern, Why string }{
	// Seen once, gone after the EndFight fix moved the walks: block 50, "a script ran N instructions on one action"
	// (opOnGotoGosub in a step entry; replay 50 0 0 6 12138934 at commit "rulesxxvc: a move's reach"). Unverified, unexplained.
	{`^LOG: script: area \d+: T_\w+\[\d+\] is past the field's \d+ entries`, "none: the entry skips the story, so the index variable (Word4C61 ...) holds 0 and the table is read past its end as the original does (engine log, TableFaults)"},
}

const (
	sweepEntriesN          = 12  // entry squares per block (heavy: x6)
	sweepSteps             = 400 // actions per walk
	sweepInstructionBudget = 200000
	sweepSeed              = 0x5eed
)

// sweepSpec is one walk, enough to replay it: br1SweepWalk(t, adv, spec, nil).
type sweepSpec struct {
	Block, X, Y, Facing int
	Seed                uint32
	State               int // the entry state (br1sweepstate_test.go sweepStatesOf); 0: the defaults
}

type sweepFinding struct {
	Key     string // message @ top frame
	Spec    sweepSpec
	Actions []string
	Frames  string
	Tail    []string // the last script positions run (block:instruction), oldest first
}

// sweepWalk runs one walk on a new engine and returns its finding, nil if none. hit collects the executed
// instructions per block.
func br1SweepWalk(t testing.TB, adv string, sp sweepSpec, hit map[int]map[int]bool) (f *sweepFinding) {
	var actions []string
	var tail []string
	defer func() {
		if r := recover(); r != nil {
			f = &sweepFinding{Spec: sp, Actions: actions, Tail: tail}
			msg, top := fmt.Sprint(r), ""
			lines := strings.Split(string(debug.Stack()), "\n")
			after := false
			for i, l := range lines {
				if strings.HasPrefix(l, "panic(") {
					after = true
					continue
				}
				if !after || l == "" || strings.HasPrefix(l, "\t") || !strings.Contains(l, "gobox/") || strings.Contains(l, "br1SweepWalk") {
					continue
				}
				frame := l[:strings.LastIndex(l, "(")]
				if top == "" {
					top = frame
				}
				if len(f.Frames) < 600 && i+1 < len(lines) {
					f.Frames += strings.TrimSpace(frame) + " " + strings.TrimSpace(lines[i+1]) + "\n"
				}
			}
			f.Key = sweepDigits.ReplaceAllString(msg, "#") + " @ " + top
		}
	}()
	e := br1BeginWith(t, br1EngineIn(t, adv), sweepParty)
	ran := 0 // instructions since the last action: a script that never asks again is a finding, not a hang
	e.vm.Trace = func(vm *ecl.VM, n int) {
		if vm.Script == nil {
			return
		}
		if ran++; ran > sweepInstructionBudget {
			panic(fmt.Sprintf("sweep: a script ran %d instructions on one action without asking (block %d)", sweepInstructionBudget, vm.Script.Area))
		}
		if hit[vm.Script.Area] == nil {
			hit[vm.Script.Area] = map[int]bool{}
		}
		hit[vm.Script.Area][vm.Cur()] = true
		if tail = append(tail, fmt.Sprintf("%d:%d", vm.Script.Area, vm.Cur())); len(tail) > 40 {
			tail = tail[1:]
		}
	}
	rng := &rules.RNG{State: sp.Seed}
	e.push(cont{kind: kStepEnd})
	e.facing = sp.Facing
	sweepApply(t, e, sweepStatesOf(t, adv, sp.Block)[sp.State])
	e.DevAt(sp.Block, sp.X, sp.Y)
	for range sweepSteps {
		a := sweepChoose(e, rng)
		if a == nil {
			if s := e.build(); !sweepWaits(s) {
				return &sweepFinding{Key: fmt.Sprintf("STUCK: %T, no allowed action, no question @ -", s), Spec: sp, Actions: actions}
			}
			actions = append(actions, "Step")
			e.Step()
			continue
		}
		ran = 0
		actions = append(actions, a.label)
		if err := a.run(e); err != nil { // a refusal is the state saying no: the walk goes on
			actions[len(actions)-1] += " !" + err.Error()
			if strings.Contains(err.Error(), rulesxxvc.ErrNotPorted.Error()) { // but not for an action the ruleset does not answer: that is a finding
				sweepNotPorted[sweepDigits.ReplaceAllString(err.Error(), "#")+" @ -"]++
			}
		}
	}
	if sweepDump {
		t.Logf("%q", actions)
	}
	return nil
}

var sweepDump bool

// sweepAct is one thing a front does: a label for the log, and the doing.
type sweepAct struct {
	label string
	run   func(e *Engine) error
}

func sweepDo(a game.Action) *sweepAct {
	return &sweepAct{fmt.Sprintf("%T%+v", a, a), func(e *Engine) error { _, err := e.doNow(a); return err }}
}

// sweepStep is gobox's step ahead with the square's event run.
var sweepStep = &sweepAct{"step", br1North}

// sweepWaits is whether the engine runs on by itself (a fight's monsters, a pause): Step moves it.
func sweepWaits(s game.State) bool {
	switch s.(type) {
	case game.Exploring, game.Fighting:
		return true
	}
	return false
}

// sweepNotPorted counts the refusals that say "not ported yet", by message: actions a front sent that the ruleset
// does not resolve (the walk goes on, so they are findings of their own).
var sweepNotPorted = map[string]int{}

// sweepDigits makes findings that differ in a number only one group.
var sweepDigits = regexp.MustCompile(`\d+`)

var sweepTexts = []string{"A", "BUCK", "12", "0", "999"}

// sweepChoose is the walk's next action from what the state allows, nil if it allows none.
func sweepChoose(e *Engine, rng *rules.RNG) *sweepAct {
	s := e.build()
	c := commonOf(s)
	if sc, ok := s.(game.Scene); ok && sc.Kind == game.SceneHandout {
		return sweepDo(game.Leave{})
	}
	if q := c.Ask; q != nil {
		switch {
		case q.Typed || q.Number:
			txt := sweepTexts[rng.N(len(sweepTexts))]
			return sweepDo(game.Reply{Text: txt})
		case q.Continue:
			return sweepDo(game.ContinueAction{})
		case len(q.Replies) > 0:
			return sweepDo(game.Reply{N: sweepReplyFor(e, rng, len(q.Replies))})
		}
		return sweepDo(game.Reply{})
	}
	switch v := s.(type) {
	case game.Fighting:
		if !myTurn(s) {
			return nil
		}
		var opts []game.Action
		for _, a := range v.Allowed.Plain {
			if _, ok := a.(game.Attack); ok {
				opts = append(opts, a, a)
			}
		}
		for _, m := range v.Allowed.Moves {
			opts = append(opts, game.MoveTo{To: m.At})
		}
		return sweepDo(append(opts, game.EndTurn{})[rng.N(len(opts)+1)])
	case game.Looting:
		if len(v.Items) > 0 && rng.N(2) == 0 {
			return sweepDo(game.Take{Item: v.Items[rng.N(len(v.Items))].Handle})
		}
		if slices.Contains(v.Allowed.Plain, game.Action(game.ContinueAction{})) {
			return sweepDo(game.ContinueAction{})
		}
		return sweepDo(game.Leave{})
	case game.Exploring:
		switch r := rng.N(6); {
		case r < 3:
			return sweepStep
		case r == 3:
			return sweepDo(game.Face{Facing: rng.N(4) * 2})
		}
		if plain := v.Allowed.Plain; len(plain) > 0 {
			return sweepDo(plain[rng.N(len(plain))])
		}
		return sweepStep
	}
	if len(c.Allowed.Plain) == 0 {
		return nil
	}
	for _, a := range c.Allowed.Plain {
		if _, ok := a.(game.Begin); ok {
			return sweepDo(a)
		}
	}
	return sweepDo(c.Allowed.Plain[rng.N(len(c.Allowed.Plain))])
}

// sweepEntries are the squares a block is entered at: its event squares first, then others, up to n.
func sweepEntries(t testing.TB, adv string, block, n int, rng *rules.RNG) (out [][2]int) {
	defer func() {
		if recover() != nil { // the block's own entry panics: the walks report it, here one square is enough
			out = [][2]int{{0, 0}}
		}
	}()
	e := br1BeginOn(t, br1EngineIn(t, adv))
	e.push(cont{kind: kStepEnd})
	e.DevAt(block, 0, 0)
	ex, ok := e.build().(game.Exploring)
	if !ok || ex.Map.Walk == nil {
		return [][2]int{{0, 0}}
	}
	size := ex.Map.Walk.Size()
	var events, others [][2]int
	for y := range size.Y {
		for x := range size.X {
			if ex.Map.Walk.Event(image.Pt(x, y)) {
				events = append(events, [2]int{x, y})
			} else {
				others = append(others, [2]int{x, y})
			}
		}
	}
	for len(out) < n && len(events)+len(others) > 0 {
		pool := &events
		if len(events) == 0 || len(others) > 0 && len(out)%3 == 2 {
			pool = &others
		}
		i := rng.N(len(*pool))
		out = append(out, (*pool)[i])
		*pool = slices.Delete(*pool, i, i+1)
	}
	return out
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
	entries := sweepEntriesN
	if os.Getenv("GOBOX_SWEEP") == "heavy" {
		entries = sweepEntriesN * 6
	}
	var logged strings.Builder // what the engine logs (a script that did not resolve a field...) is a finding too
	log.SetFlags(0)
	log.SetOutput(&logged)
	defer log.SetOutput(os.Stderr)
	clear(sweepNotPorted)
	hit := map[int]map[int]bool{}
	groups := map[string][]*sweepFinding{}
	for _, block := range br1Blocks {
		sweepRunBlock(t, adv, block, entries, hit, groups, nil)
	}

	for k, n := range sweepNotPorted {
		for range n {
			groups["REFUSED: "+k] = append(groups["REFUSED: "+k], &sweepFinding{})
		}
	}
	for _, l := range strings.Split(logged.String(), "\n") {
		if l != "" {
			groups["LOG: "+l+" @ -"] = append(groups["LOG: "+l+" @ -"], &sweepFinding{})
		}
	}
	e := br1EngineIn(t, adv)
	t.Logf("%-6s %8s %8s", "block", "run", "total")
	for _, block := range br1Blocks {
		s, err := e.parseScriptFile(block)
		if err != nil {
			t.Fatal(block, err)
		}
		t.Logf("%-6d %8d %8d", block, len(hit[block]), len(s.Code))
	}
	sweepCoverageTotal(t, e, hit)

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(groups[keys[i]]) > len(groups[keys[j]]) || len(groups[keys[i]]) == len(groups[keys[j]]) && keys[i] < keys[j]
	})
	if f := os.Getenv("GOBOX_SWEEP_DUMP"); f != "" { // one walk of that block, whole
		n, _ := strconv.Atoi(f)
		sp := sweepSpec{Block: n, Facing: 0, Seed: 77}
		t.Logf("dump %+v", sp)
		sweepDump = true
		br1SweepWalk(t, adv, sp, hit)
	}
	for _, k := range keys {
		g := groups[k]
		blocks := map[int]bool{}
		for _, f := range g {
			blocks[f.Spec.Block] = true
		}
		first := g[0]
		t.Logf("x%d blocks %v: %s\n  first: %+v\n  actions (all %d, last 15): %q\n  frames:\n%s", len(g), sortedInts(blocks), k, first.Spec, len(first.Actions), first.Actions[max(0, len(first.Actions)-15):], first.Frames)
	}
	for _, k := range keys {
		if !slices.ContainsFunc(br1SweepKnown, func(kn struct{ Pattern, Why string }) bool { return regexp.MustCompile(kn.Pattern).MatchString(k) }) {
			t.Errorf("new finding (x%d): %s", len(groups[k]), k)
		}
	}
	for _, kn := range br1SweepKnown {
		if !slices.ContainsFunc(keys, func(k string) bool { return regexp.MustCompile(kn.Pattern).MatchString(k) }) {
			t.Errorf("listed finding no longer occurs: %s (%s)", kn.Pattern, kn.Why)
		}
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

var _ = strconv.Itoa

// TestBR1SweepReplay replays one recorded walk: GOBOX_SWEEP_REPLAY="block x y facing seed [state]" (the finding's first: line; its block's earlier walks run first, GOBOX_SWEEP=heavy as the run).
// It fails with the panic a walk recorded, so a fix is proven by the same line passing.
func TestBR1SweepReplay(t *testing.T) {
	var sp sweepSpec
	if n, _ := fmt.Sscan(os.Getenv("GOBOX_SWEEP_REPLAY"), &sp.Block, &sp.X, &sp.Y, &sp.Facing, &sp.Seed, &sp.State); n < 5 { // the state is optional (0)
		t.Skip("GOBOX_SWEEP_REPLAY not set")
	}
	t.Setenv("GOBOX_NOSOUND", "1")
	adv := t.TempDir()
	if err := importer.ImportBREngine("br1", importer.BR1EXE, filepath.Join("..", "data", "br1", "sources", "dos"), adv); err != nil {
		t.Fatal(err)
	}
	entries := sweepEntriesN
	if os.Getenv("GOBOX_SWEEP") == "heavy" {
		entries *= 6
	}
	// the walks of its block before it run first: the guided replies remember them
	if f := sweepRunBlock(t, adv, sp.Block, entries, map[int]map[int]bool{}, nil, &sp); f != nil {
		t.Fatalf("%s\n%q\ntail %v\n%s", f.Key, f.Actions, f.Tail, f.Frames)
	}
}

// br1SweepFixed are walks that once panicked and no longer do with that panic (the walk is as it was when the fix
// went in, so a later change of the engine may move it off its path: then the fix needs a unit test of its own).
var br1SweepFixed = []struct {
	Spec sweepSpec
	Was  string // the finding's text: the walk may go on to another
	Why  string
}{
	{sweepSpec{Block: 16, Facing: 2, Seed: 14885212}, "fact rules.UnitFact: fight fact: no such character", "turnChoices asked UnitFact of a ground attack's Target (NoChar): an area weapon's aimed squares have no unit"},
}

func TestBR1SweepFixed(t *testing.T) {
	t.Setenv("GOBOX_NOSOUND", "1")
	adv := t.TempDir()
	if err := importer.ImportBREngine("br1", importer.BR1EXE, filepath.Join("..", "data", "br1", "sources", "dos"), adv); err != nil {
		t.Skip("no BR1 originals:", err)
	}
	for _, c := range br1SweepFixed {
		if f := br1SweepWalk(t, adv, c.Spec, map[int]map[int]bool{}); f != nil && strings.Contains(f.Key, c.Was) {
			t.Errorf("%+v (%s): %s\n%s", c.Spec, c.Why, f.Key, f.Frames)
		}
	}
}
