package rulesxxvc

import (
	"encoding/json"
	"errors"
	"fmt"

	"gobox/rules"
)

// XXVc's character file is one JSON of Character (Import, Export): the engine
// keeps it as the roster's .cha record, and the other files of rules.CharacterFiles
// stay empty (the items, effects, portrait and icon are in the JSON, or not held).

var _ rules.FileFormat = (*Rules)(nil)

var errExtraFiles = errors.New("rulesxxvc: a character is one JSON file, no .itm, .spc, .mem, .dual, .pow or picture files")

// Pack is the canonical character of its files: the record, checked as a Character.
func (r *Rules) Pack(f rules.CharacterFiles) (rules.CharacterFile, error) {
	if len(f.Items)+len(f.Affects)+len(f.Plan)+len(f.Dual)+len(f.Powers)+len(f.Portrait)+len(f.Icon) > 0 {
		return nil, errExtraFiles
	}
	if _, err := decodeCharacter(rules.CharacterFile(f.Record)); err != nil {
		return nil, err
	}
	return rules.CharacterFile(f.Record), nil
}

// Unpack is the canonical character as its files: the one record, a game save's too.
func (r *Rules) Unpack(c rules.CharacterFile, _ bool) (rules.CharacterFiles, error) {
	if _, err := decodeCharacter(c); err != nil {
		return rules.CharacterFiles{}, err
	}
	return rules.CharacterFiles{Record: c}, nil
}

// IsRecord: the bytes decode as a Character.
func (r *Rules) IsRecord(record rules.SavedRecord) bool {
	_, err := decodeCharacter(rules.CharacterFile(record))
	return err == nil
}

// RosterOf is the roster number in the record (NoRoster: none, or not a record).
func (r *Rules) RosterOf(record rules.SavedRecord) rules.RosterID {
	c, err := decodeCharacter(rules.CharacterFile(record))
	if err != nil {
		return rules.NoRoster
	}
	return c.Roster
}

// WithRoster is the record with its roster number replaced (an unreadable record stays).
func (r *Rules) WithRoster(record rules.SavedRecord, id rules.RosterID) rules.SavedRecord {
	c, err := decodeCharacter(rules.CharacterFile(record))
	if err != nil {
		return record
	}
	c.Roster = id
	b, err := json.Marshal(c)
	if err != nil {
		return record
	}
	return b
}

// SavedLevels is the level of the character's one career (none: not a record).
func (r *Rules) SavedLevels(record rules.SavedRecord) []int {
	c, err := decodeCharacter(rules.CharacterFile(record))
	if err != nil {
		return nil
	}
	return []int{c.Level}
}

func decodeCharacter(b rules.CharacterFile) (Character, error) {
	var c Character
	if err := json.Unmarshal(b, &c); err != nil {
		return Character{}, fmt.Errorf("character file: %w", err)
	}
	return c, nil
}
