package rulesxxvc

import (
	"encoding/json"
	"fmt"

	"gobox/rules"
	"gobox/rules/xxvc"
)

// OriginalClock is the game clock an original saved game holds (types.pas zone0 clock), one field a unit, each
// clockUnitScale of the one before.
type OriginalClock struct {
	Segment, Minute, Turn, Hour, Day, Month, Year int
}

// OriginalGame is what an original saved game (SAVGAMx.DAT) holds of this ruleset's state, typed: the importer fills it from
// the file's bytes and the ruleset never sees them. The members come as the engine's own character files, the script cells and
// the world words as the engine's, so this is only the ruleset's own: the clock, the party's ship, the last enemy ship, the
// difficulty and the wall sets the view loads back (saveload.pas LoadGame).
type OriginalGame struct {
	Clock      OriginalClock
	Ship       xxvc.ShipRecord
	Foe        SavedFoe
	Difficulty Difficulty
	Walls      rules.RecordView
}

// units are the clock's units in clockSegment.. order.
func (c OriginalClock) units() [clockUnits]int {
	return [clockUnits]int{c.Segment, c.Minute, c.Turn, c.Hour, c.Day, c.Month, c.Year}
}

// Setting is the level as the player's option.
func (d Difficulty) Setting() rules.Difficulty {
	switch d {
	case Novice:
		return rules.DifficultyEasiest
	case Cadet:
		return rules.DifficultyEasier
	case Veteran:
		return rules.DifficultyHarder
	case Ace:
		return rules.DifficultyHardest
	}
	return rules.DifficultyStandard
}

// Save is the ruleset's blob of a game loaded from the original: no characters (the engine loads the party's files after it),
// the purse empty (the members carry their credits).
func (g OriginalGame) Save() ([]byte, error) {
	if g.Difficulty < Novice || g.Difficulty > Ace {
		return nil, fmt.Errorf("rulesxxvc: difficulty %d is no level", g.Difficulty)
	}
	for i, n := range g.Clock.units() {
		if n < 0 || n >= clockUnitScale[i] {
			return nil, fmt.Errorf("rulesxxvc: clock unit %d is %d, past its scale %d", i, n, clockUnitScale[i])
		}
	}
	ship := g.Ship
	return json.Marshal(saved{Chars: map[rules.CharID]*Character{}, Next: 1, Ship: &ship, Foe: g.Foe, Clock: g.Clock.units(),
		Walls: g.Walls, Difficulty: g.Difficulty.Setting()})
}
