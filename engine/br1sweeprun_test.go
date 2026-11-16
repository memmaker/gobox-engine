package engine

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"testing"
	"time"

	"gobox/ecl"
	"gobox/game"
	"gobox/rules"
	"gobox/rulesxxvc"
)

// The BR1 sweep's runs (br1sweep_test.go). A run is a block entered in an entry state, then played by TOKENS: one
// token is one thing a front sends (a reply, a step, a Quick). The tokens of a run are all the run's input, so a
// run is replayed by executing its tokens on a new engine (the engine is deterministic: one seed). The policy that
// picks the next token (sweepRun.choose) looks at the state and at what the whole sweep has taken so far; it is
// never replayed, only its tokens are.

// sweepTokKind is the kind of a token.
type sweepTokKind byte

const (
	tokReply      sweepTokKind = iota // Reply{N}
	tokText                           // Reply{Text: sweepTexts[N]}
	tokContinue                       // ContinueAction
	tokLeave                          // Leave
	tokFace                           // Face{Facing: N}
	tokStep                           // a step ahead with the square's event run
	tokEngineStep                     // the engine's own pace (Step): the foes' turns, a pause
	tokPress                          // what the state waits for, the one way it has (press)
	tokQuick                          // Quick: the computer plays the member's turn
	tokTakeAll                        // Take{All}
	tokBuy                            // Buy the shop's first good for the first member
	tokHeading                        // Heading{Dir: N}
)

var sweepTokLetters = map[sweepTokKind]string{tokReply: "r", tokText: "x", tokContinue: "c", tokLeave: "l", tokFace: "f", tokStep: "s", tokEngineStep: "e", tokPress: "p", tokQuick: "q", tokTakeAll: "t", tokBuy: "b", tokHeading: "h"}

type sweepTok struct {
	Kind sweepTokKind
	N    int
}

func (t sweepTok) String() string { return fmt.Sprintf("%s%d", sweepTokLetters[t.Kind], t.N) }

func parseSweepToks(s string) (out []sweepTok, err error) {
	for _, w := range strings.Fields(s) {
		var tok sweepTok
		found := false
		for k, l := range sweepTokLetters {
			if strings.HasPrefix(w, l) {
				if _, err := fmt.Sscanf(w[len(l):], "%d", &tok.N); err == nil {
					tok.Kind, found = k, true
					break
				}
			}
		}
		if !found {
			return nil, fmt.Errorf("sweep: bad token %q", w)
		}
		out = append(out, tok)
	}
	return out, nil
}

func sweepToksString(toks []sweepTok) string {
	parts := make([]string, len(toks))
	for i, t := range toks {
		parts[i] = t.String()
	}
	return strings.Join(parts, " ")
}

// sweepExec is a token's doing: what a front would send. An error is the state refusing (the run goes on).
func sweepExec(e *Engine, t sweepTok) error {
	var a game.Action
	switch t.Kind {
	case tokReply:
		a = game.Reply{N: t.N}
	case tokText:
		a = game.Reply{Text: sweepTexts[t.N]}
	case tokContinue:
		a = game.ContinueAction{}
	case tokLeave:
		a = game.Leave{}
	case tokFace:
		a = game.Face{Facing: t.N}
	case tokStep: // N the turn from the facing the step is taken in (0 ahead): where the view cannot turn (the overland) the step goes its way as it is
		if err := stepVia(e, e.Do, t.N); err != nil {
			return err
		}
		if e.ext.owed == nil {
			return nil
		}
		_, err := e.Do(game.RunEvent{})
		return err
	case tokEngineStep: // the engine's own pace until it asks the player (the computers' turns, one after the other: Step would build the state after each)
		for n := 0; n < sweepPaceMost && e.pace() == PaceFight; n++ {
			e.CombatStep()
		}
		e.rebuild()
		return nil
	case tokPress: // what the state waits for: its Continue, its Leave, else the engine's own pace
		al := sweepCommon(e.State()).Allowed
		switch {
		case allows(al, game.ContinueAction{}):
			a = game.ContinueAction{}
		case allows(al, game.Leave{}):
			a = game.Leave{}
		default:
			e.Step()
			return nil
		}
	case tokQuick:
		a = game.Quick{}
	case tokTakeAll:
		a = game.Take{All: true}
	case tokBuy:
		tr, ok := e.State().(game.Trading)
		if !ok || len(tr.Goods) == 0 {
			return rules.ErrNotNow
		}
		a = game.Buy{Member: 0, Item: tr.Goods[0].Handle}
	case tokHeading:
		a = game.Heading{Dir: t.N}
	default:
		panic(fmt.Sprintf("sweep: token kind %d", t.Kind))
	}
	_, err := e.Do(a)
	return err
}

