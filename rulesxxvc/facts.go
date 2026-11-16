package rulesxxvc

import (
	"encoding/json"
	"errors"
	"fmt"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// ErrNoFact is a fact this ruleset's characters do not have.
var ErrNoFact = errors.New("rulesxxvc: no such fact")

// Fact is a script's read (eclvm.pas GetCharVar). KnownFact is BR 76+s, which
// reads 128 or 0: whether the character has the skill.
func (r *Rules) Fact(id rules.CharID, f rules.Fact) (rules.Value, error) {
	if v, ok, err := r.factSession(id, f); ok {
		return v, err
	}
	if v, ok, err := r.factFight(id, f); ok {
		return v, err
	}
	c := r.chars[id]
	if c == nil {
		if f == rules.FactActive {
			return rules.Yes(false), nil
		}
		return nil, fmt.Errorf("Fact %d: no character", id)
	}
	switch f := f.(type) {
	case rules.AbilityFact:
		if v, ok := c.Scores[f.Ability]; ok {
			return rules.Number(v), nil
		}
	case rules.SkillFact:
		if n, ok := skillNo(f.Skill); ok {
			return rules.Number(c.Skills.Level(n)), nil
		}
	case rules.KnownFact:
		if n, ok := skillNo(f.Skill); ok {
			return rules.Yes(c.Skills.Level(n) > 0), nil
		}
	case rules.SaveFact:
		if i := saveIndex(f.Save); i >= 0 {
			return rules.Number(c.Saves[i]), nil
		}
	case rules.CoinFact:
		if f.Coin == xxvc.Credits {
			return rules.Number(r.Purse(id)), nil
		}
	case rules.FlagFact:
		if f.Slot >= 0 && f.Slot < ScriptFlagSlots {
			return rules.Number(c.ScriptFlags[f.Slot]), nil
		}
	case rules.PlainFact:
		switch f {
		case rules.FactName:
			return rules.Text(c.Name), nil
		case rules.FactSex:
			return c.Sex, nil
		case rules.FactCompanion:
			return rules.Yes(c.Companion), nil
		case rules.FactRace:
			return c.Race, nil
		case rules.FactClasses:
			return rules.ClassSet{c.Career}, nil
		case rules.FactActive:
			return rules.Yes(c.Enabled), nil
		case rules.FactStatus:
			return coreStatus[c.Status], nil
		case rules.FactSide:
			return c.Side, nil
		case rules.FactHP:
			return rules.Number(c.HP), nil
		case rules.FactMaxHP:
			return rules.Number(c.MaxHP), nil
		}
	}
	return nil, fmt.Errorf("Fact %T %v: %w", f, f, ErrNoFact)
}

// SetFact stores; consequences are Actions (core.Fact).
func (r *Rules) SetFact(id rules.CharID, f rules.Fact, v rules.Value) error {
	if f, ok := f.(rules.ShipCellFact); ok {
		return r.setShipCell(f, v)
	}
	c := r.chars[id]
	if c == nil {
		return fmt.Errorf("SetFact %d: no character", id)
	}
	n, isNumber := v.(rules.Number)
	switch f := f.(type) {
	case rules.AbilityFact:
		if _, ok := c.Scores[f.Ability]; ok && isNumber {
			c.Scores[f.Ability] = int(n)
			return nil
		}
	case rules.SkillFact:
		if s, ok := skillNo(f.Skill); ok && isNumber {
			c.Skills.Set(s, int8(n))
			return nil
		}
	case rules.SaveFact:
		if i := saveIndex(f.Save); i >= 0 && isNumber {
			c.Saves[i] = int(n)
			return nil
		}
	case rules.FlagFact:
		if f.Slot >= 0 && f.Slot < ScriptFlagSlots && isNumber {
			c.ScriptFlags[f.Slot] = int(uint8(n)) // a byte in the record
			return nil
		}
	case rules.PlainFact:
		switch f {
		case rules.FactControl: // the npc byte: the morale in the low seven bits, NON_PLAYER_FLAG above (EclAddNpc)
			if isNumber {
				c.Morale, c.NonPlayer = int(n)&(nonPlayerFlag-1), int(n)&nonPlayerFlag != 0
				return nil
			}
		case rules.FactCompanion:
			if y, ok := v.(rules.Yes); ok {
				c.Companion = bool(y)
				return nil
			}
		case rules.FactName:
			if t, ok := v.(rules.Text); ok {
				c.Name = string(t)
				return nil
			}
		case rules.FactStatus:
			if st, ok := v.(rules.Status); ok {
				if b, ok := fromCore[st]; ok {
					c.Status = b
					c.Enabled = b == StOkay // isEnabled is kept beside status: 1 only for StOkay (types.pas; BR setChar r[213])
					return nil
				}
			}
		case rules.FactSide:
			if sd, ok := v.(rules.Side); ok {
				c.Side = sd
				return nil
			}
		case rules.FactHP:
			if isNumber {
				c.HP = int(n)
				return nil
			}
		case rules.FactMaxHP:
			if isNumber {
				c.MaxHP = int(n)
				return nil
			}
		}
	}
	return fmt.Errorf("SetFact %T %v = %v: %w", f, f, v, ErrNoFact)
}

func saveIndex(s rules.Save) int {
	for i, x := range xxvc.Saves {
		if x == s {
			return i
		}
	}
	return -1
}

// Import and Export carry one Character as JSON, the canonical format. The
// coins go with it; under the shared purse Add pools them.
func (r *Rules) Import(f rules.CharacterFile) (rules.CharID, error) {
	var c Character
	if err := json.Unmarshal(f, &c); err != nil {
		return rules.NoChar, fmt.Errorf("Import: %w", err)
	}
	return r.Add(c), nil
}

func (r *Rules) Export(id rules.CharID) (rules.CharacterFile, error) {
	c := r.chars[id]
	if c == nil {
		return nil, fmt.Errorf("Export %d: no character", id)
	}
	return json.Marshal(c)
}
