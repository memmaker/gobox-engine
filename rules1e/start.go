package rules1e

import "gobox/rules"

// A new character's per-adventure start (Felix's R3 rulings 3, 4, 7, 8, 11):
// one rule reads it, each game's importer writes its own values into
// rules.json ("creationStart"); POR's are in the POR literal.
type StartData struct {
	Money      StartMoney
	Experience []int32 // by class count (one class first); nil: none. The character trains silently up to it.
	MageSpells StartSpells
	Kit        [kitKinds][]Item // POD: each kit's items in the original's order, by kit kind (kitFighter...)
}

// StartMoney is what a new character's purse holds, in gold (the engine's
// coins are gold and copper only; an importer converts other coins by value):
// Fixed, plus each class's Dice summed times Times (divided by the class count
// when ByClass).
type StartMoney struct {
	Fixed   int32     // gold every new character gets
	Dice    [][]int16 // by class row: count, sides each class rolls; nil: none
	Times   int32     // the dice's sum times this
	ByClass bool      // the dice's total divided by the class count (POR)
}

// StartSpells are the magic-user spells a new magic-user knows, after
// training: Known always, every magic-user spell below AllBelow, BySlot's
// lists for each spell level the magic-user has a slot of (POD).
type StartSpells struct {
	Known    []SpellID
	AllBelow int32
	BySlot   map[int32][]SpellID // by spell level
	AllAt    int32               // a slot of this level: every magic-user spell (0: none)
}

// CoinGold is Character.Coins' gold slot.
const CoinGold = 3

// The kits (POD startingKits): who carries which.
const (
	kitFighter = iota // fighter, ranger, paladin
	kitThief
	kitCleric
	kitOther
	kitKinds
)

// kitOf is c's kit kind.
func kitOf(c *Character) int {
	switch {
	case int8(c.Levels[ClsFighter]) > 0 || int8(c.Levels[ClsRanger]) > 0 || int8(c.Levels[ClsPaladin]) > 0:
		return kitFighter
	case int8(c.Levels[ClsThief]) > 0:
		return kitThief
	case int8(c.Levels[ClsCleric]) > 0:
		return kitCleric
	}
	return kitOther
}

// copperPerGold is a gold piece in copper.
var copperPerGold = rules.GoldBoxRates.Coin[CoinGold]

// startMoney rolls c's starting money (before the hit points, as POR).
func (e *RuleState) startMoney(c *Character, count byte) {
	m := e.T().Start.Money
	var sum byte // a byte, as POR's
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) <= 0 || m.Dice == nil {
			continue
		}
		if d := m.Dice[classRow[PorClass[k]]]; d[0] > 0 {
			sum += byte(e.Dice(int(d[0]), int(d[1])))
		}
	}
	coins := int(sum) * int(m.Times)
	if m.ByClass {
		coins /= int(count)
	}
	c.Coins[CoinGold] = uint16(int16(coins + int(m.Fixed)))
}

// startExperience gives c its starting experience and trains it silently
// up to it; roll: the hit points of each level rolled (no: a rebuilt draft).
func (e *RuleState) startExperience(c *Character, count byte, roll bool) {
	s := e.T().Start
	if s.Experience == nil {
		return
	}
	c.Exp = uint32(s.Experience[count-1])
	for mask := e.readyMask(c); mask != 0; mask = e.readyMask(c) {
		e.raiseLevels(c, mask, roll)
	}
	e.startSpells(c)
}

// startSpells teaches a magic-user the start's spells.
func (e *RuleState) startSpells(c *Character) {
	if int8(c.Levels[ClsMage]) <= 0 {
		return
	}
	sp := e.T().Start.MageSpells
	know := func(ids []SpellID) {
		for _, s := range ids {
			c.SpellBook[s-1] = 1
		}
	}
	know(sp.Known)
	below := sp.AllBelow
	if sp.AllAt > 0 && c.SpellSlot(SpellClassMage, int(sp.AllAt)) > 0 {
		below = LastSpell // every level
	}
	for s := 1; s <= LastSpell; s++ {
		if rec := e.T().Spells[s][:2]; SpellClass(rec[0]) == SpellClassMage && int32(rec[1]) < below {
			c.SpellBook[s-1] = 1
		}
	}
	for lvl, ids := range sp.BySlot {
		if c.SpellSlot(SpellClassMage, int(lvl)) > 0 {
			know(ids)
		}
	}
}

// startKit gives c its kit, each item put first, as POD's startingEquipment.
func (e *RuleState) startKit(c *Character) error {
	kit := e.T().Start.Kit[kitOf(c)]
	if kit == nil {
		return nil
	}
	kit, err := e.gameItems(kit) // the kit is the game's
	if err != nil {
		return err
	}
	for i := range kit {
		c.Items = append([]*Item{&kit[i]}, c.Items...)
	}
	e.ApplyItems(c)
	return nil
}

// The default combat icon parts (sotsb/pod defaultIconParts, Felix's R3
// ruling 13): the head by sex and size (a halfling's its own, POD), the
// body (weapon pose) by class; the shared picker starts on them.
const (
	iconHeadHalfling                         = 3
	iconHeadFemaleLarge, iconHeadFemaleSmall = 9, 7
	iconHeadMaleLarge, iconHeadMaleSmall     = 5, 0
	iconBodyCleric, iconBodyRanger           = 23, 1
	iconBodyFighter, iconBodyMage            = 24, 29
	iconBodyThief                            = 5
)

// defaultIconParts gives c the default icon head and body.
func defaultIconParts(c *Character) {
	large := c.IconSize == 2
	switch {
	case c.Race == RaceHalfling:
		c.IconHead = iconHeadHalfling
	case c.Sex == 1 && large:
		c.IconHead = iconHeadFemaleLarge
	case c.Sex == 1:
		c.IconHead = iconHeadFemaleSmall
	case large:
		c.IconHead = iconHeadMaleLarge
	default:
		c.IconHead = iconHeadMaleSmall
	}
	switch {
	case int8(c.Levels[ClsCleric]) > 0:
		c.IconBody = iconBodyCleric
	case int8(c.Levels[ClsRanger]) > 0:
		c.IconBody = iconBodyRanger
	case int8(c.Levels[ClsPaladin]) > 0 || int8(c.Levels[ClsFighter]) > 0:
		c.IconBody = iconBodyFighter
	case int8(c.Levels[ClsMage]) > 0:
		c.IconBody = iconBodyMage
	default:
		c.IconBody = iconBodyThief
	}
}

// Modifiable: c has no experience of its own, i.e. exactly the adventure's
// starting experience for its class count (Start.Experience; none: 0), so a
// sequel's started character can still be modified (coordinator 01:30).
func (e *RuleState) Modifiable(c *Character) bool {
	s := e.T().Start
	if s.Experience == nil {
		return c.Exp == 0
	}
	count := 0
	for _, l := range c.Levels {
		if int8(l) > 0 {
			count++
		}
	}
	return count > 0 && count <= len(s.Experience) && c.Exp == uint32(s.Experience[count-1])
}