var sweepTexts = []string{"A", "BUCK", "12", "0", "999"}

// sweepPos is a place in the scripts: the block and the instruction of the command that is running.
type sweepPos struct{ Block, Instr int }

func (p sweepPos) String() string { return fmt.Sprintf("%d:%d", p.Block, p.Instr) }

// sweepEntry is where a run starts: block entered in state State (br1sweepstate_test.go) at X, Y facing Facing.
type sweepEntry struct{ Block, State, X, Y, Facing int }

func (en sweepEntry) String() string {
	return fmt.Sprintf("%d/%d/%d,%d/%d", en.Block, en.State, en.X, en.Y, en.Facing)
}

// sweepItem is a run to make: the entry and the tokens to replay (Prefix, shared with the run it was made in and
// never written, then Last). After them the policy plays on. An item made at a question has At and Of: the question
// Last answers and how many replies it had (checked on replay: the replay must come to the same question).
type sweepItem struct {
	Entry  sweepEntry
	Prefix []sweepTok
	Last   *sweepTok
	At     sweepPos
	Of     int
}

// tokens are the tokens to replay.
func (it sweepItem) tokens() []sweepTok {
	out := append([]sweepTok(nil), it.Prefix...)
	if it.Last != nil {
		out = append(out, *it.Last)
	}
	return out
}

// sweepFinding is one thing the sweep found.
type sweepFinding struct {
	Key   string // message @ top frame
	Item  sweepItem
	Toks  []sweepTok // the run's tokens up to the finding
	Tail  []string   // the last script positions run, oldest first
	Frame string
}

// The most tokens of a run, and the fights of one run (a fight on ten members costs most of a run's time); the heavy
// sweep (GOBOX_SWEEP=heavy) takes the larger.
const (
	sweepRunTokens   = 250
	sweepHeavyTokens = 2500
	sweepRunFights   = 2
	sweepHeavyFights = 1000
)

// sweepInstructionBudget bounds the instructions one token may run.
const sweepInstructionBudget = 200000

// sweepSameMost is how many map tokens in a row may change neither the square nor the facing: a step that fails is marked shut,
// so a longer run is a walk going nowhere.
const sweepSameMost = 12

// sweepPaceMost bounds the computers' turns one pace token runs.
const sweepPaceMost = 64

// sweepFightCap bounds the tokens of one fight (a pace token runs up to sweepPaceMost turns of the computers): a fight that does
// not end in them is a finding.
const sweepFightCap = 80

// sweepRun is one run in progress.
type sweepRun struct {
	w       *sweepWorld
	t       testing.TB
	e       *Engine
	item    sweepItem
	toks    []sweepTok
	visits  map[sweepPos]int
	inFight bool
	fightAt sweepPos
	fightN  int // answers of the fight in play
	fights  int // fights this run has begun
	bought  bool
	headed  int
	sawMap  bool
	last    sweepPos // the script position last run
	tail    []string
	ran     int
	hop     *sweepHop
	blocked map[[3]int]bool // x, y, side: a step there did not move the party
	stuck   int             // map tokens in a row that moved nobody and turned nobody: a walk going nowhere
	prev    *sweepSpot      // where the party stood when the last map token was chosen
}

// sweepSpot is a square and a facing.
type sweepSpot struct {
	At     image.Point
	Facing int
}

// sweepHop is a step toward a square: where it was taken from.
type sweepHop struct {
	From image.Point
	Dir  int
}

// sweepKey is (block, state): the map's exploration is shared by the runs of one entry state.
type sweepKey struct{ Block, State int }

// sweepMap is what the runs of one (block, state) have explored of its map.
type sweepMap struct {
	entered map[image.Point]bool
	tried   map[[3]int]bool // x, y, facing
	size    image.Point
	entries int // entries pushed for squares nobody reached
}

