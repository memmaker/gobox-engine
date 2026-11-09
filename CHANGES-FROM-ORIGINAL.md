# Changes from the originals

Felix: "Just keep a documentation on what has changed where." One row per bug fix or unification: what changed, in which games, original against new, ruling, commit. Rulings are numbered as in ARCHITECTURE.md section 9. This is the list of allowed differences from the reference, the pure Pascal (`por-re` `rules-only`, which itself carries the 17 fixes of `BUGS.md`): any other difference is a bug. "Decided" has no commit yet; it lands after that game's port. Sources: `~/Projects/{por,cotab,sotsb,pod}-re/src`, `buckroger-re`, `buckrogers2-re`.

## Bug fixes

| What | Games | Original | New | Ruling | Commit |
|---|---|---|---|---|---|
| Thief Dex 10, pick pockets | all four | -19 | -10 | table base | POR port earlier (por-go tables.go:931); others decided |
| Thief Dex 16, open locks | all four | -5 | +5 | table base | POR port f314dce; others decided |
| Paladin XP, L11 | POD | 1050000 | 1050001 | table base | decided |
| Paladin XP, L15 | SotSB | 2460001 | 2450001, **Felix to confirm** | table base | decided |
| Ranger XP, L11 | POD | 650000 | 650001 | table base | decided |
| Druid XP past L11 | POD | -1 + (L-11)*750000 | no progression | table base | decided |
| First hit dice, cleric | SotSB | 2 | 1, **Felix to confirm** | table base | decided |
| Party strength sum | POR, COTAB | Byte, wraps | no wrap, **Felix to confirm** | source check | decided |
| Saves | all four | byte arithmetic: a total below 0 wraps and saves | signed, clamped at the right stage | 1 | POR port f314dce or earlier, now rules1e `SaveMade`; others decided |
| Cleric XP bonus threshold | POD, SotSB | off by one | Wis >= 16, `xp*110/100` | 8 | decided |
| Charisma lookup | all four | per game | `max(3, min(cha, maxCharisma))` | 14 | decided |

## Port's own, not ruled: Felix to confirm

| What | Games | Original | New | Commit |
|---|---|---|---|---|
| Cleric and mage L1 slots | POR | row skipped below L2: record keeps its slots | 1,0,0 plus Wisdom bonus, so a level drain leaves no stale slots | port earlier; rules1e |
| Cleric, mage slots past L6 | POR | row (zeros) overwrites | record keeps its own (Genheeris) | port earlier; rules1e |
| Cleric and its multi-class alignments | POR | any | no true neutral; origin unchecked | port earlier; rules1e |
| Turn undead at cleric level 14+ | POR | last column | same; the port panicked, reading past its table | rules1e |
| THAC0 base at level -1 (monster records) and fighter 11+ | POR | the neighbouring row (F12 base 40; F-1 base 46) | same; the port panicked. Original bug kept, ruling asked | rules1e |
| Dwarf's -4 from giants | POR | only if the attacker has race 1, size 2 (so only ogres); name read from the defender's target | any attacker named on the list; origin: the port's own fix, ruling asked | port earlier; `TestEffectsAgainstPascal` |
| Paladin L9 saves | POR | table | table, as the Pascal; the port panicked | rules1e f7797ad |

## Unifications

| What | Games | Original | New | Ruling | Commit |
|---|---|---|---|---|---|
| Paladin saves | POR | own rows, L1 14,15,16,17,17 | COTAB's, L1 12,13,14,15,15 (L1-8) | 19 | port earlier; rules1e |
| Paladin, ranger attacks | POR | fighter only: 3 from L7 | paladin L7, ranger L8 also 3 (COTAB) | 19 | port earlier; rules1e |
| Races: half-elf ranger, cleric/ranger; human paladin, ranger | POR | not offered | offered; half-elf ranger takes the fighter's age | 19 | port earlier; rules1e |
| Ranger, cleric/ranger alignments | POR | any | good only (COTAB) | 19 | port earlier; rules1e |
| Paladin XP, L2-8 | POR | -1, no progression | COTAB's 2751 .. 175001 | 19 | port earlier; rules1e |
| Ranger XP, L2-7 | POR | -1, no progression | COTAB's 2251 .. 90001 | 19 | port earlier; rules1e |
| Money | all four | coins per character, with weight | one party purse, copper; coin weight **Felix to confirm** | 11 | POR port earlier; others decided |
| Initiative | all four | POR d100 per pick; POD coin toss on ties | each round: roll, modifiers, deterministic tie-break, sort | 2 | decided |
| Difficulty | SotSB, POD | setting scales monsters or to-hit | removed; normal values | 3, 5, 6 | decided |
| XP, campaign and race+class caps | all four | mixed | separate; cap stops training, keeps XP | 9 | decided |
| Traps | all four | per game; stoned skip | data; no universal stoned skip | 10 | decided |
| Shop prices | all four | factor word and per-game code | adventure answers per shop; copper | 11 | decided |
| Encumbrance | all four | per game | one, default rules | 12 | decided |
| Armour move | all four | identical | one implementation, same values | 13 | decided |
| Spells | all four | per-game tables | one registry, stable ids; rows per ARCHITECTURE.md | 15, 19 | decided |
| Spell slots | all four | per game | class and level tables plus cleric Wisdom | 16 | decided |
| Dual class | all four | per game | everywhere; adventure restricts where it starts | 17 | decided |
| Doors | all four | Knock special | adventure defines; Knock an ordinary spell | 18 | decided |
| Rule tables | all four | per game | base COTAB/SotSB, POD named exceptions | 19 | decided |
| Fighter-line and mage THAC0 | POR, POD | see ARCHITECTURE.md table comparison | COTAB/SotSB | 19 | decided |
| Attack hook | all four | per game | common interface | 20 | decided |
| Dancing | POD only | POD spell | shared condition | 21 | decided |

## Options (Felix's additions, not changes to the default)

Critical hits (ruling 4), party inventory (ruling 12).
