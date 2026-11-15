package rulesxxvc

import (
	"encoding/json"
	"errors"
	"slices"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// Character is one character the ruleset holds, typed (TCharacter's rule
// fields; the byte record stays in the BR repos).
type Character struct {
	Name       string
	Race       rules.Race
	Career     rules.Class
	Level      int
	Scores     map[rules.Ability]int
	HP, MaxHP  int
	BaseCombat int
	Skills     SkillLevels
	Saves      [8]int
	Coins      rules.Copper // its own, under SharedPurse off; a monster's creditTotal
	Enabled    bool
	Status     Status
	Side       rules.Side // SideParty, SideAuto (the team) or SideMonsters (hostile)
	Exp        int64
	XPAward    int64 // what beating it gives (xpAward)

	// The combat numbers RecalcStats leaves (types.pas TCharacter).
	// ponytail: kept as set at creation or Spawn; RecalcStats from readied items is not ported.
	AttackBonus int
	ArmorClass  [2]int        // front, rear
	Damage      [2]rules.Dice // per attack phase
	BaseScale   int           // phase 1's attacks per round times two, unarmed
	Scale       int           // phase 2's
	Movement    int
	Items       []Carried

	// RecalcStats' inputs (TCharacter): the unarmed dice per phase, the
	// armour class and movement before items, the weapon specializations
	// (item types) and the carried weight it leaves.
	Unarmed         [2]rules.Dice
	BaseArmorClass  int
	BaseMovement    int
	Specializations [SpecializationSlots]int32
	Encumbrance     int
	Armor           ArmorFlags // armorFlags, from the readied items
	Effects         []Effect   // effectList, in order

	// A companion's morale (isCompanion; npc: NON_PLAYER_FLAG and the low
	// seven bits) for CheckMorale.
	Companion bool
	NonPlayer bool
	Morale    int

	Pilot bool // isPilot: picked to fly the party's ship
}

// team is C_Side.TEAM: the party's side, steered or not.
func (c *Character) team() bool { return c.Side != rules.SideMonsters }

// Rules is the XXVc ruleset of BR1 and BR2: core.Rules. The game's tables come
// at Begin (core.Campaign.Tables, an xxvc.Tables).
type Rules struct {
	loot      map[rules.ItemHandle]rules.ItemRef // FightResult.Items' handles (lootHandle)
	nextItem  rules.ItemHandle                   // the last carried item's handle (handOut)
	chars     map[rules.CharID]*Character
	order     []rules.CharID // join order: the first member takes the purse
	next      rules.CharID
	campaign  rules.Campaign
	options   rules.Options
	tables    xxvc.Tables
	purse     rules.Copper // the party's one purse under SharedPurse
	morale    int          // GDVars.morale: the last CheckMorale worked out
	hpPercent int          // GDVars.hostileHpPercent: set at BeginFight and EndRound
	moveBonus int          // zone1 movementBonus (FightTerms.PartyMove) for this fight
	tactics   [2]bool      // GDVars.tacticsSides by side (TEAM, HOSTILE): RollTactics each round

	initiative []rules.CharID // this round's order still to act
	fight      map[rules.CharID]*combatState
	arena      arena // the arena cells, clouds and corpses over Board.Terrain

	ship     *xxvc.ShipRecord // zone0 partyShip; nil until the first ship fight loads record 50
	foeSaved savedFoe         // zone0's last enemy ship
	space    *shipFight       // the ship fight going on
}

var _ rules.Rules = (*Rules)(nil)

// ErrNotPorted is what the parts not yet derived from the reconstructions
// answer (ONE-BINARY.md Track X lists them).
var ErrNotPorted = errors.New("rulesxxvc: not ported yet")

func New() *Rules { return &Rules{chars: map[rules.CharID]*Character{}, next: 1} }

func (r *Rules) Catalog() rules.Catalog {
	return rules.Catalog{
		Races:    xxvc.Races,
		Classes:  xxvc.Careers,
		Skills:   xxvc.Skills,
		Saves:    xxvc.Saves,
		Features: []rules.Feature{rules.FeatureSpaceCombat},
	}
}

// Begin takes the adventure's tables and the options. Turning the shared purse
// off gives the whole purse to the first member (Felix 2026-11-11); turning it
// on pools every member's coins.
func (r *Rules) Begin(c rules.Campaign, o rules.Options) {
	if t, ok := c.Tables.(xxvc.Tables); ok {
		r.tables = t
	}
	switch {
	case o.SharedPurse && !r.options.SharedPurse:
		for _, id := range r.order {
			r.purse += r.chars[id].Coins
			r.chars[id].Coins = 0
		}
	case !o.SharedPurse && r.options.SharedPurse && len(r.order) > 0:
		r.chars[r.order[0]].Coins += r.purse
		r.purse = 0
	}
	r.campaign, r.options = c, o
}

// Add keeps c and gives its id (creation's Make and Import come through here).
func (r *Rules) Add(c Character) rules.CharID {
	id := r.next
	r.next++
	if r.options.SharedPurse {
		r.purse += c.Coins
		c.Coins = 0
	}
	r.chars[id] = &c
	r.handOut(&c)
	r.order = append(r.order, id)
	return id
}

// Purse is the coins a character can spend: its own, or the party's one purse.
func (r *Rules) Purse(id rules.CharID) rules.Copper {
	if r.options.SharedPurse {
		return r.purse
	}
	if c := r.chars[id]; c != nil {
		return c.Coins
	}
	return 0
}

func (r *Rules) Dismiss(id rules.CharID) {
	delete(r.chars, id)
	r.order = slices.DeleteFunc(r.order, func(x rules.CharID) bool { return x == id })
}

func (r *Rules) Sheet(id rules.CharID) rules.Sheet {
	c := r.chars[id]
	if c == nil {
		return rules.Sheet{}
	}
	s := rules.Sheet{Name: c.Name, Race: c.Race, Levels: map[rules.Class]int32{c.Career: int32(c.Level)},
		HP: int32(c.HP), MaxHP: int32(c.MaxHP), Status: coreStatus[c.Status], Exp: c.Exp, Scores: map[rules.Ability]int32{},
		Saves: map[rules.Save]int32{}, Skills: map[rules.Skill]int32{}}
	for a, v := range c.Scores {
		s.Scores[a] = int32(v)
	}
	for i, v := range c.Saves {
		s.Saves[xxvc.Saves[i]] = int32(v)
	}
	for n := SkillNo(1); n <= SkillCount; n++ {
		if v := c.Skills.Level(n); v > 0 {
			s.Skills[xxvc.Skills[n-1]] = int32(v)
		}
	}
	return s
}

// skillNo is the number of a roster skill.
func skillNo(s rules.Skill) (SkillNo, bool) {
	i := slices.Index(xxvc.Skills, s)
	return SkillNo(i + 1), i >= 0
}

// skillValue is charhlp.pas SkillValue for c: the level plus its group's
// ability when the prerequisite is met.
func skillValue(c *Character, n SkillNo) int8 {
	ability := 0
	if a, ok := groupAbility[Group(n)]; ok {
		ability = c.Scores[a]
	}
	return SkillValue(c.Skills.Level(n), int8(ability), MeetsPrerequisite(&c.Skills, c.Career == xxvc.Medic, n), c.Enabled)
}

// groupAbility is the ability each group adds (SkillAbilityScore).
var groupAbility = map[SkillGroup]rules.Ability{
	GroupDexterity:    rules.Dexterity,
	GroupIntelligence: rules.Intelligence,
	GroupWisdom:       rules.Wisdom,
	GroupCharisma:     rules.Charisma,
	GroupTechnical:    xxvc.Tech,
}

// Check answers SkillRoll (eclvm.pas EclSkillCheck over SelectSkillCharacter).
func (r *Rules) Check(q rules.Check, rng *rules.RNG) rules.Answer {
	switch q := q.(type) {
	case rules.SkillRoll:
		n, ok := skillNo(q.Skill)
		if !ok {
			return rules.Answer{Grade: rules.FailedBadly, Who: rules.NoChar}
		}
		var values []int8
		var whom []rules.CharID
		for _, id := range q.Whom {
			if c := r.chars[id]; c != nil {
				values, whom = append(values, skillValue(c, n)), append(whom, id)
			}
		}
		g, i := SkillCheck(values, q.Hardness, rng)
		a := rules.Answer{Grade: g, Yes: g >= rules.Made, Who: rules.NoChar}
		if i >= 0 {
			a.Who = whom[i]
		}
		return a
	}
	return rules.Answer{}
}

// Save and Load are the characters and the purse, JSON.
type saved struct {
	Chars map[rules.CharID]*Character
	Order []rules.CharID
	Next  rules.CharID
	Purse rules.Copper
	Ship  *xxvc.ShipRecord `json:",omitempty"`
	Foe   savedFoe
}

func (r *Rules) Save() ([]byte, error) {
	return json.Marshal(saved{r.chars, r.order, r.next, r.purse, r.ship, r.foeSaved})
}

func (r *Rules) Load(b []byte) error {
	var s saved
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	r.chars, r.order, r.next, r.purse, r.ship, r.foeSaved = s.Chars, s.Order, s.Next, s.Purse, s.Ship, s.Foe
	if r.chars == nil {
		r.chars = map[rules.CharID]*Character{}
	}
	return nil
}

// Not ported yet: each answers ErrNotPorted or nothing (Track X lists them).
func (r *Rules) Draft() *rules.Draft                      { return nil }
func (r *Rules) Shared() []rules.Carried                  { return nil }
func (r *Rules) Spell(rules.SpellID) (rules.Spell, error) { return rules.Spell{}, ErrNotPorted }