// sweepWorld is what all runs share.
type sweepWorld struct {
	adv       string
	blocks    []int // the game's script blocks
	states    map[int][][]sweepCell
	facing    map[int]bool // the block's script reads the facing
	hit       map[int]map[int]bool
	taken     map[sweepPos]map[int]bool
	sched     map[sweepPos]map[int]bool
	offered   map[sweepPos]int // replies a question offered
	queue     map[int][]sweepItem
	maps      map[sweepKey]*sweepMap
	groups    map[string][]*sweepFinding
	lost      map[sweepPos]int
	won       map[sweepPos]int
	runs      map[int]int
	dropped   int
	diverged  int
	wedged    map[int]int // walks given up for going nowhere, by block
	notPorted map[string]int
	heavy     bool                  // GOBOX_SWEEP=heavy
	work      int                   // the work the runs have taken
	took      map[int]time.Duration // how long a block took
	over      map[sweepPos]int      // where runs ended with the game over (the party menu again): the last script place run
	workCap   int
	statesMax int            // the most entry states of a block
	tokenCap  int            // the most tokens of a run
	fightCap  int            // the most fights of a run
	seen      map[string]int // the states' types the policy met
}

func newSweepWorld(adv string, blocks []int, heavy bool) *sweepWorld {
	w := &sweepWorld{adv: adv, blocks: blocks, workCap: sweepWorkCap, tokenCap: sweepRunTokens, fightCap: sweepRunFights, statesMax: sweepStatesMax, states: map[int][][]sweepCell{}, facing: map[int]bool{}, hit: map[int]map[int]bool{},
		taken: map[sweepPos]map[int]bool{}, sched: map[sweepPos]map[int]bool{}, offered: map[sweepPos]int{},
		queue: map[int][]sweepItem{}, maps: map[sweepKey]*sweepMap{}, groups: map[string][]*sweepFinding{},
		lost: map[sweepPos]int{}, won: map[sweepPos]int{}, took: map[int]time.Duration{}, over: map[sweepPos]int{}, runs: map[int]int{}, wedged: map[int]int{}, notPorted: map[string]int{}, seen: map[string]int{}}
	if heavy {
		w.heavy, w.workCap, w.tokenCap, w.fightCap, w.statesMax = true, sweepHeavyWork, sweepHeavyTokens, sweepHeavyFights, sweepHeavyStates
	}
	return w
}

func (w *sweepWorld) mapOf(en sweepEntry) *sweepMap {
	k := sweepKey{en.Block, en.State}
	if w.maps[k] == nil {
		w.maps[k] = &sweepMap{entered: map[image.Point]bool{}, tried: map[[3]int]bool{}}
	}
	return w.maps[k]
}

// sweepCommon is the common part of any state the sweep meets.
func sweepCommon(s game.State) game.Common {
	if f, ok := s.(game.ShipFight); ok {
		return f.Common
	}
	return commonOf(s)
}

