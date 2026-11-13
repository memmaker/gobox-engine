package rules

import "strconv"

// Set is one of a Gold Box game's numbered data sets: most files come in
// eight numbered copies (ECL1..8, GEO, WALLDEF, 8X8D, ITEM, MONnCHA, HEAD,
// BODY, FINAL ...), one per part of the world, and the current area reads its
// set's (POR DS:5376, saved with the game, changed by the area scripts).
// Loaders name the set they read: the area's, or a fixed one.
type Set byte

// File is the name of family's File in set s ("ECL" -> "ECL3").
func (s Set) File(family string) string { return family + strconv.Itoa(int(s)) }
