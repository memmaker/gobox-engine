package engine

import (
	"gobox/adventure"
	"gobox/rules"
)

// pairLen is the bytes of a record's score pair: the own score, then the one in use (pod stats[]).
const pairLen = 2

// memberFields are the selected member's fields that read what the original's SpecialReadZ1 reads
// of the record (eclmenu.pas of each game); its SpecialWriteZ1 has no case for them, so a script's
// write is a bug (it panics) unless the field has a set.
func (e *Engine) memberFields() map[string]engineField {
	m := map[string]engineField{
		// cotab $15 stats[INTELLIGENCE].current (its SpecialWriteZ1 has no case; no cotab script writes it); BR Party[19], which setChar stores (BR2 area 80)
		"SelInt":      e.selNumber(rules.AbilityFact{Ability: rules.Intelligence}),
		"SelCha":      e.facts(func(int) rules.Fact { return rules.AbilityFact{Ability: rules.Charisma} }), // sotsb $19 chaCur
		"SelStats":    e.facts(func(i int) rules.Fact { return scoreAt(rules.Abilities[i/pairLen], i) }),   // pod $10.. stats[0..4]: (base, cur) by C_Stat
		"SelCharisma": e.facts(func(i int) rules.Fact { return scoreAt(rules.Charisma, i) }),               // pod stats[CHARISMA]: (base, cur)
		// pod $B2 hpMax, sotsb $76 hpMaxV1; pod $1FC hpCur (SpecialWriteZ1 stores the byte), sotsb $119 hpCurrent (likewise)
		"SelHpMax": e.selNumber(rules.FactMaxHP),
		"SelHpCur": e.selNumber(rules.FactHP),
		// the gender and class bytes: the record's own numbers (genderNameTable, C_Class), the place in the ruleset's lists
		"SelGender": {get: func(int) (int64, bool) { return int64(selIndex(e, rules.FactSex, adventure.SexOrder)), true }},
		"SelCharClass": {get: func(int) (int64, bool) {
			return int64(selIndex(e, rules.FactClasses, e.ruleset().Catalog().ClassSets)), true
		}},
		// pod classLevels[0..6], oldLevels[0..6]: the columns of the ruleset's classes
		"SelClassLevels": e.facts(func(i int) rules.Fact { return rules.OwnLevelFact{Class: e.ruleset().Catalog().Classes[i]} }),
		"SelOldLevels":   e.facts(func(i int) rules.Fact { return rules.FormerLevelFact{Class: e.ruleset().Catalog().Classes[i]} }),
		// cotab $C9, sotsb $D0: the class's level and the class left behind's while that one is active
		"SelMageLevel":   e.facts(func(int) rules.Fact { return rules.LevelFact{Class: rules.ClassMage} }),
		"SelRangerLevel": e.facts(func(int) rules.Fact { return rules.LevelFact{Class: rules.ClassRanger} }),
		"SelGems": {get: func(int) (int64, bool) { return e.selFact(rules.PiecesFact{}), true }, // sotsb $C5 coinsA[GEMS]: the party's gems, as SelCoins' element of them
			set: func(_ int, v int64) { e.setSelFact(rules.PiecesFact{}, int64(uint16(v))) }},
		// por $F2: SpecialWriteZ1 has no case for it, the window is not read
		"SelWindowF2": fact(nil, nil),
	}
	// BR Party[77..160] (charVar r[76+s]): the level of skill s, the ruleset's skills in their catalog's order
	m["SelSkills"] = e.facts(func(i int) rules.Fact { return rules.SkillFact{Skill: e.ruleset().Catalog().Skills[i]} })
	for name, skill := range thiefFields {
		m[name] = e.facts(func(int) rules.Fact { return rules.SkillFact{Skill: skill} })
	}
	return m
}

// thiefFields are the thief skill fields not bound by name elsewhere (thiefSkills[2], [4], [5], [6]
// of pod's record, rules.Skill by the order of a record's thief bytes).
var thiefFields = map[string]rules.Skill{
	"SelOpenLocks": rules.OpenLocks, "SelMoveSilently": rules.MoveQuietly, "SelHideShadows": rules.Hide,
	"SelHearNoise": rules.Hear, "SelReadLanguages": rules.ReadLanguages,
}

// facts is a read-only field whose element i is the number of the selected member's Fact f(i).
func (e *Engine) facts(f func(i int) rules.Fact) engineField {
	return engineField{get: func(i int) (int64, bool) { return e.selFact(f(i)), true }}
}

// scoreAt is the score of ability a that byte i of a pair table reads: even bytes the own score, odd the one in use.
func scoreAt(a rules.Ability, i int) rules.Fact {
	if i%pairLen == 0 {
		return rules.OwnScoreFact{Ability: a}
	}
	return rules.AbilityFact{Ability: a}
}