// run plays one item on a new engine; a panic is the finding returned.
func (w *sweepWorld) run(t testing.TB, it sweepItem) (f *sweepFinding) {
	r := &sweepRun{w: w, t: t, item: it, visits: map[sweepPos]int{}, blocked: map[[3]int]bool{}}
	defer func() { w.work += len(r.toks) + sweepFightWork*r.fights }()
	defer func() {
		if p := recover(); p != nil {
			f = &sweepFinding{Item: it, Toks: r.toks, Tail: r.tail}
			msg, top := fmt.Sprint(p), ""
			lines := strings.Split(string(debug.Stack()), "\n")
			after := false
			for i, l := range lines {
				if strings.HasPrefix(l, "panic(") {
					after = true
					continue
				}
				if !after || l == "" || strings.HasPrefix(l, "\t") || !strings.Contains(l, "gobox/") || strings.Contains(l, "sweepRun") || strings.Contains(l, "sweepWorld") {
					continue
				}
				cut := strings.LastIndex(l, "(")
				if cut < 0 {
					continue
				}
				frame := l[:cut]
				if top == "" {
					top = frame
				}
				if len(f.Frame) < 600 && i+1 < len(lines) {
					f.Frame += strings.TrimSpace(frame) + " " + strings.TrimSpace(lines[i+1]) + "\n"
				}
			}
			f.Key = sweepDigits.ReplaceAllString(msg, "#") + " @ " + top
		}
	}()
	en := it.Entry
	if w.states[en.Block] == nil {
		w.states[en.Block] = sweepStatesOf(t, w.adv, en.Block, w.statesMax)
		w.facing[en.Block] = sweepReadsFacing(t, w.adv, en.Block)
	}
	r.e = br1BeginWith(t, br1EngineIn(t, w.adv), sweepParty)
	e := r.e
	r.wrapOps(e.vm)
	e.push(cont{kind: kStepEnd})
	e.facing = en.Facing
	e.DevAtWith(en.Block, en.X, en.Y, func() { sweepApply(t, e, w.states[en.Block][en.State]) })
	e.rebuild()
	replay := it.tokens()
	for len(r.toks) < max(w.tokenCap, len(replay)) {
		s := e.State()
		r.observe(s)
		var tok sweepTok
		if len(r.toks) < len(replay) {
			tok = replay[len(r.toks)]
			if len(r.toks) == len(replay)-1 && it.Of > 0 {
				if q := sweepCommon(s).Ask; q == nil || r.pos() != it.At || sweepChoices(q) != it.Of {
					w.diverged++
					w.add(&sweepFinding{Key: fmt.Sprintf("REPLAY DIVERGED: no question of %d at %v @ -", it.Of, it.At), Item: it, Toks: r.toks, Tail: r.tail})
					return nil
				}
			}
		} else {
			var ok bool
			if tok, ok = r.choose(s); !ok {
				break
			}
		}
		r.toks = append(r.toks, tok)
		r.ran = 0
		if err := sweepExec(e, tok); err != nil && strings.Contains(err.Error(), rulesxxvc.ErrNotPorted.Error()) { // a refusal is the state saying no; an action the ruleset does not answer is a finding
			w.notPorted[sweepDigits.ReplaceAllString(err.Error(), "#")]++
		}
		if r.lostFight(e) {
			break
		}
	}
	w.runEnd(r)
	return nil
}

func (w *sweepWorld) add(f *sweepFinding) { w.groups[f.Key] = append(w.groups[f.Key], f) }

// exec is the instruction about to run (every command the engine runs goes through the VM's op table, wrapped by
// wrapOps; a command an If skips never does, and the commands with no operands, If and Exit among them, do): the
// coverage and the places the sweep reports are of these.
func (r *sweepRun) exec(vm *ecl.VM) {
	if vm.Script == nil {
		return
	}
	if r.ran++; r.ran > sweepInstructionBudget {
		panic(fmt.Sprintf("sweep: a script ran %d instructions on one action without asking (block %d)", sweepInstructionBudget, vm.Script.Area))
	}
	h := r.w.hit[vm.Script.Area]
	if h == nil {
		h = map[int]bool{}
		r.w.hit[vm.Script.Area] = h
	}
	h[vm.Cur()] = true
	r.last = sweepPos{vm.Script.Area, vm.Cur()}
	if r.tail = append(r.tail, r.last.String()); len(r.tail) > 40 {
		r.tail = r.tail[1:]
	}
}

// wrapOps makes every command of the VM's op table note itself first.
func (r *sweepRun) wrapOps(vm *ecl.VM) {
	for b, op := range vm.Ops {
		if op != nil {
			vm.Ops[b] = func(vm *ecl.VM, h ecl.Host) {
				r.exec(vm)
				op(vm, h)
			}
		}
	}
}

// pos is the place the engine's running script command stands at.
func (r *sweepRun) pos() sweepPos {
	if r.e.vm.Script == nil {
		return sweepPos{}
	}
	return sweepPos{r.e.vm.Script.Area, r.e.vm.Cur()}
}

// observe takes note of the state: the square stood on, the fight begun or ended.
func (r *sweepRun) observe(s game.State) {
	r.w.seen[fmt.Sprintf("%T", s)]++
	switch v := s.(type) {
	case game.Exploring:
		m := r.w.mapOf(r.item.Entry)
		if v.Map.Walk != nil {
			m.size = v.Map.Walk.Size()
		}
		m.entered[image.Pt(v.Where.X, v.Where.Y)] = true
		r.sawMap = true
	case game.Fighting:
		if !r.inFight {
			r.inFight, r.fightAt, r.fightN = true, r.last, 0
			r.fights++
		}
		return
	}
	if r.inFight {
		r.inFight = false
		if _, over := s.(game.Assembling); !over {
			r.w.won[r.fightAt]++
		}
	}
}

