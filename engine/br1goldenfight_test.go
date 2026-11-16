package engine

import (
	"fmt"
	"slices"
	"testing"

	"gobox/rules"
)

// The old engine's fights (rulesxxvc/testdata/oldengine/fights.json): the rolls the old engine made with gobox's one
// generator, the order the units acted in, what each attack did, the end's experience, credits and loot. The
// shared engine's fight is played from the same inputs with the same seed, every unit by the ruleset's computer
// player (Decide), as the dump's computer played every member.

// The golden's trace events (the dump's traceEv/traceUnit).
type goldenTraceUnit struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Side      int    `json:"side"`
	HP        int    `json:"hp"`
	AC        int    `json:"ac"`
	ACRear    int    `json:"acRear"`
	Hit       int    `json:"hit"`
	Bonus     int    `json:"bonus"`
	Move      int    `json:"move"`
	Attacks   int    `json:"attacks"`
	Attacks2  int    `json:"attacks2"`
	DiceCount [2]int `json:"diceCount"`
	DiceSides [2]int `json:"diceSides"`
	DiceBonus [2]int `json:"diceBonus"`
	X         int    `json:"x"`
	Y         int    `json:"y"`
	Facing    int    `json:"facing"`
	Size      int    `json:"size"`
	Init      int    `json:"init"`
	Tie       int    `json:"tie"`
}

type goldenTraceEv struct {
	Kind    string            `json:"kind"`
	Round   int               `json:"round"`
	State   uint32            `json:"state"`
	Tactics [2]int            `json:"tactics"`
	Units   []goldenTraceUnit `json:"units"`
	Order   []int             `json:"order"`
	Att     int               `json:"att"`
	Tgt     int               `json:"tgt"`
	Roll    int               `json:"roll"`
	Need    int               `json:"need"`
	Hit     bool              `json:"hit"`
	Dmg     int               `json:"dmg"`
	HP      int               `json:"hp"`
	Down    bool              `json:"down"`
}

