package rules1e

import "gobox/rules"

// wordPtr is the record word w names, nil when it is not a plain record
// word (core.WordIs, core.SetWord).
func (r *Rules) wordPtr(w rules.WorldWord) *uint16 {
	a := &r.S.Area
	switch w {
	case rules.WordTenMinutes:
		return &a.Clock[2]
	case rules.WordHours:
		return &a.Clock[3]
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
	case rules.WordWho:
		r.S.Who = int(a.Value)
	case rules.WordBeginFlag:
		r.S.BeginFlag = byte(a.Value)
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