// lostFight is whether the party was beaten (the game is back at the party menu: postcbt.pas RESULT_TEAM_BEATEN): the
// run ends and the fight is noted as lost.
func (r *sweepRun) lostFight(e *Engine) bool {
	if !r.inFight {
		return false
	}
	if _, over := e.State().(game.Assembling); over {
		r.w.lost[r.fightAt]++
		r.inFight = false
		return true
	}
	return false
}

// sweepChoices is how many different replies a question has: the sweep takes each at least once.
func sweepChoices(q *game.Question) int {
	switch {
	case q.Typed || q.Number:
		return len(sweepTexts)
	case q.Continue:
		return 1
	case len(q.Replies) > 0:
		return len(q.Replies)
	}
	return 1
}

func sweepAnswer(q *game.Question, i int) sweepTok {
	switch {
	case q.Typed || q.Number:
		return sweepTok{tokText, i}
	case q.Continue:
		return sweepTok{tokContinue, 0}
	}
	return sweepTok{tokReply, i}
}

// choose is the policy: the next token for state s, ok false when the run is over.
func (r *sweepRun) choose(s game.State) (sweepTok, bool) {
	e := r.e
	c := sweepCommon(s)
	if sc, ok := s.(game.Scene); ok && sc.Kind == game.SceneHandout {
		return sweepTok{tokLeave, 0}, true
	}
	if q := c.Ask; q != nil {
		if _, ship := s.(game.ShipFight); ship {
			return r.shipAnswer(q), true
		}
		if q.Heading {
			if r.headed++; r.headed == 1 {
				return sweepTok{tokHeading, (len(r.toks) % 4) * 2}, true
			}
			return sweepTok{tokLeave, 0}, true
		}
		return r.question(q), true
	}
	switch v := s.(type) {
	case game.Fighting:
		if r.fightN++; r.fightN > sweepFightCap {
			panic(fmt.Sprintf("sweep: a fight did not end in %d answers (block %d)", sweepFightCap, r.fightAt.Block))
		}
		switch {
		case myTurn(s):
			return sweepTok{tokQuick, 0}, true
		case e.waiting(PendFight):
			return sweepTok{tokEngineStep, 0}, true
		}
		return sweepTok{tokPress, 0}, true
	case game.Looting:
		if playAllows(v, game.ContinueAction{}) {
			return sweepTok{tokContinue, 0}, true
		}
		if len(v.Items) > 0 && len(r.toks) > 0 && r.toks[len(r.toks)-1].Kind != tokTakeAll {
			return sweepTok{tokTakeAll, 0}, true
		}
		return sweepTok{tokLeave, 0}, true
	case game.Trading:
		if !r.bought && v.Kind == game.TradeShop && len(v.Goods) > 0 {
			r.bought = true
			return sweepTok{tokBuy, 0}, true
		}
		return sweepTok{tokLeave, 0}, true
	case game.Exploring:
		if r.fights > r.w.fightCap {
			return sweepTok{}, false
		}
		return r.step(v)
	case game.Assembling: // the party menu again: a script opened it (Program 0, HQ's team menu; the run ends there: a long game on is not what the sweep is for) or nobody is left
		r.w.over[r.last]++
		if sc := e.vm.Script; sc != nil && sc.Area == r.last.Block && r.last.Instr < len(sc.Code) && sc.Code[r.last.Instr].Op == ecl.OpCombat && !r.inFight {
			r.w.lost[r.last]++ // the party was beaten by a Combat whose fight was over inside one token
		}
		return sweepTok{}, false
	case game.Over: // the game is over
		r.w.over[r.last]++
		return sweepTok{}, false
	}
	if playAllows(s, game.ContinueAction{}) {
		return sweepTok{tokContinue, 0}, true
	}
	if len(c.Allowed.Plain) == 0 {
		panic(fmt.Sprintf("STUCK: %T, no allowed action, no question", s))
	}
	return sweepTok{tokPress, 0}, true
}

// shipAnswer is the ship fight's menu: Fire a while, then Quit (the simplest complete fight).
func (r *sweepRun) shipAnswer(q *game.Question) sweepTok {
	if q.Continue {
		return sweepTok{tokContinue, 0}
	}
	r.fightN++
	want := "Fire"
	if r.fightN > 60 {
		want = "Quit"
	}
	for i, name := range q.Replies {
		if name == want {
			return sweepTok{tokReply, i}
		}
	}
	return sweepTok{tokReply, 0}
}

