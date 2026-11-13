package rules1e

import (
	"gobox/core"
)

// The town: the training hall, temple, shop and purse (core/town.go).

// checkTown answers the Checks of the town topic.
func (r *Rules) checkTown(k core.Check) (core.Answer, bool) {
	s := r.S
	switch k.(type) {
	case core.TrainFee:
		return core.Answer{Value: s.T().Town.TrainPrice}, true
	case core.BundlePrices:
		return core.Answer{Yes: s.T().Town.Bundles}, true
	case core.HallOpen:
		a := core.Answer{Yes: s.Training}
		if s.Learning {
			a.Value = 1
		}
		return a, true
	default:
		return core.Answer{}, false
	}
}

// resolveTown answers the Actions of the town topic.
func (r *Rules) resolveTown(a core.Action) (core.Outcome, bool, error) {
	s := r.S
	switch a := a.(type) {
	case core.SetPurseGold:
		s.Purse.Gold = a.Gold
	case core.OpenHall:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		if _, _, best := s.TrainCheck(c); best > 0 {
			c.Exp = uint32(best) // as the original, before its question and refusals
		}
		s.Training, s.Learning = true, false
		return core.Outcome{}, true, nil
	case core.GiveTempleCure:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		if sv, ok := s.templeService(int(a.Cure)); ok && sv.Kind.Wound() && !s.CureNeeded(c, int(a.Cure)) {
			return core.Outcome{}, true, errWoundNotNeeded
		}
		s.TempleCure(c, int(a.Cure))
		return core.Outcome{}, true, nil
	case core.CloseHall:
		s.Training, s.Learning = false, false
		return core.Outcome{}, true, nil
	case core.EnterShop:
		s.BeginFlag = 1
		return core.Outcome{}, true, nil
	case core.StockShop:
		for _, it := range s.Treasure {
			s.ItemName(nil, it, false)
			if it.Price == 0 { // ovr07:0068, the Buy list's: here at the door, so the goods published before Buy show it too (BUG-CLASSES 68)
				it.Price = 1
			}
		}
		s.Emit(core.Sounded{Cue: core.SoundCue("store_enter")})
		return core.Outcome{}, true, nil
	case core.Train:
		c := r.chars[a.Who]
		if c == nil {
			return core.Outcome{}, true, errNoChar
		}
		mask, why, _ := s.TrainCheck(c)
		if why != "" {
			return core.Outcome{Effects: []core.Effect{
				core.Refused{Who: a.Who, Why: core.NotTrainable}, core.Narrated{Text: why},
			}}, true, core.ErrNotNow
		}
		s.TrainLevels(c, mask)
		s.Learning = len(s.Learnable(c)) > 0 && s.Learning
		if !s.Learning {
			s.Training = false
		}
		return core.Outcome{}, true, nil
	case core.Pay:
		if !s.Pay(a.Gold) { // ovr05:00BF
			return core.Outcome{Effects: []core.Effect{core.Refused{Why: core.NoMoney}}}, true, core.ErrNotNow
		}
		return core.Outcome{}, true, nil
	default:
		return core.Outcome{}, false, nil
	}
	return core.Outcome{}, true, nil
}

// factTown answers the Facts of the town topic.
func (r *Rules) factTown(id core.CharID, f core.Fact) (core.Value, bool, error) {
	s, c := r.S, r.chars[id]
	switch f := f.(type) {
	case core.CoinFact:
		if c == nil {
			return nil, true, errNoChar
		}
		if v, ok := purseCoin(&s.Purse, f.Coin); ok {
			return core.Number(min(*v, 0xFFFF)), true, nil
		}
		return nil, true, errNoFact
	case core.PiecesFact:
		return core.Number(min(*pursePieces(&s.Purse, f.Jewel), 0xFFFF)), true, nil
	case core.PurseFact:
		p := s.Purse
		if f.Loot {
			p = s.Loot
		}
		return core.Purse{Gold: p.Gold, Copper: p.Copper, Gems: p.Gems, Jewelry: p.Jewelry}, true, nil
	case core.TrainFact:
		if c == nil {
			return nil, true, errNoChar
		}
		mask, why, best := s.TrainCheck(c)
		return core.TrainCard{Why: why, Becomes: s.TrainsTo(c, mask), Best: best, Rises: why == "" && mask != 0}, true, nil
	case core.TempleCureFact:
		sv, ok := s.templeService(f.Cure)
		if !ok {
			return nil, true, errNoFact
		}
		return core.TempleCure{Name: sv.Name, Needed: c != nil && s.CureNeeded(c, f.Cure), Price: sv.Price, Not: sv.Not, Done: sv.Done, Declines: sv.Declines, Only: sv.Kind.Wound()}, true, nil
	default:
		return nil, false, nil
	}
}
