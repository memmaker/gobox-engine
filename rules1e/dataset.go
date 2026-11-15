package rules1e

import (
	"gobox/rules"
)

// checkMonsters answers the Checks of the monster archive.
func (r *Rules) checkMonsters(k rules.Check) (rules.Answer, bool) {
	switch k := k.(type) {
	case rules.MonsterKnown:
		if r.Monsters == nil {
			return rules.Answer{}, true
		}
		_, ok := r.Monsters(k.ID)
		return rules.Answer{Yes: ok}, true
	case rules.FirstMonster:
		if r.FirstMonster == nil {
			return rules.Answer{}, true
		}
		id, ok := r.FirstMonster()
		return rules.Answer{Yes: ok, Value: int32(id)}, true
	}
	return rules.Answer{}, false
}

// addTreasureRecords is AddTreasureRecords: the records, whole ones only, as a block.
func (r *Rules) addTreasureRecords(a rules.AddTreasureRecords) error {
	items := ReadItems(a.Records[:len(a.Records)/ItemSize*ItemSize])
	if len(items) == 0 {
		return nil
	}
	b, err := r.S.gameItems(items) // the records are the game's
	if err != nil {
		return err
	}
	r.S.AddTreasureBlock(b)
	return nil
}
