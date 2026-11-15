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
- [x] Placement = the original's, square for square
- [x] Walls: placed on the original's squares, then squared up (`squareWalls`): one square thick both ways, doors 2 squares between posts, a corner at every wall end (Felix: the original's 2-wide west walls felt odd), floors 5x4
- [x] Tests: TestStraightIsOriginalUnskewed, TestStraightWalls; TestSidesApart, TestDeskewPlacement and the watch fight's "sides apart" check removed (BUG-CLASSES 29 is superseded: the original mixes the front rows at distance 0)
- [ ] Clutter: its dice fall by board square, so the straight board's clutter differs from the skewed one's (same odds)
- [x] Secret doors on the automap only once found: `WallTexture.Solid` (importer/walls/solid.txt, marked by Felix per texture), a passable side behind a solid-looking wall is secret until gone through or searched beside (Search, or a step while sweeping: `searchSecrets`; `foundSecret`, SECRETS.GOB); TestSecretDoorOnMap
- [ ] Full suite + a real fight looked at in the game