// question is the reply to an ordinary question: the first reply never taken at this place by any run, else the
// reply by how often this run has been here (a menu that comes back is left by its other replies in turn). The
// replies no run has taken are put on the work list.
func (r *sweepRun) question(q *game.Question) sweepTok {
	w := r.w
	pos := r.pos()
	n := sweepChoices(q)
	w.offered[pos] = max(w.offered[pos], n)
	if w.taken[pos] == nil {
		w.taken[pos], w.sched[pos] = map[int]bool{}, map[int]bool{}
	}
	pick := -1
	for i := range n {
		if !w.taken[pos][i] && !w.sched[pos][i] {
			pick = i
			break
		}
	}
	if pick < 0 {
		pick = r.visits[pos] % n
	}
	r.visits[pos]++
	w.taken[pos][pick], w.sched[pos][pick] = true, true
	for i := range n {
		if i == pick || w.taken[pos][i] || w.sched[pos][i] {
			continue
		}
		w.sched[pos][i] = true
		last := sweepAnswer(q, i)
		w.push(sweepItem{Entry: r.item.Entry, Prefix: r.toks[:len(r.toks):len(r.toks)], Last: &last, At: pos, Of: n})
	}
	return sweepAnswer(q, pick)
}

// sweepQueueMost bounds the items waiting for one block.
const sweepQueueMost = 4000

func (w *sweepWorld) push(it sweepItem) {
	if len(w.queue[it.Entry.Block]) >= sweepQueueMost {
		w.dropped++
		return
	}
	w.queue[it.Entry.Block] = append(w.queue[it.Entry.Block], it)
}

