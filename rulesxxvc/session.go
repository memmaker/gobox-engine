package rulesxxvc

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"gobox/rules"
)

// The session: what the shared engine asks and writes by typed kinds from
// start-up to the first step (REFACTOR.md B1f1): the new game, the random
// stream, the clock, the party's count and acting member, joining, the sheet
// and card facts of a member. Where XXVc has no such thing the answer is the
// zero value of the kind's type, said so beside it.

var (
	errNoMember      = errors.New("rulesxxvc: no such character")
	errClockReadOnly = errors.New("rulesxxvc: the minute of the day is read only")
)

// The game clock (types.pas zone0 clock, gametime.pas): units from the smallest
// to years, clockUnitScale of each making one of the next (10, 10, 6, 24, 30, 12, 100).
const (
	clockSegment = iota
	clockMinute
	clockTurn // ten minutes
	clockHour
	clockDay
	clockMonth
	clockYear
	clockUnits
)

const (
	minutesPerTurn = 10 // clockUnitScale[clockMinute]
	minutesPerHour = 60 // minutesPerTurn * clockUnitScale[clockTurn]
)

// nonPlayerFlag is the NON_PLAYER_FLAG of the record's morale byte (BR r[231]):
// a hireling's control byte, which the engine reads as above 0x7F.
const nonPlayerFlag = 0x80

// levelColumns is the class columns Book.Levels has: the engine reads the first eight.
const levelColumns = 8

// daySkyColour is the sky colour of daytime (the cyan of O120:010A); skyColours
// maps the area's sky byte (low four bits) to its colour (d0D86).
const daySkyColour = 11

var skyColours = [16]int{0, 15, 4, 11, 13, 2, 9, 14, 0, 15, 4, 11, 13, 2, 9, 14}

// stream is the generator a roll draws from: the engine's, when the call
// brings one (a fight), else the one SetSeed seeded.
func (r *Rules) stream(rng *rules.RNG) *rules.RNG {
	if rng != nil {
		return rng
	}
	return &r.seed
}

// health is how fit a character is behind its status (Sheet.Health): the okay
// are hale, every other status cannot act.
func health(s Status) rules.Health {
	if s == StOkay || s == StNoStatus {
		return rules.Hale
	}
	return rules.Disabled
}

// minuteOfDay is the clock's time of day (the position line shows the hour, the
// ten-minute turn and the minute).
func (r *Rules) minuteOfDay() int {
	return r.clock[clockHour]*minutesPerHour + r.clock[clockTurn]*minutesPerTurn + r.clock[clockMinute]
}

// clockUnit is the clock unit a ClockPart names (ok false: none).
func clockUnit(p rules.ClockPart) (int, bool) {
	switch p {
	case rules.ClockTenMinutes:
		return clockTurn, true
	case rules.ClockHours:
		return clockHour, true
	}
	return 0, false
}

// clockUnitScale is how many of each clock unit make one of the next (gdconsts.pas clockUnitScale).
var clockUnitScale = [clockUnits]int{10, 10, 6, 24, 30, 12, 100}

// advanceClock is gametime.pas AdvanceTime (BR1 pO0E3_0388): count of unit pass, carrying up, and
// the effects of everyone run down by the minutes that is (pO0E3_0000: the fight's, else the party's).
// A year's carry ages the characters in the original; a Character keeps no age.
func (r *Rules) advanceClock(count, unit int) {
	if unit < 0 || unit >= clockUnits {
		return // pO0E3_0388: no such unit, no time
	}
	for range count {
		r.clock[unit]++
		for i := range r.clock {
			if r.clock[i] >= clockUnitScale[i] {
				r.clock[i] -= clockUnitScale[i]
				if i+1 < clockUnits {
					r.clock[i+1]++
				}
			}
		}
	}
	minutes := count
	for u := clockMinute; u < unit; u++ {
		minutes *= clockUnitScale[u]
	}
	who := r.party
	if r.fight != nil {
		who = slices.Collect(maps.Keys(r.fight))
	}
	r.Elapse(who, int32(minutes), nil)
}