type goldenItem struct {
	Type  int    `json:"type"`
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type goldenMemberEnd struct {
	Name   string `json:"name"`
	Exp    int    `json:"expGained"`
	HP     int    `json:"hp"`
	Status int    `json:"status"`
}

type goldenFight struct {
	Name     string            `json:"name"`
	Place    string            `json:"place"`
	Members  int               `json:"members"`
	Group    string            `json:"group"`
	Seed     uint32            `json:"seed"`
	Approach int               `json:"approach"`
	Dist     int               `json:"dist"`
	Rounds   int               `json:"rounds"`
	Trace    []goldenTraceEv   `json:"trace"`
	Members2 []goldenMemberEnd `json:"members_end"`
	Credits  int               `json:"credits"`
	ExpEach  int               `json:"expEach"`
	Loot     []goldenItem      `json:"loot"`
	Rewards  []goldenReward    `json:"rewards"`
	Over     string            `json:"over"`
}

// goldenReward is what a monster is worth beaten, from its record at the start (the dump's goldenReward).
type goldenReward struct {
	Unit    int          `json:"unit"`
	Name    string       `json:"name"`
	Exp     int          `json:"exp"`
	Credits int          `json:"credits"`
	Items   []goldenItem `json:"items"`
}

// goldenBlow is one attack's result: who, whom, hit, the damage.
type goldenBlow struct {
	Att, Tgt int
	Hit      bool
	Dmg      int
}

// goldenRoll is one d20 the old engine rolled against a number: who, at whom, the d20 value that hits (armour minus
// the attacker's to-hit and side bonus, plus a blind attacker's 4).
type goldenRoll struct{ Att, Tgt, Need int }

// goldenRound is what the golden says of one round: the head, the blows and the rolls in order.
type goldenRound struct {
	Head  goldenTraceEv
	Blows []goldenBlow
	Rolls []goldenRoll
}

func goldenRounds(trace []goldenTraceEv) []goldenRound {
	var out []goldenRound
	for _, ev := range trace {
		switch ev.Kind {
		case "round":
			out = append(out, goldenRound{Head: ev})
		case "blow":
			r := &out[len(out)-1]
			r.Blows = append(r.Blows, goldenBlow{ev.Att, ev.Tgt, ev.Hit, ev.Dmg})
		case "roll":
			r := &out[len(out)-1]
			r.Rolls = append(r.Rolls, goldenRoll{ev.Att, ev.Tgt, ev.Need})
		}
	}
	return out
}

// goldenFightLimit bounds the rounds one played fight may take (the rules' own limit is the engine's).
const goldenFightLimit = 40

// goldenComputerSteps bounds the actions of one computer turn (the engine's maxComputerSteps).
const goldenComputerSteps = 1000

// goldenPlayed is the shared engine's fight, round by round.
type goldenPlayed struct {
	rounds []goldenPlayedRound
	result rules.FightResult
	units  map[rules.CharID]int // the unit's place in the party list, 1-based: the golden's unit id
	number map[int]int          // the board unit number an effect carries: the same place
	beaten map[int]bool         // the monsters (by place) beaten when the fight ended: not okay, not running
	upLeft int                  // the members okay when the fight ended
}

type goldenPlayedRound struct {
	state   uint32
	order   []int
	blows   []goldenBlow
	hp      map[int]int
	at      map[int]rules.BoardPos
	units   []rules.Unit
	moves   map[int]int
	damage  map[int]string // the dice the unit's next attack rolls, as the units fact shows them, by place
	chances []goldenChance // the hit chance the facts gave just before each single-target attack
}

// goldenChance is HitFact's answer before an attack: from, at whom (places), percent.
type goldenChance struct{ att, tgt, percent int }

// play runs the fight the engine has set up to its end with the generator at seed, the computer on every side.
func (e *Engine) goldenPlay(t *testing.T, seed uint32) goldenPlayed {
	t.Helper()
	rs := e.ruleset()
	rng := rules.RNG{State: seed}
	var p goldenPlayed
	p.units, p.number = map[rules.CharID]int{}, map[int]int{}
	for i, id := range e.partyIDs() {
		p.units[id] = i + 1
	}
	for round := 0; round < goldenFightLimit && rs.Verdict(e.fight()) == rules.FightOn; round++ {
		r := goldenPlayedRound{state: rng.State, hp: map[int]int{}, at: map[int]rules.BoardPos{}, moves: map[int]int{}, damage: map[int]string{}}
		rs.BeginRound(e.fight(), &rng)
		v, err := rs.Fact(rules.NoChar, rules.UnitsFact{})
		if err != nil {
			t.Fatal(err)
		}
		r.units = v.(rules.Units)
		for _, u := range r.units {
			r.hp[p.units[u.ID]] = u.HP
			r.at[p.units[u.ID]] = u.At
			r.damage[p.units[u.ID]] = u.Damage
			p.number[u.Number] = p.units[u.ID]
		}
		for id := range p.units {
			c, err := rs.Fact(id, rules.FightCardFact{})
			if err != nil {
				t.Fatal(err)
			}
			r.moves[p.units[id]] = c.(rules.FightCard).Moves
		}
		for {
			id, ok := rs.NextActor(e.fight(), &rng)
			if !ok {
				break
			}
			r.order = append(r.order, p.units[id])
			if turn := rs.BeginTurn(e.fight(), id, &rng); !turn.Skipped {
				for step := 0; ; step++ {
					if step == goldenComputerSteps {
						t.Fatalf("unit %d: no action used the turn up in %d steps", p.units[id], step)
					}
					f := e.fightBy(id)
					a := rs.Decide(f, id, &rng)
					if at, ok := a.(rules.Attack); ok && at.At == nil {
						v, err := rs.Fact(id, rules.HitFact{Target: at.Target})
						if err != nil {
							t.Fatal(err)
						}
						r.chances = append(r.chances, goldenChance{p.units[id], p.units[at.Target], int(v.(rules.Number))})
					}
					out, err := rs.Resolve(f, a, &rng)
					if err != nil {
						t.Fatalf("unit %d %T: %v", p.units[id], a, err)
					}
					for _, ef := range out.Effects {
						if at, ok := ef.(rules.Attacked); ok {
							r.blows = append(r.blows, goldenBlow{p.number[int(at.From)], p.number[int(at.To)], at.Hit, int(at.Damage)})
						}
					}
					if out.TurnOver {
						break
					}
				}
			}
			rs.EndTurn(e.fight(), id, &rng)
			if rs.Verdict(e.fight()) != rules.FightOn {
				break
			}
		}
		p.rounds = append(p.rounds, r)
		rs.EndRound(e.fight(), &rng)
	}
	p.beaten = map[int]bool{}
	for id, n := range p.units {
		switch st := rs.Sheet(id).Status; {
		case e.fact(id, rules.FactSide) != rules.SideMonsters:
			if st == rules.StatusOkay {
				p.upLeft++
			}
		case st != rules.StatusOkay && st != rules.StatusRunning:
			p.beaten[n] = true
		}
	}
	p.result = rs.EndFight(e.fight(), &rng)
	return p
}

// TestBR1GoldenFight compares the fight, in the order of what can be compared without a shared draw order:
//  1. the numbers each unit fights with at the first round's head (the order-free ones): armour, rear armour, to-hit,
//     movement, hit points, damage dice, and for each foe pair the chance to hit;
//  2. the draws: the dice's state at each round's head, then the order the units act in, then what each attack did,
//     and the hit points at the next head: the first round that differs is reported, the later ones are not
//     compared (they follow from it);
//  3. the end: experience each member gains, the credits, the loot;
//     and, over the rounds, the damage the units fact shows (the dice the next attack rolls) and the hit chance the facts gave
//     before each attack (the numbers the old engine rolled against).
func TestBR1GoldenFight(t *testing.T) {
	var cases []goldenFight
	readGolden(t, "fights.json", &cases)
	var places []goldenArena
	readGolden(t, "arenas.json", &places)
	at := map[string]goldenPlace{}
	for _, p := range places {
		at[p.Name] = p.goldenPlace
	}
	var groups map[string][][2]int
	readGolden(t, "groups.json", &groups)
	fixtures := goldenFixtures{}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			rounds := goldenRounds(c.Trace)
			head := rounds[0].Head
			var members []goldenUnit
			for _, u := range head.Units[:c.Members] {
				members = append(members, goldenUnit{Name: u.Name})
			}
			e := fixtures.ready(t, c.Members, members)
			e.goldenFight(t, at[c.Place], groups[c.Group], c.Approach, c.Dist)
			exp0 := map[rules.CharID]int64{}
			for _, id := range e.partyIDs()[:c.Members] {
				exp0[id] = e.ruleset().Sheet(id).Exp
			}
			p := e.goldenPlay(t, head.State) // the dice as they stood at the first round's head: the old engine drew before it (leadership)
			e.goldenCompareHead(t, c, head, p)
			known, isKnown := goldenDiverges[c.Name]
			upto := len(rounds) // the rounds whose damage and chances follow the old engine's: up to the divergence
			if isKnown {
				upto = known.round + 1
			}
			goldenCompareDamage(t, c, rounds[:min(upto, len(rounds))], p)
			goldenCompareChances(t, c, rounds[:min(upto, len(rounds))], p)
			round, why := goldenCompareRounds(t, c, rounds, p)
			switch {
			case isKnown && round == known.round:
				t.Logf("%s: equal to the old engine up to round %d; there %s (%s)", c.Name, round, why, known.why)
			case isKnown && round != known.round:
				t.Errorf("%s: listed as diverging in round %d, it diverges in round %d: %s", c.Name, known.round, round, why)
			case round >= 0:
				t.Errorf("%s round %d: %s", c.Name, round, why)
			}
			e.goldenCompareEnd(t, c, p, exp0)
		})
	}
}

