# Straight combat map and secret doors (2026-11-15)

Goal: the Deskew dungeon combat board is the original's (ovr11:0810 map,
ovr11:19F5 placement) with only the skew taken out, never a generator or a
placement of its own; the automap shows a secret door only once found.

How: one mapping, `rules1e.StraightX` (original square -> board column),
used by the ground's `put`, the rubble and `placeAt`. Invariant test:
`engine.TestStraightIsOriginalUnskewed`.

## Checklist
- [x] Root cause of foes behind the Barracks' secret door: Deskew made distance 0 a cell ahead, through any side (38106227 stopped that; superseded below)
- [x] Straight board = original map unskewed (dungeonMapFlat, PlaceStart*Flat, the floor formation, the distance bump: deleted)
- [x] Placement = the original's algorithm and squares through `StraightX`; a 6-wide block made 5 is not one-to-one, so on a wall line open to the west two squares share one and the second combatant takes the next free slot
- [x] Walls: a cell is a square 5x5 block (the original's 4x4 floor, walls one square thick, doors 2 squares between posts, corners at wall ends: `squareWalls`, laid before placement so no one stands on a post); Felix: a map cell is a square
- [x] Tests: TestStraightIsOriginalUnskewed, TestStraightWalls; TestSidesApart, TestDeskewPlacement and the watch fight's "sides apart" check removed (BUG-CLASSES 29 is superseded: the original mixes the front rows at distance 0)
- [ ] Clutter: its dice fall by board square, so the straight board's clutter differs from the skewed one's (same odds)
- [x] Secret doors on the automap only once found: `WallTexture.Solid` (importer/walls/solid.txt, marked by Felix per texture), a passable side behind a solid-looking wall is secret until gone through or searched beside (Search, or a step while sweeping: `searchSecrets`; `foundSecret`, SECRETS.GOB); TestSecretDoorOnMap
- [ ] Full suite + a real fight looked at in the game
