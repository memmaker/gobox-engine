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
)

// The BR1 sweep: every script block entered at several squares, each walked by a bounded random sequence of the
// actions a front could send (chosen from what the state allows, by a seeded rules.RNG). A panic ends its walk and
// is a finding; findings are grouped and compared with br1SweepKnown. GOBOX_SWEEP=heavy runs ten times the walks.

// br1SweepKnown are the findings still open: Pattern is matched against "message @ top frame".
// The test fails on a finding no pattern matches and on a pattern no finding matches.
var br1SweepKnown = []struct{ Pattern, Why string }{
	{`^rules: rules.EndSession: .*not ported yet`, "sw1: the party killed by a script (Damage) ends the game; rulesxxvc answers no EndSession"},
	{`^fact rules.FlightFact: Fact 0: no character`, "open, owner presentation (vw1): the engine's missileArt is the Pool games' art; BR1's AnimateMissile (cbtacts.pas) has its own, rulesxxvc answers no FlightFact"},
	{`^rules: rules.AddLoot: .*not ported yet`, "sw1: BR's TreasureOf(credits, count, items) goes through the Pool handler (4 fixed operands, platinum/gems/jewelry), and rulesxxvc answers no AddLoot"},
	{`^rules: rules.GainExperience: .*not ported yet`, "sw1: BR's GainExperience (eclvm.pas EclAwardExperience), rulesxxvc answers no GainExperience"},
	{`^ecl: view kind 2 not ported`, "vw1: the rocket's view (AreaKind view kind 2)"},
	{`^ecl: opcode 4B .* not ported`, "sh1: the Shop command"},
	{`^computer turn: Move .*not reachable with the movement left`, "sw1: a computer-turn move the ruleset offered and then refused"},
	{`^LOG: script: area \d+: T_\w+\[\d+\] is past the field's \d+ entries`, "none: the entry skips the story, so the index variable (Word4C61 ...) holds 0 and the table is read past its end as the original does (engine log, TableFaults)"},
}

const (
	sweepEntriesN = 12  // entry squares per block (heavy: x6)
	sweepSteps    = 400 // actions per walk
	sweepSeed     = 0x5eed
)

// sweepSpec is one walk, enough to replay it: br1SweepWalk(t, adv, spec, nil).
type sweepSpec struct {
	Block, X, Y, Facing int
	Seed                uint32
}

type sweepFinding struct {
	Key     string // message @ top frame
	Spec    sweepSpec
	Actions []string
	Frames  string
}

// sweepWalk runs one walk on a new engine and returns its finding, nil if none. hit collects the executed
// instructions per block.
func br1SweepWalk(t testing.TB, adv string, sp sweepSpec, hit map[int]map[int]bool) (f *sweepFinding) {
	var actions []string
	defer func() {
		if r := recover(); r != nil {
			f = &sweepFinding{Spec: sp, Actions: actions}
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
			f.Key = msg + " @ " + top
		}
	}()
	e := br1BeginOn(t, br1EngineIn(t, adv))
	e.vm.Trace = func(vm *ecl.VM, n int) {
		if vm.Script == nil {
			return
		}
		if hit[vm.Script.Area] == nil {
			hit[vm.Script.Area] = map[int]bool{}
		}
		hit[vm.Script.Area][vm.Cur()] = true
	}
	rng := &rules.RNG{State: sp.Seed}
	e.push(cont{kind: kStepEnd})
	e.facing = sp.Facing
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
		actions = append(actions, a.label)
		if err := a.run(e); err != nil { // a refusal is the state saying no: the walk goes on
			actions[len(actions)-1] += " !" + err.Error()
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
			return sweepDo(game.Reply{N: rng.N(len(q.Replies))})
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
		// MakeCamp is left out: BR's rest is Program 9 (vw1's), every walk that tried it ended there (rules.CampOpen unanswered, x162 in the first sweep)
		plain := slices.DeleteFunc(slices.Clone(v.Allowed.Plain), func(a game.Action) bool { _, ok := a.(game.MakeCamp); return ok })
		if len(plain) > 0 {
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
	entries, walks := sweepEntriesN, 2
	if os.Getenv("GOBOX_SWEEP") == "heavy" {
		entries, walks = sweepEntriesN*6, 2
	}
	var logged strings.Builder // what the engine logs (a script that did not resolve a field...) is a finding too
	log.SetFlags(0)
	log.SetOutput(&logged)
	defer log.SetOutput(os.Stderr)
	rng := &rules.RNG{State: sweepSeed}
	hit := map[int]map[int]bool{}
	groups := map[string][]*sweepFinding{}
	for _, block := range br1Blocks {
		for _, at := range sweepEntries(t, adv, block, entries, rng) {
			for range walks {
				sp := sweepSpec{Block: block, X: at[0], Y: at[1], Facing: rng.N(4) * 2, Seed: uint32(rng.N(1<<30)) + 1}
				if f := br1SweepWalk(t, adv, sp, hit); f != nil {
					groups[f.Key] = append(groups[f.Key], f)
				}
			}
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