// goldenDiceText is rules.Unit.Damage's text for the old engine's dice: NdS, a bonus with its sign.
func goldenDiceText(n, sides, bonus int) string {
	s := fmt.Sprintf("%dd%d", n, sides)
	if bonus != 0 {
		s += fmt.Sprintf("%+d", bonus)
	}
	return s
}

// goldenCompareDamage is the units fact's Damage at each round's head against the dice the old engine's next attack
// rolls: its slot 2's (the natural attacks, spent first) while it has attacks left there, else its slot 1's (the weapon's).
func goldenCompareDamage(t *testing.T, c goldenFight, want []goldenRound, p goldenPlayed) {
	t.Helper()
	for i, w := range want[:min(len(want), len(p.rounds))] {
		for _, u := range w.Head.Units {
			slot := 0
			if u.Attacks2 > 0 {
				slot = 1
			}
			text := goldenDiceText(u.DiceCount[slot], u.DiceSides[slot], u.DiceBonus[slot])
			got, on := p.rounds[i].damage[u.ID]
			if !on {
				continue // gone from the ground: fled
			}
			if got != text {
				t.Errorf("%s round %d unit %d %s: damage shown %q, old engine's next attack (slot %d) rolls %s", c.Name, i, u.ID, u.Name, got, slot+1, text)
			}
		}
	}
}

