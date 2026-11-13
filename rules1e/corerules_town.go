package rules1e

import (
	"gobox/rules"
)

// The town: the training hall, temple, shop and purse (core/town.go).

// checkTown answers the Checks of the town topic.
func (r *Rules) checkTown(k rules.Check) (rules.Answer, bool) {
	s := r.S
	switch k.(type) {
	case rules.TrainFee:
		return rules.Answer{Value: s.T().Town.TrainPrice}, true
	case rules.BundlePrices:
		return rules.Answer{Yes: s.T().Town.Bundles}, true
	case rules.HallOpen:
		a := rules.Answer{Yes: s.Training}
		if s.Learning {
			a.Value = 1
		}
		return a, true
	default:
		return rules.Answer{}, false
	}
}

// resolveTown answers the Actions of the town topic.
func (r *Rules) resolveTown(a rules.Action) (rules.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case rules.SetPurseGold:
		s.Purse.Gold = a.Gold
	case rules.OpenHall:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		if _, _, best := s.TrainCheck(c); best > 0 {
			c.Exp = uint32(best) // as the original, before its question and refusals
		}
		s.Training, s.Learning = true, false
		return rules.Outcome{}, true, nil
	case rules.GiveTempleCure:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		if sv, ok := s.templeService(int(a.Cure)); ok && sv.Kind.Wound() && !s.CureNeeded(c, int(a.Cure)) {
			return rules.Outcome{}, true, errWoundNotNeeded
		}
		s.TempleCure(c, int(a.Cure))
		return rules.Outcome{}, true, nil
	case rules.CloseHall:
		s.Training, s.Learning = false, false
		return rules.Outcome{}, true, nil
	case rules.EnterShop:
		s.BeginFlag = 1
		return rules.Outcome{}, true, nil
	case rules.StockShop:
		for _, it := range s.Treasure {
			s.ItemName(nil, it, false)
			if it.Price == 0 { // ovr07:0068, the Buy list's: here at the door, so the goods published before Buy show it too (BUG-CLASSES 68)
				it.Price = 1
			}
		}
		s.Emit(rules.Sounded{Cue: rules.SoundCue("store_enter")})
		return rules.Outcome{}, true, nil
	case rules.Train:
		c := r.chars[a.Who]
		if c == nil {
			return rules.Outcome{}, true, errNoChar
		}
		mask, why, _ := s.TrainCheck(c)
		if why != "" {
			return rules.Outcome{Effects: []rules.Effect{
				rules.Refused{Who: a.Who, Why: rules.NotTrainable}, rules.Narrated{Text: why},
			}}, true, rules.ErrNotNow
		}
		s.TrainLevels(c, mask)
		s.Learning = len(s.Learnable(c)) > 0 && s.Learning
		if !s.Learning {
			s.Training = false
		}
		return rules.Outcome{}, true, nil
	case rules.Pay:
		if !s.Pay(a.Gold) { // ovr05:00BF
			return rules.Outcome{Effects: []rules.Effect{rules.Refused{Why: rules.NoMoney}}}, true, rules.ErrNotNow
		}
		return rules.Outcome{}, true, nil
	default:
		return rules.Outcome{}, false, nil
	}
	return rules.Outcome{}, true, nil
}

// factTown answers the Facts of the town topic.
func (r *Rules) factTown(id rules.CharID, f rules.Fact) (rules.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case rules.CoinFact:
		if c == nil {
			return nil, true, errNoChar
		}
		if v, ok := purseCoin(&s.Purse, f.Coin); ok {
			return rules.Number(min(*v, 0xFFFF)), true, nil
		}
		return nil, true, errNoFact
	case rules.PiecesFact:
		return rules.Number(min(*pursePieces(&s.Purse, f.Jewel), 0xFFFF)), true, nil
	case rules.PurseFact:
		p := s.Purse
		if f.Loot {
			p = s.Loot
		}
		return rules.Purse{Gold: p.Gold, Copper: p.Copper, Gems: p.Gems, Jewelry: p.Jewelry}, true, nil
	case rules.TrainFact:
		if c == nil {
			return nil, true, errNoChar
		}
		mask, why, best := s.TrainCheck(c)
		return rules.TrainCard{Why: why, Becomes: s.TrainsTo(c, mask), Best: best, Rises: why == "" && mask != 0}, true, nil
	case rules.TempleCureFact:
		sv, ok := s.templeService(f.Cure)
		if !ok {
			return nil, true, errNoFact
		}
		return rules.TempleCure{Name: sv.Name, Needed: c != nil && s.CureNeeded(c, f.Cure), Price: sv.Price, Not: sv.Not, Done: sv.Done, Declines: sv.Declines, Only: sv.Kind.Wound()}, true, nil
	default:
		return nil, false, nil
	}
}
