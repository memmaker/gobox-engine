package rules1e

import "gobox/rules"

// Files is one character's files (rules.CharacterFiles).
type Files = rules.CharacterFiles

var _ rules.FileFormat = (*Rules)(nil)

// Pack is Pack285.
func (*Rules) Pack(f Files) (rules.CharacterFile, error) { return Pack285(f) }

// Unpack is Unpack285.
func (*Rules) Unpack(c rules.CharacterFile, game bool) (Files, error) { return Unpack285(c, game) }

// IsRecord: a record is its 0x11D bytes, then the trailer.
func (*Rules) IsRecord(record rules.SavedRecord) bool { return len(record) >= int(Layout285) }

// RosterOf is the roster number in a saved record's trailer.
func (*Rules) RosterOf(record rules.SavedRecord) rules.RosterID { return ReadPoolIDs(record).Roster }

// WithRoster is record's 0x11D bytes and its trailer with the roster number id.
func (*Rules) WithRoster(record rules.SavedRecord, id rules.RosterID) rules.SavedRecord {
	ids := ReadPoolIDs(record)
	ids.Roster = id
	return append(record[:Layout285:Layout285], PoolTrailer(ids)...)
}

// SavedLevels are the levels of a saved record, each a signed byte.
func (*Rules) SavedLevels(record rules.SavedRecord) []int {
	var lv []int
	for _, l := range SheetOf(record).Levels {
		lv = append(lv, int(int8(l)))
	}
	return lv
}

// MakeStrong rewrites a .SAV character record so the character always hits
// and takes long to die (cmd/fixture -strong: a scripted fight is won).
func MakeStrong(rec rules.SavedRecord) rules.SavedRecord {
	c := Decode[Record285](rec)
	c.Thac0Base = 40 // THAC0 is recomputed from it
	c.MaxHP, c.HP = 250, 250
	c.BaseAC, c.AC[0], c.AC[1] = 80, 80, 80 // AC -20 (60 - AC)
	c.BaseDice[0], c.BaseSides[0] = 6, 12   // unarmed damage 6d12
	return rules.SavedRecord(Encode(c))
}
