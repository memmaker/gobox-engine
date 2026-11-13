// Package rules4e is the "4e Combat" rule set (Esc menu, Gameplay): combat
// options in the manner of D&D 4th edition laid over the original's rules.
// It plugs in through Engine.HandIn and nothing else; to take it out,
// delete the package, its lines in cmd/por/main.go (setRules) and the
// Gameplay page's "4e Combat" (gobox-engine/blobber, Options.Combat4e).
//
// Monsters: each kind has its 4e role and rank (Monster Manual, matched
// by name; kinds the book lacks are placed by their abilities). A minion
// dies from any damage (DMG p. 55); a special attack recharges on 5-6
// (MM p. 6) instead of striking every time; the roles steer the
// computer's play (ai.go). Players: powers.go.
package rules4e

import (
	"gobox/rules1e"
	"strings"
)

// Role is how a monster fights (DMG pp. 54-55).
type Role int

const (
	Brute Role = iota
	Soldier
	Skirmisher
	Artillery
	Controller
	Lurker
)

// Rank is how much of a fight a monster is.
type Rank int

const (
	Standard Rank = iota
	Minion        // dies from any damage; four are one standard monster
	Elite         // two standard monsters
	Solo          // five standard monsters
)

// Monster is one kind's place in 4e.
type Monster struct {
	Role   Role
	Rank   Rank
	Leader bool   // buffs its allies (a subrole)
	MM     string // the Monster Manual entry it follows ("" when the book has none)
}

// kinds are the bestiary kinds (rules1e.BeastKind); variants are the
// MON*CHA names that differ from their kind (a leader among minions).
var (
	kinds = map[string]Monster{
		"Anhkheg":         {Artillery, Standard, false, ""},
		"Basilisk":        {Soldier, Standard, false, "Stone-eye Basilisk"},
		"Bugbear":         {Brute, Standard, false, "Bugbear Warrior"},
		"Centaur":         {Skirmisher, Standard, false, ""},
		"Displacer Beast": {Skirmisher, Standard, false, "Displacer Beast"},
		"Drider":          {Brute, Standard, false, "Drider Fanglord"},
		"Efreeti":         {Soldier, Standard, false, "Efreet Fireblade"},
		"Ettin":           {Soldier, Elite, false, "Ettin Marauder"},
		"Fire Giant":      {Soldier, Standard, false, "Fire Giant"},
		"Giant Frog":      {Skirmisher, Standard, false, ""},
		"Giant Lizard":    {Brute, Standard, false, ""},
		"Giant Mantis":    {Lurker, Standard, false, ""},
		"Giant Scorpion":  {Soldier, Standard, false, "Stormclaw Scorpion"},
		"Giant Snake":     {Controller, Standard, false, ""},
		"Ghoul":           {Soldier, Standard, false, "Ghoul"},
		"Gnoll":           {Brute, Standard, false, "Gnoll Marauder"},
		"Goblin":          {Skirmisher, Minion, false, "Goblin Cutter"},
		"Hill Giant":      {Brute, Standard, false, "Hill Giant"},
		"Hippogriff":      {Skirmisher, Standard, false, "Hippogriff"},
		"Hobgoblin":       {Soldier, Minion, false, "Hobgoblin Grunt"},
		"Kobold":          {Skirmisher, Minion, false, "Kobold Minion"},
		"Lizardman":       {Skirmisher, Standard, false, "Greenscale Hunter"},
		"Medusa":          {Controller, Elite, false, "Medusa Archer"},
		"Minotaur":        {Soldier, Standard, false, "Minotaur Warrior"},
		"Mummy":           {Brute, Standard, false, "Mummy Guardian"},
		"Nymph":           {Skirmisher, Standard, false, "Dryad"},
		"Ogre":            {Brute, Standard, false, "Ogre Savage"},
		"Orc":             {Brute, Minion, false, "Orc Drudge"},
		"Phase Spider":    {Skirmisher, Standard, false, ""},
		"Quickling":       {Skirmisher, Standard, false, "Quickling Runner"},
		"Skeleton":        {Soldier, Minion, false, "Decrepit Skeleton"},
		"Spectre":         {Lurker, Standard, false, ""},
		"Stirge":          {Lurker, Minion, false, "Stirge"},
		"Thri-Kreen":      {Skirmisher, Standard, false, ""},
		"Tiger":           {Skirmisher, Standard, false, ""},
		"Troll":           {Brute, Standard, false, "Troll"},
		"Vampire":         {Lurker, Elite, false, "Vampire Lord"},
		"Wardog":          {Skirmisher, Standard, false, "Gray Wolf"},
		"Wight":           {Skirmisher, Standard, false, "Wight"},
		"Wild Boar":       {Brute, Standard, false, "Dire Boar"},
		"Wraith":          {Lurker, Standard, false, "Wraith"},
		"Wyvern":          {Skirmisher, Standard, false, "Wyvern"},
		"Zombie":          {Brute, Minion, false, "Zombie Rotter"},
	}
	variants = map[string]Monster{
		"GOBLIN GUARD":   {Skirmisher, Standard, false, "Goblin Warrior"},
		"GOBLIN LEADER":  {Controller, Standard, true, "Goblin Hexer"},
		"HOBGOBLIN LDR":  {Soldier, Standard, true, "Hobgoblin Commander"},
		"KOBOLD LEADER":  {Artillery, Standard, true, "Kobold Wyrmpriest"},
		"ORC LEADER":     {Brute, Elite, true, "Orc Chieftain"},
		"GIANT SKELETON": {Soldier, Standard, false, "Skeleton"},
		"JUJU ZOMBIE":    {Brute, Standard, false, "Zombie"},
	}
)

