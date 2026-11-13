package rules1e

import "gobox/core"

// wordPtr is the record word w names, nil when it is not a plain record
// word (core.WordIs, core.SetWord).
func (r *Rules) wordPtr(w core.WorldWord) *uint16 {
	a, g := &r.S.Area, &r.S.Game
	switch w {
	case core.WordMapX:
		return &a.MapX
	case core.WordMapY:
		return &a.MapY
	case core.WordGeo:
		return &a.Geo
	case core.WordTenMinutes:
		return &a.Clock[2]
	case core.WordHours:
		return &a.Clock[3]
	case core.WordNoMagic:
		return &a.NoMagic
	case core.WordIn3D:
		return &a.In3D
	case core.WordWallSplitA:
		return &a.WallSplitA
	case core.WordWallSplitB:
		return &a.WallSplitB
	case core.WordPrevX:
		return &a.PrevX
	case core.WordPrevY:
		return &a.PrevY
	case core.WordAreaKind, core.WordLastBlock: // POR keeps the block NewECL left in the area's kind word
		return &a.AreaKind
	case core.WordNoCoords:
		return &a.NoCoords
	case core.WordMsgDelay:
		return &a.MsgDelay
	case core.WordViewFlags:
		return &a.ViewFlags
	case core.WordMarkFlags4:
		return &a.MarkFlags4
	case core.WordMark3:
		return &a.Mark3
	case core.WordMark4Off:
		return &a.Mark4Off
	case core.WordMarkFlags57:
		return &a.MarkFlags57
	case core.WordRiverOpen:
		return &a.RiverOpen
	case core.WordBoss:
		return &a.Boss
	case core.WordMark4:
		return &a.Mark4
	case core.WordSky:
		return &a.Sky
	case core.WordSkyInside:
		return &a.SkyInside
	case core.WordTrainable:
		return &g.PartyExp
	case core.WordMaxDist:
		return &g.MaxDist
	case core.WordMonsterDist:
		return &g.MonsterDist
	case core.WordMorale:
		return &g.Morale
	case core.WordCombatEnd:
		return &g.CombatEnd
	case core.WordBeaten:
		return &g.Beaten
	case core.WordHoldMove:
		return &g.HoldMove
	case core.WordSearch:
		return &g.Search
	case core.WordSurprise:
		return &g.Surprise
	case core.WordRestEvery:
		return &g.RestEvery
	case core.WordRestChance:
		return &g.RestChance
	case core.WordLeftMap:
		return &g.LeftMap
	case core.WordPortrait:
		return &g.Portrait
	case core.WordTemple:
		return &g.Temple
	case core.WordNoLoot:
		return &g.NoLoot
	case core.WordDuel:
		return &g.Duel
	case core.WordAttackedAlly:
		return &g.AttackedAlly
	case core.WordShop:
		return &g.Shop
	case core.WordMonsterHit:
		return &g.MonsterHit
	case core.WordPartyHit:
		return &g.PartyHit
	case core.WordPartyMove:
		return &g.PartyMove
	}
	return nil
}

// word answers core.WordIs.
func (r *Rules) word(k core.WordIs) core.Answer {
	flag := func(b bool) core.Answer {
		if b {
			return core.Answer{Yes: true, Value: 1}
		}
		return core.Answer{Yes: true}
	}
	switch k.Word {
	case core.WordPartyCount:
		return core.Answer{Yes: true, Value: int32(r.S.PartyCount())}
	case core.WordWho:
		return core.Answer{Yes: true, Value: int32(r.S.Who)}
	case core.WordBeginFlag:
		return core.Answer{Yes: true, Value: int32(r.S.BeginFlag)}
	case core.WordSymbolSet:
		return core.Answer{Yes: true, Value: int32(r.S.Game.SymbolSet)}
	case core.WordDebug:
		return flag(r.S.Debug)
	case core.WordTraining:
		return flag(r.S.Training)
	case core.WordCampOn:
		return flag(r.S.Camp.On)
	case core.WordClockMinutes:
		return core.Answer{Yes: true, Value: int32(r.S.ClockMinutes())}
	case core.WordEffect:
		v, ok := r.S.EffectTable(k.Index)
		return core.Answer{Yes: ok, Value: int32(v)}
	}
	if p := r.wordPtr(k.Word); p != nil {
		return core.Answer{Yes: true, Value: int32(*p)}
	}
	return core.Answer{}
}

// setWord resolves core.SetWord: a record word keeps the value's low bits.
func (r *Rules) setWord(a core.SetWord) error {
	switch a.Word {
	case core.WordPartyCount:
		r.S.Game.PartyCount = uint16(a.Value)
	case core.WordSymbolSet:
		r.S.Game.SymbolSet = uint16(a.Value)
	case core.WordWho:
		r.S.Who = int(a.Value)
	case core.WordBeginFlag:
		r.S.BeginFlag = byte(a.Value)
	case core.WordDebug:
		r.S.Debug = a.Value != 0
	case core.WordTraining:
		r.S.Training = a.Value != 0
	case core.WordCampOn:
		r.S.Camp.On = a.Value != 0
	default:
		p := r.wordPtr(a.Word)
		if p == nil {
			return core.ErrUnknownWord
		}
		*p = uint16(a.Value)
	}
	return nil
}
