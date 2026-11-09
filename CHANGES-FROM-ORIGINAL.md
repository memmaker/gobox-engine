# Changes from the originals

Felix: "Just keep a documentation on what has changed where." One row per bug fix or unification: what changed, in which games, original against new, ruling, commit. Rulings are numbered as in ARCHITECTURE.md section 9. This is the list of allowed differences from the reference, the pure Pascal (`por-re` `bugfixed-rules`, which itself carries the 18 fixes of `BUGS.md` plus the reference fixes listed there): any other difference is a bug. "Decided" has no commit yet; it lands after that game's port. Sources: `~/Projects/{por,cotab,sotsb,pod}-re/src`, `buckroger-re`, `buckrogers2-re`.

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
| NPC share of the coins | POR | NpcTreasureShare cuts each of the five coin types and leaves each type's remainder | money is copper (Felix's ruling): one copper pile is cut; equal whenever every pile divides (`TestNpcShareAgainstPascal`) | por-go hiredShare |
|---|---|---|---|---|
| Cleric and mage L1 slots | POR | row skipped below L2: record keeps its slots | 1,0,0 plus Wisdom bonus, so a level drain leaves no stale slots | port earlier; rules1e |
| Cleric, mage slots past L6 | POR | row (zeros) overwrites | record keeps its own (Genheeris) | port earlier; rules1e |
| Cleric and its multi-class alignments | POR | any | no true neutral; origin unchecked | port earlier; rules1e |
| Turn undead at cleric level 14+ | POR | last column | same; the port panicked, reading past its table | rules1e |
| THAC0 base at level -1 (monster records) and fighter 11+ | POR | the neighbouring row (F12 base 40; F-1 base 46) | same; the port panicked. Original bug kept, ruling asked | rules1e |
| Mirror Image | POR | only a spell being cast takes an image; blows never do | blows and spells take one; origin: the port's own fix, ruling asked | port earlier; `TestEffectsAgainstPascal` |
| Magic resistance against items | POR | GetCasterLevel gives 6 for an item spell, so the chance falls by 25 points | item counts as level 11, no cut. Ruling asked. Verdict: design choice in the original: item level 6 is deliberate, the effect on resistance is its consequence | port earlier; fixed: comment |
| Acid spit at a fallen target | POR | spits at a fallen target (nil fixed in the Pascal, por-re BUGS.md) | no spit at a fallen target. Verdict: design or oversight, unclear | port earlier; fixed: comment |
| Backstab multiple | POR | thief level div 4 + 2: levels 4 and 8 get the next multiple early | (level - 1) div 4 + 2. Ruling asked. Verdict: real defect against the AD&D table, or design choice if the game meant it; the code alone can't tell. Compared now: differs only at levels 4 and 8 (`TestDecideAgainstPascal`) | port earlier; fixed: comment |
| Damage message length | POR | game cut to 40 characters. Presentation, not in src/pure; from the decomp, unconfirmed there | wrapped whole. Ruling asked. Verdict: real defect: cut text; presentation only | port earlier; fixed: comment |
| Hold Person save | POR | table saveType: vs spell | vs paralysis. Ruling asked. Verdict: design choice in the original: vs spell is AD&D's rule. The port's change is the odd one out; original not a defect | port earlier; fixed: comment |
| Dispel Magic targets | POR | the game's binary read the first target every pass; src/pure already has spellTargets[i] | spellTargets[i] | no difference from the reference; real defect in the binary, fixed by BUGS.md's pure. No ruling needed |
| Computer's approach step | POR | direction toward the target, offset by the AI stance | `aiPathDir` path search, stance reset to straight on (`extra:` comment). Ruling asked. Verdict: design choice in the original; the port's addition, not a defect fix. Moves the walker only in the compared turns | port earlier; `TestDecideAgainstPascal` |
| Surrendering creature's rolls | POR | src/pure CombatAiControl: the removed 'flees in panic' left its `if` on `if done then Exit`, so a surrendered creature goes on through the turn and spends rolls | the turn ends at the surrender, as in por-re src/combatai.pas | defect of the reference, not of the game; asked |
| Team count after turning undead | POR | src/pure CombatTurnUndead: the removed 'Nothing Happens...' left CountTeams under `if not turnedAny`, so a destroyed undead stays counted | counted always, as in por-re src/cbtacts.pas | defect of the reference; asked |
| Aligned weapon outside combat | POR | src/pure EffAlignedWeapon: the removed redraw left the damage under `if IN_COMBAT` | damage always, as in por-re src/effhndlr.pas | defect of the reference; asked |
| Combat aim of a spell | POR | src/pure CastSpell has only NoncombatCastTargetMenu: caster, or the whole list | `chooseSpellTargets` in real combat; the vectors run with it off | gap of the reference: combat aiming is not compared; asked |
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
