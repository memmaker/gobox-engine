package rules4e

import "gobox/rules1e"

// def is a defence (PHB p. 274).
type def int

const (
	ac def = iota
	fort
	ref
	will
)

// defences are Fortitude, Reflex and Will less AC of the Monster Manual
// entries the kinds follow (Monster.MM).
var defences = map[string][3]int{
	"Bugbear Warrior":     {-1, -3, -4},
	"Decrepit Skeleton":   {-3, -2, -3},
	"Dire Boar":           {4, 0, -1},
	"Displacer Beast":     {-2, -1, -3},
	"Drider Fanglord":     {1, -1, -3},
	"Dryad":               {-1, -2, -2},
	"Efreet Fireblade":    {-1, -2, -4},
	"Ettin Marauder":      {-2, -10, -9},
	"Fire Giant":          {0, -6, -6},
	"Ghoul":               {-3, -1, -4},
	"Gnoll Marauder":      {0, -3, -3},
	"Goblin Cutter":       {-4, -2, -5},
	"Goblin Hexer":        {-3, -2, -1},
	"Goblin Warrior":      {-4, -2, -5},
	"Gray Wolf":           {-2, -2, -3},
	"Greenscale Hunter":   {-2, -3, -4},
	"Hill Giant":          {2, -4, -2},
	"Hippogriff":          {-1, -3, -5},
	"Hobgoblin Commander": {0, -3, -2},
	"Hobgoblin Grunt":     {-2, -4, -5},
	"Kobold Minion":       {-4, -2, -4},
	"Kobold Wyrmpriest":   {-4, -2, -2},
	"Medusa Archer":       {-3, -2, -1},
	"Minotaur Warrior":    {1, -5, -3},
	"Mummy Guardian":      {-2, -4, -3},
	"Ogre Savage":         {2, -3, -3},
	"Orc Chieftain":       {0, -3, -1},
	"Orc Drudge":          {-1, -4, -4},
	"Quickling Runner":    {-4, 0, -4},
	"Skeleton":            {-3, -2, -3},
	"Stirge":              {-3, -2, -5},
	"Stone-eye Basilisk":  {-2, -6, -6},
	"Stormclaw Scorpion":  {-2, -4, -5},
	"Troll":               {1, -2, -3},
	"Vampire Lord":        {1, -2, -4},
	"Wight":               {-1, -2, -3},
	"Wraith":              {-3, 0, -2},
	"Wyvern":              {0, -4, -5},
	"Zombie":              {0, -4, -3},
	"Zombie Rotter":       {0, -4, -3},
}

// roleDefences are the same by role, for a kind the book lacks: other
// defences level + 12, AC by role (DMG p. 184).
var roleDefences = map[Role]int{Skirmisher: -2, Brute: 0, Soldier: -4, Lurker: -2, Controller: -2, Artillery: 0}

// classDefences are the class bonuses to Fortitude, Reflex and Will (PHB ch. 4).
var classDefences = map[rules1e.Class][3]int{
	rules1e.Fighter: {2, 0, 0}, rules1e.Paladin: {1, 1, 1}, rules1e.Ranger: {1, 1, 0},
	rules1e.Cleric: {0, 0, 2}, rules1e.Thief: {0, 2, 0}, rules1e.Mage: {0, 0, 2},
}

// defence is b's defence d less its AC.
func defence(e *rules1e.RuleState, b *rules1e.Character, d def) int {
	if d == ac {
		return 0
	}
	if m, ok := Lookup(b.Name()); ok && b.Foe() {
		if o, ok := defences[m.MM]; ok {
			return o[d-1]
		}
		return roleDefences[m.Role]
	}
	// a character: 10 + half level + the better of two abilities + the
	// class's bonus, against AC 10 + half level + armour (PHB p. 274); the
	// armour is what the original's armour class is below 10, and the
	// half levels cancel
	pair := [3][2]rules1e.Ability{{rules1e.Strength, rules1e.Constitution}, {rules1e.Dexterity, rules1e.Intelligence}, {rules1e.Wisdom, rules1e.Charisma}}[d-1]
	o := max(e.Mod(b, pair[0]), e.Mod(b, pair[1]))
	cls := 0
	for k, bonus := range classDefences {
		if b.Level(k) > 0 {
			cls = max(cls, bonus[d-1]) // a multi-class character: its best class's
		}
	}
	return o + cls - (int(b.AC[0]) - 50) // AC[0]: 60 - armour class
}