// goldenChance is the percent of the d20 that hits at the number needed: a natural 1 misses, a natural 20 hits.
func goldenChanceOf(need int) int {
	hits := 0
	for d := 2; d <= 20; d++ {
		if d == 20 || d >= need {
			hits++
		}
	}
	return hits * 5
}

// goldenCompareChances is the hit chance the facts gave just before each single-target attack, per attacker and
// target, against the numbers the old engine rolled against in that round for that pair. The old engine's rolls
// also hold the guards' blows and the area attacks, so each chance must be among the pair's, not the other way round.
func goldenCompareChances(t *testing.T, c goldenFight, want []goldenRound, p goldenPlayed) {
	t.Helper()
	checked := 0
	for i, w := range want[:min(len(want), len(p.rounds))] {
		for _, got := range p.rounds[i].chances {
			var seen []int
			for _, r := range w.Rolls {
				if r.Att == got.att && r.Tgt == got.tgt {
					seen = append(seen, goldenChanceOf(r.Need))
				}
			}
			if len(seen) == 0 {
				continue // only natural 1s were rolled (they are not traced)
			}
			checked++
			if !slices.Contains(seen, got.percent) {
				t.Errorf("%s round %d: unit %d at unit %d: chance shown %d%%, old engine rolled against %v%%", c.Name, i, got.att, got.tgt, got.percent, seen)
			}
		}
	}
	t.Logf("%s: %d hit chances compared", c.Name, checked)
}

// goldenCompareHead is 1: what the first round's head shows, order-free.
func (e *Engine) goldenCompareHead(t *testing.T, c goldenFight, head goldenTraceEv, p goldenPlayed) {
	t.Helper()
	first := p.rounds[0]
	for _, g := range head.Units {
		who := fmt.Sprintf("%s unit %d %s", c.Name, g.ID, g.Name)
		if hp, ok := first.hp[g.ID]; !ok || hp != g.HP {
			t.Errorf("%s: hit points %d, old engine %d", who, hp, g.HP)
		}
		if m := first.moves[g.ID]; m != g.Move {
			t.Errorf("%s: movement %d half squares, old engine %d", who, m, g.Move)
		}
	}
}

// goldenDiverges is, per case, the first round in which the shared engine's fight stops following the old engine's
// (the rounds before it are equal: the dice's state at the head, the order, the squares, the hit points, every
// attack's result), and why. A case not listed is equal throughout. The test fails when a case diverges earlier,
// and when one no longer diverges where it is listed: the entry then goes.
var goldenDiverges = map[string]struct {
	round int
	why   string
}{
	// two-vs-41x4+32+39x3: the two members, played by the computer, bandage the dying NPC beside them in round 0
	// (combatai.pas AiMoveAttack:471, a team member's FindPatients/TreatPatients, its turn over); the old engine
	// had them attack instead. The Pascal decides.
	"two-vs-41x4+32+39x3": {0, "the members bandage the dying NPC (combatai.pas AiMoveAttack:471) instead of attacking"},
} // every other fight is equal to the old engine (the golden branch's patches are in the README) to its end

