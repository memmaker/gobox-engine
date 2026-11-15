package engine

import "gobox/rules"

// The place, rest and hall words are the engine's (e.world); a ruleset is handed the values a
// question needs, with the question.

// noMagic: the area forbids spells and items' magic.
func (e *Engine) noMagic() bool { return e.world.NoMagic != 0 }

// in3D: the area is a 3D one, not the overland map (not viewMode 1, the 3D view up: Board.Dungeon).
func (e *Engine) in3D() bool { return e.world.In3D != 0 }

// hallTerms are what the training hall offers: the classes its word names, and the debug switch.
func (e *Engine) hallTerms() rules.HallTerms {
	return rules.HallTerms{Teaches: rules.ClassBits(e.world.PartyExp), Free: e.FreeTraining}
}

// areaMarks are the words that decide which landmarks the area map shows.
func (e *Engine) areaMarks() rules.AreaMarks {
	w := &e.world
	return rules.AreaMarks{Flags4: w.MarkFlags4, Mark3: w.Mark3, Mark4Off: w.Mark4Off, Flags57: w.MarkFlags57, Mark4: w.Mark4, Boss: w.Boss}
}