// Lookup is the 4e place of a monster by its name; false for a kind the table lacks (a human foe).
func Lookup(name string) (Monster, bool) {
	if m, ok := variants[strings.ToUpper(strings.TrimSpace(name))]; ok {
		return m, true
	}
	m, ok := kinds[rules1e.BeastKind(name)]
	return m, ok
}

// Rules are the 4e combat rules. Use New.
type Rules struct {
	rules1e.ClassicRules
	spent     map[*rules1e.Character]bool               // special attacks used and not recharged
	used      map[*rules1e.Character]map[int]bool       // powers once a fight, used
	conds     map[*rules1e.Character]map[cond]timer     // timed conditions on it
	temp      map[*rules1e.Character]int                // temporary hit points (PHB p. 293)
	quarry    map[*rules1e.Character]*rules1e.Character // the ranger's quarry (Hunter's Quarry)
	quarried  map[*rules1e.Character]bool               // the ranger's quarry damage dealt this round
	hitRun    map[*rules1e.Character]*rules1e.Character // the foe that gets no free attack on it this round (Hit and Run)
	marks     map[*rules1e.Character]*rules1e.Character // the soldier or paladin that marked it (ai.go, Divine Challenge)
	chal      map[*rules1e.Character]*rules1e.Character // the paladin whose divine challenge it is under
	stung     map[*rules1e.Character]bool               // its divine challenge's radiant damage taken before the paladin's next turn
	engaged   map[*rules1e.Character]bool               // the paladin attacked its challenged foe this turn
	lance     map[*rules1e.Character]*rules1e.Character // the cleric whose ally's next attack on it gets +2 (Lance of Faith)
	turned    map[*rules1e.Character]map[int]bool       // powers once a turn, used this turn
	second    map[*rules1e.Character]bool               // its power's second attack is being made (Passing Attack)
	reacted   map[*rules1e.Character]bool               // its immediate action of the round is spent (react.go)
	riposte   map[*rules1e.Character]*rules1e.Character // the rogue whose Riposte Strike hit it
	bonus     map[*rules1e.Character]int                // to hit on the attack being made (Fox's Cunning)
	wandTry   map[*rules1e.Character]int                // the Wand of Accuracy's bonus on the attack being made, spent only if it makes the hit (OnHit)
	brand     map[*rules1e.Character]brand              // the ally a Righteous Brand on it helps
	shifted   map[*rules1e.Character]bool               // it shifted before its power's attack (Nimble Strike)
	acted     map[*rules1e.Character]bool               // it has had a turn in this fight (First Strike)
	sneaked   map[*rules1e.Character]bool               // its Sneak Attack of the round is dealt
	rounded   map[*rules1e.Character]map[int]bool       // powers once a round, used this round
	count     map[*rules1e.Character]map[int]int        // powers a few times a fight, the times used
	channeled map[*rules1e.Character]bool               // its Channel Divinity of the fight is used
	strength  map[*rules1e.Character]bool               // Divine Strength rides its next attack this turn
	lapsed    map[*rules1e.Character]bool               // its divine challenge lapsed: none this turn
	zones     []zone                                    // the powers' areas on the map (zones.go)
}

