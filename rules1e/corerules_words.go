package rules1e

import "gobox/rules"

// wordPtr is the record word w names, nil when it is not a plain record
// word (core.WordIs, core.SetWord).
func (r *Rules) wordPtr(w rules.WorldWord) *uint16 {
	a, g := &r.S.Area, &r.S.Game
	switch w {
	case rules.WordMapX:
		return &a.MapX
	case rules.WordMapY:
		return &a.MapY
	case rules.WordGeo:
		return &a.Geo
	case rules.WordTenMinutes:
		return &a.Clock[2]
	case rules.WordHours:
		return &a.Clock[3]
	case rules.WordNoMagic:
		return &a.NoMagic
	case rules.WordIn3D:
		return &a.In3D
	case rules.WordWallSplitA:
		return &a.WallSplitA
	case rules.WordWallSplitB:
		return &a.WallSplitB
	case rules.WordPrevX:
		return &a.PrevX
	case rules.WordPrevY:
		return &a.PrevY
	case rules.WordAreaKind, rules.WordLastBlock: // POR keeps the block NewECL left in the area's kind word
		return &a.AreaKind
	case rules.WordNoCoords:
		return &a.NoCoords
	case rules.WordMsgDelay:
		return &a.MsgDelay
	case rules.WordViewFlags:
		return &a.ViewFlags
	case rules.WordMarkFlags4:
		return &a.MarkFlags4
	case rules.WordMark3:
		return &a.Mark3
	case rules.WordMark4Off:
		return &a.Mark4Off
	case rules.WordMarkFlags57:
		return &a.MarkFlags57
	case rules.WordRiverOpen:
		return &a.RiverOpen
	case rules.WordBoss:
		return &a.Boss
	case rules.WordMark4:
		return &a.Mark4
	case rules.WordSky:
		return &a.Sky
	case rules.WordSkyInside:
		return &a.SkyInside
	case rules.WordTrainable:
		return &g.PartyExp
	case rules.WordMaxDist:
		return &g.MaxDist
	case rules.WordMonsterDist:
		return &g.MonsterDist
	case rules.WordMorale:
		return &g.Morale
	case rules.WordCombatEnd:
		return &g.CombatEnd
	case rules.WordBeaten:
		return &g.Beaten
	case rules.WordHoldMove:
		return &g.HoldMove
	case rules.WordSearch:
		return &g.Search
	case rules.WordSurprise:
		return &g.Surprise
	case rules.WordRestEvery:
		return &g.RestEvery
	case rules.WordRestChance:
		return &g.RestChance
	case rules.WordLeftMap:
		return &g.LeftMap
	case rules.WordPortrait:
		return &g.Portrait
	case rules.WordTemple:
		return &g.Temple
	case rules.WordNoLoot:
		return &g.NoLoot
	case rules.WordDuel:
		return &g.Duel
	case rules.WordAttackedAlly:
		return &g.AttackedAlly
	case rules.WordShop:
		return &g.Shop
	case rules.WordMonsterHit:
		return &g.MonsterHit
	case rules.WordPartyHit:
		return &g.PartyHit
	case rules.WordPartyMove:
		return &g.PartyMove
	}
	return nil
}

// word answers core.WordIs.
func (r *Rules) word(k rules.WordIs) rules.Answer {
	flag := func(b bool) rules.Answer {
		if b {
			return rules.Answer{Yes: true, Value: 1}
		}
		return rules.Answer{Yes: true}
	}
	switch k.Word {
	case rules.WordPartyCount:
		return rules.Answer{Yes: true, Value: int32(r.S.PartyCount())}
	case rules.WordWho:
		return rules.Answer{Yes: true, Value: int32(r.S.Who)}
	case rules.WordBeginFlag:
		return rules.Answer{Yes: true, Value: int32(r.S.BeginFlag)}
	case rules.WordSymbolSet:
		return rules.Answer{Yes: true, Value: int32(r.S.Game.SymbolSet)}
	case rules.WordDebug:
		return flag(r.S.Debug)
	case rules.WordTraining:
		return flag(r.S.Training)
	case rules.WordCampOn:
		return flag(r.S.Camp.On)
	case rules.WordClockMinutes:
		return rules.Answer{Yes: true, Value: int32(r.S.ClockMinutes())}
	case rules.WordEffect:
		v, ok := r.S.EffectTable(k.Index)
		return rules.Answer{Yes: ok, Value: int32(v)}
	}
	if p := r.wordPtr(k.Word); p != nil {
		return rules.Answer{Yes: true, Value: int32(*p)}
	}
	return rules.Answer{}
}

// setWord resolves core.SetWord: a record word keeps the value's low bits.
func (r *Rules) setWord(a rules.SetWord) error {
	switch a.Word {
	case rules.WordPartyCount:
		r.S.Game.PartyCount = uint16(a.Value)
	case rules.WordSymbolSet:
		r.S.Game.SymbolSet = uint16(a.Value)
	case rules.WordWho:
		r.S.Who = int(a.Value)
	case rules.WordBeginFlag:
		r.S.BeginFlag = byte(a.Value)
	case rules.WordDebug:
		r.S.Debug = a.Value != 0
	case rules.WordTraining:
		r.S.Training = a.Value != 0
	case rules.WordCampOn:
		r.S.Camp.On = a.Value != 0
	default:
		p := r.wordPtr(a.Word)
		if p == nil {
			return rules.ErrUnknownWord
		}
		*p = uint16(a.Value)
	}
	return nil
}
