package rulesxxvc

import (
	"strconv"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// The training hall (chargen.pas TrainingMenu, the party menu's Train Character while zone1's trainingMask is not 0):
// a conscious member whose experience reaches the next level's, below the maximum level, asks "Do you wish to train?"
// and gains a level: the career's base combat of the new level, AddLevelHitPoints, a weapon specialization for a
// warrior at an odd level (SpecializationMenu(True)), the experience held just under the level after, EditSkills(40,
// 20, 15), and every third level one off each saving throw. The hall costs nothing.

// The hall's words (TrainingMenu's PrintAndPause).
const (
	hallUnconscious = "we only train conscious people"
	hallMaxLevel    = "maximum level in game has been reached"
	hallNoExp       = "Not enough experience"
)

const (
	maxSpecialization = 3 // viewchar.pas MAX_SPECIALIZATION: a weapon is trained at most that often
	saveLevelStep     = 3 // TrainingMenu: level mod 3 = 0 takes one off each saving throw
)

// trainWhy is why the hall will not train c ("": it will).
func (r *Rules) trainWhy(c *Character) string {
	cr := r.tables.Creation
	switch {
	case c.Status != StOkay:
		return hallUnconscious
	case int32(c.Level) >= cr.MaxLevel:
		return hallMaxLevel
	case cr.Experience(c.Career, int32(c.Level)+1) > c.Exp:
		return hallNoExp
	}
	return ""
}

// trainCard is the hall's offer to c: what it becomes (TrainingMenu: "a level N career").
func (r *Rules) trainCard(c *Character) rules.TrainCard {
	why := r.trainWhy(c)
	card := rules.TrainCard{Why: why, Rises: why == ""}
	if why == "" {
		card.Becomes = []string{"a level " + strconv.Itoa(c.Level+1) + " " + xxvc.CareerNames[c.Career]}
	}
	return card
}

// resolveHall answers the hall's Actions (ok false: not one of them).
func (r *Rules) resolveHall(a rules.Action, rng *rules.RNG) (rules.Outcome, bool, error) {
	switch a := a.(type) {
	case rules.OpenHall:
		if r.chars[a.Who] == nil {
			return rules.Outcome{}, true, ErrNoMember
		}
		r.hall = true
	case rules.CloseHall:
		r.hall = false
	case rules.Train:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, ErrNoMember
		}
		if why := r.trainWhy(c); why != "" {
			return rules.Outcome{Effects: []rules.Effect{rules.Refused{Who: a.Who, Why: rules.NotTrainable}, rules.Narrated{Text: why}}}, true, rules.ErrNotNow
		}
		r.rise(c, rng)
		r.hall = false // TrainingMenu returns to the party menu
		return rules.Outcome{Effects: []rules.Effect{rules.PicksOwed{Who: a.Who}}}, true, nil
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// rise is TrainingMenu's level: the base combat, the hit points, the experience held under the next level, the saves.
// The picks (weapon, skills) follow in the draft Rise.
func (r *Rules) rise(c *Character, rng *rules.RNG) {
	cr := r.tables.Creation
	c.Level++
	c.BaseCombat = int(cr.Combat(c.Career, int32(c.Level)))
	gain := levelHitPoints(cr, c.Career, c.Level, c.Scores[rules.Constitution], r.stream(rng))
	c.MaxHP += gain
	c.HP += gain
	if next := cr.Experience(c.Career, int32(c.Level)+1); next > 0 && next <= c.Exp { // a level at a time
		c.Exp = next - 1
	}
	if c.Level%saveLevelStep == 0 {
		for i := range c.Saves {
			c.Saves[i]--
		}
	}
	r.RecalcStats(c)
}

// specializedCount is how many of the member's weapon specializations name the item type of weapon.
func (r *Rules) specializedCount(id rules.CharID, weapon rules.ItemID) int {
	c := r.chars[id]
	if c == nil {
		return 0
	}
	return c.Specialized(r.tables.Items[weapon].Type)
}

// risePicks puts the picks of a level gained on c: the skill levels, and the weapon a warrior trained
// (SpecializationMenu: the first free slot, none if the weapon is at its most).
func (r *Rules) risePicks(c *Character, d rules.Draft) error {
	if (creation{r: r}).owesWeapon(d) {
		opts := (creation{r: r}).specializable()
		if o := d.Picks[fieldSpecialization]; o >= 0 && o < len(opts) && r.specializedCount(d.Of, opts[o]) < maxSpecialization {
			for i, t := range c.Specializations {
				if t == 0 {
					c.Specializations[i] = r.tables.Items[opts[o]].Type
					break
				}
			}
		}
	}
	c.Skills = levelsOfDraft(d)
	r.RecalcStats(c)
	return nil
}
