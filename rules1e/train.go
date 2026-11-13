package rules1e

import (
	"fmt"
	"gobox/core"
)

// ovr17 (second half): Modify, Train and Add on the party screen, the
// icon copy the designer uses (icondesign.go), and the hit point helpers
// they share.

// PartyExp = es:[game+0x550]: the classes this training hall teaches
// (a mask that is ANDed with DS:5DE; the party menu only offers Train
// when it is set).
func (e *RuleState) PartyExp() int { return int(e.Game.PartyExp) }

// conHPBonus = ovr17:3F31 (char): hit point bonus per class from
// constitution (DS:408E), plus one or two for a class 2 character with
// constitution above 16 and 17. Returned as a signed byte.
func (e *RuleState) conHPBonus(c *Character) int {
	var sum byte
	for i := 0; i <= 7; i++ {
		if int8(c.Levels[i]) <= 0 || Dormant(c, i) {
			continue
		}
		adj := ConAdj(c, c.Stats[StatCon])
		if i == ClsRanger && c.Levels[i] == 1 {
			adj *= 2 // the ranger's two first level dice (CotAB ovr018)
		}
		sum += byte(adj)
	}
	return int(int8(sum))
}

// ConAdj is conHPBonus's adjustment per hit die at constitution con.
// extra: below 3 (Raise Dead takes a point each time) it is 3's, -2, and
// above 20 it is 20's; the original read past the table there (0 and +8,
// or -35 and -71).
func ConAdj(c *Character, con Score) int8 {
	con = min(max(con, 3), 20)
	adj := int8(ConHP(int(con)))
	if con > 0x10 && c.Class >= ClsFighter && c.Class <= ClsRanger {
		adj++
	}
	if con > 0x11 && c.Class >= ClsFighter && c.Class <= ClsRanger {
		adj++
	}
	return adj
}

// hitPointsFrom = the tail of ovr17:3FB3/408C: sum plus the constitution
// bonus, averaged over the classes (count 0 would be a divide error in
// the original; 0 here).
func (e *RuleState) hitPointsFrom(c *Character, sum byte, count int) byte {
	if count == 0 {
		return 0
	}
	adj := e.conHPBonus(c)
	if adj < 0 {
		if int(sum) > count+(-adj) {
			return byte((int(sum) + adj) / count)
		}
		return 1
	}
	return byte((int(sum) + adj) / count)
}

// MinHitPoints = ovr17:3FB3 (char): the least hit points the levels allow.
func (e *RuleState) MinHitPoints(c *Character) byte {
	var sum byte
	count := 0
	for i := 0; i <= 7; i++ {
		if lv := int8(c.Levels[i]); lv > 0 {
			sum += byte(lv)
			count++
			if i == 4 || i == 7 {
				sum++
			}
		}
	}
	return e.hitPointsFrom(c, sum, count)
}

// MaxHitPoints = ovr17:408C (char): the most hit points the levels allow
// (a full hit die per level, the die of each class).
func (e *RuleState) MaxHitPoints(c *Character) byte {
	var sum byte
	count := 0
	for i := 0; i <= 7; i++ {
		lv := int(int8(c.Levels[i]))
		if lv <= 0 {
			continue
		}
		count++
		switch i {
		case 0, 1:
			sum += byte(lv << 3)
		case 2, 3:
			sum += byte(lv * 10)
		case 4:
			sum += byte((lv + 1) << 3)
		case 5:
			sum += byte(lv << 2)
		case 6:
			sum += byte(lv * 6)
		case 7:
			sum += byte((lv + 1) << 2)
		}
	}
	return e.hitPointsFrom(e.Acting(), sum, count) // the original reads the bonus of the selected character, not c
}