// sweepReadsFacing is whether the block's script reads the facing (its move or step entry tests FacingOrRaw or WallAhead).
func sweepReadsFacing(t testing.TB, adv string, block int) bool {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(adv, "scripts", "area"+strconv.Itoa(block)+".txt"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	return strings.Contains(text, "FacingOrRaw") || strings.Contains(text, "WallAhead")
}

// step is the map walk: the nearest square nobody has entered, or (where the script reads the facing) tried in a
// facing, by a breadth-first walk over the sides the map lets a step through.
func (r *sweepRun) step(x game.Exploring) (sweepTok, bool) {
	m := r.w.mapOf(r.item.Entry)
	walk := x.Map.Walk
	if walk == nil {
		return sweepTok{}, false
	}
	at := image.Pt(x.Where.X, x.Where.Y)
	here := sweepSpot{at, x.Where.Facing}
	if r.prev != nil && *r.prev == here { // the last Face or step changed nothing
		if r.stuck++; r.stuck > sweepSameMost {
			r.w.wedged[r.item.Entry.Block]++
			return sweepTok{}, false
		}
	} else {
		r.stuck = 0
	}
	r.prev = &here
	faces := []int{0, 2, 4, 6}
	needFacing := r.w.facing[r.item.Entry.Block]
	if r.hop != nil && at == r.hop.From { // the step did not move the party: that side is shut
		r.blocked[[3]int{at.X, at.Y, r.hop.Dir}] = true
	}
	r.hop = nil
	// facing: the 3D view turns on Face; where it cannot (the overland) a step takes its way from the facing as it is
	facing := func(f int) (tok sweepTok, turned bool) {
		if playAllows(x, game.Face{Facing: f}) {
			return sweepTok{tokFace, f}, true
		}
		return sweepTok{tokStep, (f - x.Where.Facing + 8) % 8}, false
	}
	untried := func(p image.Point) (int, bool) {
		if !needFacing {
			return 0, false
		}
		for _, f := range faces {
			if !m.tried[[3]int{p.X, p.Y, f}] {
				return f, true
			}
		}
		return 0, false
	}
	size := walk.Size()
	type node struct {
		p     image.Point
		first int // the first side of the way here
	}
	seen := map[image.Point]bool{at: true}
	queue := []node{{at, -1}}
	for len(queue) > 0 {
		n := queue[0]
		queue = queue[1:]
		if f, ok := untried(n.p); ok && n.p == at {
			m.tried[[3]int{at.X, at.Y, f}] = true
			tok, turned := facing(f)
			if turned && x.Where.Facing != f {
				m.tried[[3]int{at.X, at.Y, f}] = false
				return tok, true
			}
			r.hop = &sweepHop{at, f}
			return tok, true
		}
		if n.p != at && (!m.entered[n.p] || func() bool { _, ok := untried(n.p); return ok }()) {
			tok, turned := facing(n.first)
			if turned && x.Where.Facing != n.first {
				return tok, true
			}
			m.tried[[3]int{at.X, at.Y, n.first}] = true
			r.hop = &sweepHop{at, n.first}
			return tok, true
		}
		for _, d := range faces {
			to := n.p.Add(sideSteps[d])
			if to.X < 0 || to.Y < 0 || to.X >= size.X || to.Y >= size.Y || seen[to] || r.blocked[[3]int{n.p.X, n.p.Y, d}] || walk.Edge(n.p, d).Kind == game.EdgeWall {
				continue
			}
			seen[to] = true
			first := n.first
			if first < 0 {
				first = d
			}
			queue = append(queue, node{to, first})
		}
	}
	return sweepTok{}, false
}

// runEnd takes the end of a run: squares nobody has entered become entries of their own.
func (w *sweepWorld) runEnd(r *sweepRun) {
	m := w.mapOf(r.item.Entry)
	if !r.sawMap || m.size == (image.Point{}) || m.entries >= sweepEntriesMost {
		return
	}
	for y := range m.size.Y {
		for x := range m.size.X {
			if p := image.Pt(x, y); !m.entered[p] && !m.entryTried(p) {
				m.entries++
				m.entered[p] = false
				w.push(sweepItem{Entry: sweepEntry{r.item.Entry.Block, r.item.Entry.State, x, y, 0}})
				return
			}
		}
	}
}

// sweepEntriesMost bounds the entries made for the squares the walks did not reach, per (block, state).
const sweepEntriesMost = 3

func (m *sweepMap) entryTried(p image.Point) bool {
	_, ok := m.entered[p]
	return ok
}

// runBlock sweeps one entry block's world: its entry states first, then the work list in the order it was made, until
// it is empty, its work or (heavy) its time is used up.
func (w *sweepWorld) runBlock(t testing.TB, block int) {
	if w.states[block] == nil {
		w.states[block] = sweepStatesOf(t, w.adv, block, w.statesMax)
		w.facing[block] = sweepReadsFacing(t, w.adv, block)
	}
	for s := range w.states[block] {
		w.push(sweepItem{Entry: sweepEntry{Block: block, State: s}})
	}
	began := time.Now()
	defer func() { w.took[block] = time.Since(began) }()
	until := began.Add(sweepHeavyTime)
	for len(w.queue[block]) > 0 && w.work < w.workCap && time.Now().Before(until) {
		it := w.queue[block][0]
		w.queue[block] = w.queue[block][1:]
		w.runs[block]++
		if f := w.run(t, it); f != nil {
			w.add(f)
		}
	}
}

// sweepMerge is the report's world: what the blocks' worlds hit, took, found.
func sweepMerge(adv string, blocks []int, heavy bool, ws []*sweepWorld) *sweepWorld {
	m := newSweepWorld(adv, blocks, heavy)
	for _, w := range ws {
		for b, h := range w.hit {
			if m.hit[b] == nil {
				m.hit[b] = map[int]bool{}
			}
			for i := range h {
				m.hit[b][i] = true
			}
		}
		for pos, n := range w.offered {
			m.offered[pos] = max(m.offered[pos], n)
			if m.taken[pos] == nil {
				m.taken[pos] = map[int]bool{}
			}
			for i := range w.taken[pos] {
				m.taken[pos][i] = true
			}
		}
		for pos, n := range w.lost {
			m.lost[pos] += n
		}
		for pos, n := range w.won {
			m.won[pos] += n
		}
		for pos, n := range w.over {
			m.over[pos] += n
		}
		for k, g := range w.groups {
			m.groups[k] = append(m.groups[k], g...)
		}
		for b, n := range w.runs {
			m.runs[b] += n
		}
		for b, q := range w.queue {
			m.queue[b] = append(m.queue[b], q...)
		}
		for b, n := range w.wedged {
			m.wedged[b] += n
		}
		for k, n := range w.notPorted {
			m.notPorted[k] += n
		}
		for k, n := range w.seen {
			m.seen[k] += n
		}
		for b, d := range w.took {
			m.took[b] = d
		}
		m.dropped += w.dropped
		m.diverged += w.diverged
		m.work += w.work
	}
	return m
}
