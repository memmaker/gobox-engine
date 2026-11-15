package rules1e

// Convenience features that are not in the DOS game. The port is a
// modernized build: they are always on (there is no mode with the
// original's exact behaviour). Most are reached through keys the original
// command bar ignores:
//
//	Tab  minimap of the whole area over the roster panel, where the game's
//	     own Area map was allowed; it replaces Area, which is not offered
//	M    the full-screen map, also a click on the minimap: labels (fullmap.go)
//	     and Walk, which takes the party to a square (walk.go)
//	Down a step backwards (the original turns round)
//	P    party overview: hit points, armour class and status of everyone
//	?      keyboard help, at any prompt
//	F1-F8  select party member 1-8 (Left/Right the previous/next on the
//	       menu screens); P switches between items and the View page
//	Ctrl-F next text font (Ctrl-Shift-F the one before; gobox, its font pool)
//	Q    quickstart on the empty party menu: a ready-made party of six
//	F10  named save slots: save and load at any time (slots.go)
//
// One set of navigation keys for every menu, the mouse and the hint line
// below the picture are in nav.go.
//
// Menus also get a selection indicator: the highlighted item sits on a
// coloured background (the original only recolours its text, which is the
// colour the hotkey letters already have).