// rollHitDice = ovr17:4235 (char, mask): hit dice rolled for the classes in
// mask (DS:5DE bits); a first level die is at least two thirds of the die.
func (e *RuleState) rollHitDice(c *Character, mask int) int {
	var total byte
	for i := 0; i <= 7; i++ {
		if int8(c.Levels[i]) <= 0 || byte(e.T().HpClassMask[i])&byte(mask) == 0 {
			continue
		}
		n, sides := e.T().TrainingDie(PorClass[i])
		if i == ClsRanger && c.Levels[i] > 1 {
			n = 1 // two dice at first level only (CotAB hp_calc)
		}
		r := e.Dice(n, sides) // ovr25:0DE3
		if c.Levels[i] == 1 && r < (sides<<1)/3 {
			r = (sides << 1) / 3
		}
		total += byte(r)
	}
	return int(total)
}

// LevelBlocked: the race's level limit stops class i at level lv.
func LevelBlocked(race RaceID, str Score, i int, lv Level) bool {
	// Level limits of the races.
	skip := false
	switch race {
	case 1:
		skip = i == 2 && (lv == 8 && str == 0x11 || lv == 7 && str < 0x11)
	case 2:
		skip = i == 2 && (lv == 7 || lv == 6 && str == 0x11 || lv == 5 && str < 0x11)
	case 3:
		skip = i == 2 && (lv == 6 || lv == 5 && str < 0x12)
	case 4:
		if i == 0 && lv == 5 {
			skip = true
		} else if i == 2 || i == ClsRanger { // the half-elf ranger as the fighter (CotAB)
			skip = lv == 8 || lv == 7 && str == 0x11 || lv == 6 && str < 0x11
		}
	case 5:
		skip = i == 2 && (lv == 6 || lv == 5 && str == 0x11 || lv == 4 && str < 0x11)
	}
	return skip
}

// levelReady (extra): c has the experience for a level in one of its
// classes; the roster shows the name in purple.
func levelReady(c *Character) bool {
	for i := 0; i <= 7; i++ {
		lv := c.Levels[i]
		if int8(lv) <= 0 || Dormant(c, i) || LevelBlocked(c.Race, c.Stats[StatStr], i, lv) {
			continue
		}
		if need := XPForLevel(PorClass[i], int(int8(lv+1))); need > 0 && need <= int32(c.Exp) {
			return true
		}
	}
	return false
}

// TrainCheck = the tests of ovr17:29AC on c: the classes it would rise
// in (DS:5DE bits, the hall's game +0x550 among them), why not when the
// hall refuses, and the experience to hold it at below a second level
// (0: none).
func (e *RuleState) TrainCheck(c *Character) (mask byte, why string, best int32) {
	if c.Status != StatusOkay && !e.Debug {
		return 0, "we only train conscious people", 0
	}
	if fee := e.T().Town.TrainPrice; e.Purse.Gold < uint64(fee) && !e.Debug {
		return 0, fmt.Sprintf("Training costs %d gp.", fee), 0
	}
	var have, can byte
	race, str := c.Race, c.Stats[StatStr]
	for i := 0; i <= 7; i++ {
		lv := c.Levels[i]
		if int8(lv) <= 0 || Dormant(c, i) {
			continue
		}
		have += byte(e.T().HpClassMask[i])
		if LevelBlocked(race, str, i, lv) {
			continue
		}
		exp := int32(c.Exp)
		need := e.T().XPForLevel(PorClass[i], int(int8(lv+1)))
		if need > exp && !e.Debug || need <= 0 {
			continue
		}
		can += byte(e.T().HpClassMask[i])
		next := e.T().XPForLevel(PorClass[i], int(int8(lv+2)))
		if next <= 0 || exp < next {
			continue
		}
		// Enough for two levels: the experience is held below the second.
		if next > best {
			best = next - 1
		}
	}
	if e.Debug {
		return can, "", best
	}
	hall := e.PartyExp()
	switch {
	case byte(int(have)&hall) == 0:
		return 0, "We don't train that class here", best
	case byte(int(can)&hall) == 0:
		return 0, "Not Enough Experience", best
	}
	return byte(int(can) & hall), "", best
}

