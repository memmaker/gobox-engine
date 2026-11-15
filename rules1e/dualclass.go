package rules1e

import (
	"slices"
	"strconv"

	"gobox/rules"
)

// Human dual-classing (extra, after Secret of the Silver Blades' "Human
// Change Classes", sotsb-go/engine/classchange.go). In the training hall a
// human who has never changed class may take up another: the levels of the
// class left are kept in OldLevels and the character starts again at level
// 1 with no experience. The old class stays dormant (its Levels entry 0)
// until the new class passes its level; then its levels come back and both
// classes count, though only the new one trains and gains hit points.

// DualClassed: c left the class with OldLevels set (dormant or back).
func DualClassed(c *Character) bool {
	for _, l := range c.OldLevels {
		if l > 0 {
			return true
		}
	}
	return false
}

// ChangeClass = sotsb O164:0F49: c leaves its class for class at level 1.
func (e *RuleState) ChangeClass(c *Character, class int) {
	old := int(c.Class)
	c.OldLevels[old], c.Levels[old], c.Levels[class] = c.Levels[old], 0, 1
	c.Class, c.Exp, c.MaxLevel = ClassID(class), 0, 1
	c.SpellSlots, c.Memorized, c.Plan = SpellSlots{}, [MemorySlots]SpellID{}, [MemorySlots]SpellID{}
	switch class {
	case ClsCleric:
		c.SpellSlots[SpellClassCleric][0] = 1
		e.wisdomSlots(c)
		for s := 1; s <= LastSpell; s++ {
			if rec := e.T().Spells[s][:2]; SpellClass(rec[0]) == SpellClassCleric && rec[1] == 1 {
				c.SpellBook[s-1] = 1
			}
		}
	case ClsMage:
		c.SpellSlots[SpellClassMage][0] = 1
		c.SpellBook[10], c.SpellBook[17], c.SpellBook[18], c.SpellBook[20] = 1, 1, 1, 1
	}
	if old == ClsPaladin {
		e.RemoveAffect(c, AffProtEvil, nil)
	}
	if class == ClsPaladin {
		e.AddAffect(c, AffProtEvil, 0, 0xFF, 0)
	}
	e.dualRecompute(c)
	for _, it := range c.Items { // what the new class may not use comes off
		if it.Readied != 0 && e.Base(it.Type).Classes&c.ClassMask == 0 {
			it.Readied = 0
		}
	}
	e.ApplyItems(c)
	e.narrate(c.Name() + " is now a 1st level " + ClassNames[class] + ".")
}

// dualRecompute: the derived values and the class mask from the levels.
func (e *RuleState) dualRecompute(c *Character) {
	e.RecomputeCharacter(c)
	c.ClassMask = 0
	for k := 0; k <= 7; k++ {
		if int8(c.Levels[k]) > 0 {
			c.ClassMask += byte(e.T().ClassMask[k])
		}
	}
}

// Dormant: class i of c is the one c left; it neither trains nor gains
// hit points, even once it is back.
func Dormant(c *Character, i int) bool { return c.OldLevels[i] > 0 }

// dualTrained: after a level in c's new class, the old class comes back
// when the new one has passed it. hp tells whether the level gives hit
// points (only above the old class's level).
func (e *RuleState) dualTrained(c *Character) (hp bool) {
	if !DualClassed(c) {
		return true
	}
	lv := c.Levels[c.Class]
	for i, old := range c.OldLevels {
		if old > 0 && c.Levels[i] == 0 && lv > old {
			c.Levels[i] = old
			e.dualRecompute(c)
			if i == ClsPaladin {
				e.AddAffect(c, AffProtEvil, 0, 0xFF, 0)
			}
			e.narrate(c.Name() + " may use the " + ClassNames[i] + "'s " + strconv.Itoa(int(old)) + " levels again.")
		}
		if old > 0 && lv <= old {
			return false
		}
	}
	return true
}

// Ability scores a human needs to change class: every prime requisite
// (class minimum above 8) of the old class above 14, of the new class above 16.
const (
	primeRequisite = 9
	oldClassPrime  = 14
	newClassPrime  = 16
)

// NewClasses are the classes c may change to: none unless c is a human of
// one class who never changed before.
func NewClasses(c *Character) []int {
	cur := int(c.Class)
	if c.Race != RaceHuman || cur > ClsMonk || DualClassed(c) || c.Control > 0x7F {
		return nil
	}
	var out []int
	for _, b := range RaceCareers(RaceHuman) {
		cl := int(b)
		if cl == cur || cl > ClsMonk || !mayTakeClass(c, cur, cl) {
			continue
		}
		out = append(out, cl)
	}
	return out
}

// mayTakeClass = sotsb O164:0D97: the old class's prime requisites above
// 14, the new one's above 16, and an alignment the new class allows.
func mayTakeClass(c *Character, cur, class int) bool {
	for k, v := range c.Stats {
		if CareerMinimum(Career(cur), Ability(k)) >= primeRequisite && v <= oldClassPrime ||
			CareerMinimum(Career(class), Ability(k)) >= primeRequisite && v <= newClassPrime {
			return false
		}
	}
	return slices.Contains(CareerAlignments(Career(class)), AlignmentNumber(c.Alignment))
}

// classChangeNote is what ChangeClass to class costs c, said before it is taken.
func classChangeNote(c *Character, class int) string {
	old, lv := ClassNames[c.Class], int(c.Levels[c.Class])
	return c.Name() + " becomes a 1st level " + ClassNames[class] + " with 0 experience. " +
		old + " level " + strconv.Itoa(lv) + " is lost until " + ClassNames[class] + " level " + strconv.Itoa(lv+1) + " and never rises again. " +
		"Memorized spells are lost. This is final."
}

// DualClassLine is the sheet's line on the class c left ("": none): its
// levels, and whether they are dormant or back.
func DualClassLine(c *Character) string {
	for i, old := range c.OldLevels {
		if !Dormant(c, i) {
			continue
		}
		s := ClassNames[i] + " " + strconv.Itoa(int(old))
		if c.Levels[i] == 0 {
			return s + ", dormant until " + ClassNames[c.Class] + " " + strconv.Itoa(int(old)+1)
		}
		return s + ", in use again, never advances"
	}
	return ""
}

// classOffers: in a hall that allows it, a ChangeClass for every party
// member and class it may change to, named by the class.
func (r *Rules) classOffers(t rules.Training) rules.Allowed {
	var out rules.Allowed
	if !t.DualClass {
		return out
	}
	for _, id := range t.Party {
		c := r.chars[id]
		if c == nil {
			continue
		}
		for _, cl := range NewClasses(c) {
			out.Offers = append(out.Offers, rules.Offer{Action: rules.ChangeClass{Who: id, To: PorClass[cl]}, Name: ClassNames[cl], Note: classChangeNote(c, cl)})
		}
	}
	return out
}

// changeClass resolves a ChangeClass the member may take (classOffers).
func (r *Rules) changeClass(a rules.ChangeClass) error {
	c := r.chars[a.Who]
	if c == nil {
		return errNoChar
	}
	cl := slices.Index(PorClass[:], a.To)
	if cl < 0 || !slices.Contains(NewClasses(c), cl) {
		return ErrNotNow
	}
	if i := slices.Index(r.S.Party, c); i >= 0 {
		r.S.Who = i
	}
	r.S.ChangeClass(c, cl)
	return nil
}