// cond is a timed condition.
type cond int

const (
	guard1   cond = iota // +1 defence (Priest's Shield)
	guard2               // +2 defence (Second Wind); power bonuses do not add up, the most counts
	slowed               // speed 2 (PHB p. 277)
	dazed                // one action a turn, no free attacks, combat advantage against it (PHB p. 277)
	marked               // a soldier's or paladin's mark (marks: whose)
	prone                // +2 to melee attacks on it, -2 to ranged ones; standing up is its move (PHB p. 277)
	weakened             // half damage dealt (PHB p. 277)
	feeble               // -2 to its attack rolls (Enfeebling Strike)
	glow                 // +2 to its attack rolls (Divine Glow)
	fear                 // n less to its attack rolls (Fearsome Smite)
	ward                 // n more to its AC (Shielding Smite)
	branded              // an ally's melee attacks on it get more (Righteous Brand: brand)
	fortune              // +1 to its next attack (Divine Fortune)
	immobile             // it cannot move (Turn Undead; PHB p. 277)
)

// brand is a Righteous Brand: whose melee attacks it helps, and by how much.
type brand struct {
	ally *rules1e.Character
	n    int
}

// timer ends a condition by its giver's turn: at its start ("until the
// start of your next turn") or at the end of the next one ("until the
// end of your next turn"; begun: that turn has started).
type timer struct {
	by         *rules1e.Character
	end, begun bool
	n          int // how much (fear, ward)
}

// New is a fresh rule set.
func New() *Rules {
	r := &Rules{}
	r.CombatStart(nil)
	return r
}

// CombatStart forgets the last fight.
func (r *Rules) CombatStart(*rules1e.RuleState) {
	r.spent, r.used, r.conds, r.temp = map[*rules1e.Character]bool{}, map[*rules1e.Character]map[int]bool{}, map[*rules1e.Character]map[cond]timer{}, map[*rules1e.Character]int{}
	r.quarry, r.quarried, r.hitRun, r.marks = map[*rules1e.Character]*rules1e.Character{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]*rules1e.Character{}, map[*rules1e.Character]*rules1e.Character{}
	r.chal, r.stung, r.engaged, r.lance, r.turned = map[*rules1e.Character]*rules1e.Character{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]*rules1e.Character{}, map[*rules1e.Character]map[int]bool{}
	r.second, r.reacted, r.riposte, r.bonus, r.wandTry = map[*rules1e.Character]bool{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]*rules1e.Character{}, map[*rules1e.Character]int{}, map[*rules1e.Character]int{}
	r.brand, r.shifted, r.acted, r.sneaked = map[*rules1e.Character]brand{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]bool{}
	r.rounded, r.count, r.channeled, r.strength = map[*rules1e.Character]map[int]bool{}, map[*rules1e.Character]map[int]int{}, map[*rules1e.Character]bool{}, map[*rules1e.Character]bool{}
	r.lapsed, r.zones = map[*rules1e.Character]bool{}, nil
}

// lay puts condition k on c, given by by in its turn: until the start of
// by's next turn, or with end until the end of it.
func (r *Rules) lay(c *rules1e.Character, k cond, by *rules1e.Character, end bool) {
	if r.conds[c] == nil {
		r.conds[c] = map[cond]timer{}
	}
	r.conds[c][k] = timer{by: by, end: end}
}