// resolveSession resolves the session's Actions; ok false: not one of them.
func (r *Rules) resolveSession(a rules.Action) (o rules.Outcome, ok bool, err error) {
	switch a := a.(type) {
	case rules.InitRules:
		// No XXVc counterpart: 1e builds its spell table and affect handler table
		// here, XXVc's effects are code (effects.go) and it has no spells.
	case rules.LoadItemTable:
		// No XXVc counterpart: the item templates came with the tables at Begin.
	case rules.SetSeed:
		r.seed = rules.RNG{State: a.Seed}
	case rules.NewAdventure:
		// A new game's: the clock, the count, the acting member, the shared purse and
		// the ship, which the first ship fight loads again (record 50). The characters
		// and the party list stay, as the 1e rules keep theirs.
		r.clock, r.memberCount, r.acting, r.purse = [clockUnits]int{}, 0, 0, 0
		r.ship, r.foeSaved, r.space = nil, savedFoe{}, nil
	case rules.ActFor:
		r.acting = int(a.Member)
	case rules.SetMemberCount:
		r.memberCount = a.N
	case rules.SetClock:
		u, ok := clockUnit(a.Part)
		if !ok {
			return rules.Outcome{}, true, errClockReadOnly
		}
		r.clock[u] = int(a.Value)
	case rules.AdvanceClock:
		r.advanceClock(int(a.Count), int(a.Unit))
	case rules.JoinParty:
		if r.chars[a.Who] == nil {
			return rules.Outcome{}, true, errNoMember
		}
		r.party = append(r.party, a.Who) // pO0CA_194D: at the end, current, counted
		r.acting = len(r.party) - 1
		r.memberCount++
	case rules.EmptyParty:
		r.memberCount = 0 // the count only, as the 1e rules: the list is ForgetParty's
	case rules.ForgetParty:
		r.party = nil
	case rules.EmptyMember:
		if r.empty == rules.NoChar {
			r.empty = r.next
			r.next++
		}
		r.chars[r.empty] = &Character{Scores: map[rules.Ability]int{}}
		return rules.Outcome{Effects: []rules.Effect{rules.EmptyMemberIs{Who: r.empty}}}, true, nil
	case rules.SetRoster:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoMember
		}
		c.Roster = a.Roster
	case rules.Toughen:
		// No XXVc counterpart: BR has no Toughness rule; hit points stay as they are.
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// sessionCheck answers the session's Checks; the rest is the zero Answer.
func (r *Rules) sessionCheck(q rules.Check, rng *rules.RNG) rules.Answer {
	switch q := q.(type) {
	case rules.RollDice:
		return rules.Answer{Value: int32(rules.RNGDice(r.stream(rng), int(q.N), int(q.Sides)))}
	case rules.RollRandom:
		return rules.Answer{Value: int32(r.stream(rng).N(int(q.N)))}
	case rules.ClockIs:
		if q.Part == rules.ClockMinuteOfDay {
			return rules.Answer{Yes: true, Value: int32(r.minuteOfDay())}
		}
		if u, ok := clockUnit(q.Part); ok {
			return rules.Answer{Yes: true, Value: int32(r.clock[u])}
		}
	case rules.MemberCount:
		return rules.Answer{Yes: true, Value: r.memberCount}
	case rules.ActingFor:
		return rules.Answer{Yes: true, Value: int32(r.acting)}
	case rules.PartySize:
		return rules.Answer{Yes: true, Value: int32(len(r.party))}
	case rules.IsDay:
		// BR1 has no day by the clock (the sky the area sets is the day: SkyDay).
		return rules.Answer{}
	case rules.SkyDay:
		sky := q.Sky
		if q.Inside {
			sky = q.SkyInside
		}
		return rules.Answer{Yes: skyColours[sky&15] == daySkyColour}
	case rules.PoolCap, rules.PoolCount:
		// No shared inventory: each member carries its own (Options.PartyInventory is a 1e rule).
		return rules.Answer{}
	case rules.ItemTableMissing:
		return rules.Answer{Yes: len(r.tables.Items) == 0}
	case rules.CarryCount:
		if c := r.chars[q.Who]; c != nil {
			return rules.Answer{Value: int32(len(c.Items))}
		}
	case rules.Fit:
		c := r.chars[q.Who]
		return rules.Answer{Yes: c != nil && c.Status == StOkay && c.Enabled}
	case rules.Modifiable:
		// Not yet: modifying a character is creation (not ported).
		return rules.Answer{}
	case rules.PaladinCan:
		// No paladins.
		return rules.Answer{}
	case rules.InFight:
		return rules.Answer{Yes: r.fight != nil || r.space != nil}
	case rules.HallOpen:
		// No training hall in XXVc.
		return rules.Answer{}
	}
	return rules.Answer{}
}

// factSession answers the Facts the state build asks (ok false: not one of them).
func (r *Rules) factSession(id rules.CharID, f rules.Fact) (v rules.Value, ok bool, err error) {
	switch f := f.(type) {
	case rules.PartyFact:
		return rules.Party(slices.Clone(r.party)), true, nil
	case rules.PurseFact:
		return r.partyPurse(f.Loot), true, nil
	}
	c := r.chars[id]
	if c == nil {
		return nil, false, nil // Fact's own: no character
	}
	switch f := f.(type) {
	case rules.MemberCardFact:
		// No spells: nothing ready, no slots. The portrait is the engine's pool's.
		return rules.MemberCard{Party: c.team(), Class: string(c.Career)}, true, nil
	case rules.PoolPicsFact:
		return rules.PoolPics{}, true, nil // no pool pictures held
	case rules.PortraitFact:
		return rules.Portrait{}, true, nil // no portrait numbers held
	case rules.IconPicsFact:
		return rules.IconPics{}, true, nil // no combat icon held
	case rules.BookFact:
		// No spells: the empty book. Its one level column is the career's level.
		b := rules.Book{Levels: make([]int32, levelColumns)}
		b.Levels[0] = int32(c.Level)
		return b, true, nil
	case rules.SpellMenuFact:
		return rules.SpellMenu(nil), true, nil // no spells to cast
	case rules.OwedFact:
		return rules.PowerOwed{}, true, nil // no powers to pick
	case rules.FiguresFact:
		return rules.Figures(nil), true, nil // adds no lines to the sheet
	case rules.TrainFact:
		return rules.TrainCard{}, true, nil // no training: it never rises
	case rules.DamageFact:
		return rules.Text(diceText(c.Damage[0])), true, nil
	case rules.ListedFact:
		if f.Goods {
			return nil, false, nil
		}
		var h rules.Handles
		for _, ci := range c.Items {
			h = append(h, ci.Handle)
		}
		return h, true, nil
	case rules.ItemCardFact:
		card, err := r.itemCard(c, f.Item)
		return card, true, err
	case rules.WordFact:
		return r.word(c, f.Of), true, nil
	case rules.PlainFact:
		switch f {
		case rules.FactControl:
			ctl := c.Morale & (nonPlayerFlag - 1) // the low seven bits
			if c.NonPlayer {
				ctl |= nonPlayerFlag
			}
			return rules.Number(ctl), true, nil
		case rules.FactLevelReady:
			return rules.Yes(false), true, nil // levels are not trained
		}
	}
	return nil, false, nil
}

// partyPurse is the credits the party holds as a Purse (credits are Copper,
// rules.Purse's comment): the one purse under SharedPurse, else every member's
// own. The loot's is empty outside a fight's looting, whose coins the Situation holds.
func (r *Rules) partyPurse(loot bool) rules.Purse {
	if loot {
		return rules.Purse{}
	}
	total := r.purse
	if !r.options.SharedPurse {
		for _, id := range r.party {
			if c := r.chars[id]; c != nil {
				total += c.Coins
			}
		}
	}
	return rules.Purse{Copper: uint64(max(total, 0))}
}

// word is the sheet word of c's fact of: what XXVc keeps. It has no sex,
// alignment or exceptional strength (the zero word).
func (r *Rules) word(c *Character, of rules.PlainFact) rules.Text {
	switch of {
	case rules.FactStatus:
		return rules.Text(coreStatus[c.Status])
	case rules.FactRace:
		return rules.Text(c.Race)
	case rules.FactClasses:
		return rules.Text(c.Career)
	}
	return ""
}

// diceText is d as the sheet words it: "1d6+2".
func diceText(d rules.Dice) string {
	s := fmt.Sprintf("%dd%d", d.N, d.Sides)
	if d.Bonus != 0 {
		s += fmt.Sprintf("%+d", d.Bonus)
	}
	return s
}

// itemCard is the card of c's item h from its template and type (Item).
func (r *Rules) itemCard(c *Character, h rules.ItemHandle) (rules.ItemCard, error) {
	i := heldAt(c, h)
	if i < 0 {
		return rules.ItemCard{}, fmt.Errorf("ItemCard: no item %d", h)
	}
	ci := c.Items[i]
	it, err := r.Item(ci.Ref)
	if err != nil {
		return rules.ItemCard{}, err
	}
	t := r.tables.Items[ci.Ref.ID]
	return rules.ItemCard{Item: h, Name: strings.Join(it.NameWords, " "), Type: t.Type, Count: it.Count, Price: it.Price,
		Kind: ci.Ref.ID, Weight: t.Weight, Readied: ci.Readied, Identified: it.Identified, Tab: string(it.Kind)}, nil
}