// Learnable = SpellList kind 4 (ovr20:299C): the spells c has slots for
// and not in the book.
func (e *RuleState) Learnable(c *Character) []SpellID {
	var s []SpellID
	for i := 1; i <= LastSpell; i++ {
		if c.SpellSlot(e.SpellClass(SpellID(i)), int(e.SpellRec(SpellID(i), 1))) > 0 && c.SpellBook[i-1] == 0 {
			s = append(s, SpellID(i))
		}
	}
	return s
}

// TrainLevels = the training of ovr17:29AC: a level in each class of
// mask for the fee, and the hit points; a mage level gained sets
// learning (the original's spell pick, ovr20:299C, came before the hit
// points; it rolls no dice).
func (e *RuleState) TrainLevels(c *Character, mask byte) {
	e.Emit(core.Sounded{Cue: core.SoundCue("level")})
	e.narrate("Congratulations...")
	if !e.Debug {
		e.Purse.Gold -= uint64(e.T().Town.TrainPrice)
	}
	mage := int8(c.Levels[ClsMage])
	e.raiseLevels(c, mask, true)
	e.Learning = int8(c.Levels[ClsMage]) > mage // (the original, with nothing to choose, set +0x32, the max HP, to 1)
}

// raiseLevels is a level in each class of mask, and with roll its hit
// points: the training without the hall (TrainLevels, and the silent
// training of a new character's starting experience).
func (e *RuleState) raiseLevels(c *Character, mask byte, roll bool) {
	count := 0
	for i := 0; i <= 7; i++ {
		if int8(c.Levels[i]) <= 0 || Dormant(c, i) {
			continue
		}
		count++
		if byte(e.T().HpClassMask[i])&mask > 0 {
			c.Levels[i]++
		}
	}
	e.RecomputeCharacter(c)         // ovr24:0000
	if !roll || !e.dualTrained(c) { // extra: a dual-classed human gains no hit points until the new class passes the old
		return
	}
	hits := e.rollHitDice(c, int(mask))
	per := hits / count
	if per == 0 {
		per = 1
	}
	c.HPRoll += byte(per)
	per = (hits + e.conHPBonus(c)) / count
	if per < 1 {
		per = 1
	}
	damage := c.MaxHP - c.HP
	c.MaxHP += byte(per)
	c.HP = c.MaxHP - damage
}

// readyMask is the classes (DS:5DE bits) c has the experience to rise in,
// the race's limits allowing: the silent training's (sotsb/pod trainSilently).
func (e *RuleState) readyMask(c *Character) (mask byte) {
	for i := 0; i <= 7; i++ {
		lv := c.Levels[i]
		if int8(lv) <= 0 || Dormant(c, i) || LevelBlocked(c.Race, c.Stats[StatStr], i, lv) {
			continue
		}
		if need := e.T().XPForLevel(PorClass[i], int(int8(lv+1))); need > 0 && need <= int32(c.Exp) {
			mask += byte(e.T().HpClassMask[i])
		}
	}
	return mask
}

// JoinRefused (CotAB ovr018): paladins do not join the evil, the evil do not
// join a paladin, and a party takes at most three rangers.
func (e *RuleState) JoinRefused(c *Character) bool {
	evil := func(c *Character) bool { return (c.Alignment+1)%3 == 0 }
	msg, rangers := "", 0
	for _, m := range e.Party {
		switch {
		case int8(c.Levels[ClsPaladin]) > 0 && evil(m):
			msg = "paladins do not join with evil scum"
		case evil(c) && int8(m.Levels[ClsPaladin]) > 0:
			msg = m.Name() + " will tolerate no evil!"
		}
		if int8(m.Levels[ClsRanger]) > 0 {
			rangers++
		}
	}
	if msg == "" && int8(c.Levels[ClsRanger]) > 0 && rangers > 2 {
		msg = "too many rangers in party"
	}
	if msg != "" {
		e.narrate(msg)
	}
	return msg != ""
}