// goldenCompareRounds is 2: it returns the first round that differs, -1 when all are equal.
func goldenCompareRounds(t *testing.T, c goldenFight, want []goldenRound, p goldenPlayed) (int, string) {
	t.Helper()
	for i, w := range want {
		if i >= len(p.rounds) {
			return i, fmt.Sprintf("the old engine played %d rounds, the shared engine %d", len(want), len(p.rounds))
		}
		g := p.rounds[i]
		for _, u := range w.Head.Units {
			if at, ok := g.at[u.ID]; ok && (at.X != u.X || at.Y != u.Y) && w.Head.Round > 0 {
				return i, fmt.Sprintf("unit %d %s stands at %d,%d at the head, old engine %d,%d", u.ID, u.Name, at.X, at.Y, u.X, u.Y)
			}
			if hp, ok := g.hp[u.ID]; ok && hp != u.HP && w.Head.Round > 0 {
				return i, fmt.Sprintf("unit %d has %d hit points at the head, old engine %d", u.ID, hp, u.HP)
			}
		}
		n := min(len(g.order), len(w.Head.Order))
		switch {
		case g.state != w.Head.State:
			return i, fmt.Sprintf("the dice stand at %d at the head, old engine %d", g.state, w.Head.State)
		case !slices.Equal(g.order[:n], w.Head.Order[:n]):
			return i, fmt.Sprintf("the units act in the order %v, old engine %v", g.order, w.Head.Order)
		case !slices.Equal(g.blows, w.Blows):
			return i, fmt.Sprintf("the attacks %+v, old engine %+v", g.blows, w.Blows)
		}
	}
	if len(p.rounds) != len(want) {
		return len(want), fmt.Sprintf("%d rounds, old engine %d", len(p.rounds), len(want))
	}
	return -1, ""
}

// goldenCompareEnd is 3: the end the shared engine's own fight came to (who was beaten), worth what the old engine's
// records say each monster is worth: experience over the members still up, credits, the gear in the order the
// old engine's share (pO05E_005B) heaps it (each piece in front of the rest).
func (e *Engine) goldenCompareEnd(t *testing.T, c goldenFight, p goldenPlayed, exp0 map[rules.CharID]int64) {
	t.Helper()
	var xp, credits int
	var gear []goldenItem
	for _, r := range c.Rewards {
		if !p.beaten[r.Unit] {
			continue
		}
		xp += r.Exp
		credits += r.Credits
		for _, it := range r.Items {
			gear = append([]goldenItem{it}, gear...)
		}
	}
	each := 0
	if p.upLeft > 0 {
		each = xp / p.upLeft
	}
	if int(p.result.Exp) != each {
		t.Errorf("%s: experience each %d, old engine's rule on the same beaten %d", c.Name, p.result.Exp, each)
	}
	if int(p.result.Money.Copper) != credits {
		t.Errorf("%s: credits %d, old engine's rule on the same beaten %d", c.Name, p.result.Money.Copper, credits)
	}
	if p.upLeft > 0 {
		for _, id := range e.partyIDs() {
			if _, ok := exp0[id]; ok && int(e.ruleset().Sheet(id).Exp-exp0[id]) != each && e.ruleset().Sheet(id).Status == rules.StatusOkay {
				t.Errorf("%s: %s gains %d experience, want %d", c.Name, e.ruleset().Sheet(id).Name, e.ruleset().Sheet(id).Exp-exp0[id], each)
			}
		}
	}
	var loot []int
	for _, h := range p.result.Items {
		v, err := e.ruleset().Fact(rules.NoChar, rules.ItemCardFact{Item: h})
		if err != nil {
			t.Fatal(err)
		}
		loot = append(loot, int(v.(rules.ItemCard).Type))
	}
	var want []int
	for _, it := range gear {
		want = append(want, it.Type)
	}
	if !slices.Equal(loot, want) {
		t.Errorf("%s: loot item types %v, old engine's rule on the same beaten %v", c.Name, loot, want)
	}
}