// layN is lay until the end of by's next turn with an amount; the larger
// stays (power bonuses do not add up).
func (r *Rules) layN(c *rules1e.Character, k cond, by *rules1e.Character, n int) {
	n = max(n, r.conds[c][k].n)
	r.lay(c, k, by, true)
	t := r.conds[c][k]
	t.n = n
	r.conds[c][k] = t
}

func (r *Rules) has(c *rules1e.Character, k cond) bool {
	_, ok := r.conds[c][k]
	return ok
}

// tick runs down the timers c gave at the start (start) or end of its turn.
func (r *Rules) tick(c *rules1e.Character, start bool) {
	for _, ks := range r.conds {
		for k, t := range ks {
			switch {
			case t.by != c:
			case start && !t.end:
				delete(ks, k)
			case start:
				t.begun = true
				ks[k] = t
			case t.begun:
				delete(ks, k)
			}
		}
	}
}

// RoundStart rolls the recharge of a spent special attack (5 or 6 on a
// d6) and starts the round's once-a-round quarry damage.
func (r *Rules) RoundStart(e *rules1e.RuleState, c *rules1e.Character) {
	if r.spent[c] && e.Dice(1, 6) >= 5 {
		delete(r.spent, c)
	}
	delete(r.quarried, c)
	delete(r.sneaked, c)
	delete(r.rounded, c)
	r.roleRound(c)
}

// TurnStart ends what c gave until its next turn and lays on c's own conditions.
func (r *Rules) TurnStart(e *rules1e.RuleState, c *rules1e.Character) {
	r.tick(c, true)
	r.zoneStart(e, c)
	if r.has(c, slowed) {
		c.Combat.Moves = min(c.Combat.Moves, 4) // half squares: 2 squares
	}
	if r.has(c, immobile) {
		c.Combat.Moves = 0
	}
	if r.has(c, prone) {
		delete(r.conds[c], prone)
		c.Combat.Moves = 0 // standing up: the move action
	}
	for t, p := range r.chal {
		if p == c {
			delete(r.stung, t)
		}
	}
	delete(r.reacted, c)
	delete(r.shifted, c)
	r.acted[c] = true
	for t, p := range r.riposte {
		if p == c {
			delete(r.riposte, t)
		}
	}
}

// TurnEnd ends what c gave until the end of this turn, its Hit and Run,
// and its divine challenge when it did not engage the foe.
func (r *Rules) TurnEnd(e *rules1e.RuleState, c *rules1e.Character) {
	r.tick(c, false)
	delete(r.hitRun, c)
	delete(r.turned, c)
	delete(r.strength, c)
	r.zoneEnd(c)
	delete(r.lapsed, c)
	for t, p := range r.chal { // a divine challenge lasts while the paladin engages its foe
		if p != c {
			continue
		}
		if r.engaged[c] || e.Distance(c, t) == 1 {
			r.lay(t, marked, c, true)
		} else {
			delete(r.chal, t)
			r.lapsed[c] = true // and none on its next turn
		}
	}
	delete(r.engaged, c)
}

// Moved: a dazed c's step is its one action.
func (r *Rules) Moved(_ *rules1e.RuleState, c *rules1e.Character) {
	if r.has(c, dazed) {
		c.Attacks = [2]rules1e.Count{}
		c.Combat.CanCast, c.Combat.CanUse = 0, 0
	}
}

// Special lets a foe's special attack strike once, then waits for its recharge.
func (r *Rules) Special(e *rules1e.RuleState, c *rules1e.Character) bool {
	if !c.Foe() || !e.HasSpecial(c) {
		return true
	}
	if r.spent[c] {
		return false
	}
	r.spent[c] = true
	return true
}

// Damage takes temporary hit points first; any damage destroys a minion.
func (r *Rules) Damage(_ *rules1e.RuleState, c *rules1e.Character, dmg int) int {
	if t := min(r.temp[c], max(dmg, 0)); t > 0 {
		r.temp[c] -= t
		dmg -= t
	}
	if m, ok := Lookup(c.Name()); ok && m.Rank == Minion && c.Foe() && dmg > 0 {
		return int(c.HP) + 10 // past -9: dead, not dying
	}
	return dmg
}
