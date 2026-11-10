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
| ROB's item chance | POR | EclRob read var 2 (the money percent) for each item's chance | var 3, as all 14 script ROBs give it (0 none, 255 all) | U1 | por-re 11eb406, por-go 03178bc |
| Movement on the sheet under Slow, Haste | POR | shown doubled under Slow, halved under Haste | the real movement; Pascal fix lands with viewchar (src/pure) | U5 | por-go 54c4549 (comment); Pascal pending |

## Port's own, not ruled: Felix to confirm

| What | Games | Original | New | Commit |
|---|---|---|---|---|
| Acid spit at a fallen target | POR | spits at a fallen target (nil fixed in the Pascal, por-re BUGS.md) | no spit at a fallen target. Verdict: design or oversight, unclear | port earlier; fixed: comment |
| Damage message length | POR | game cut to 40 characters. Presentation, not in src/pure; from the decomp, unconfirmed there | wrapped whole. Ruling asked. Verdict: real defect: cut text; presentation only | port earlier; fixed: comment |

## Unifications

| What | Games | Original | New | Ruling | Commit |
|---|---|---|---|---|---|
| Paladin saves | POR | own rows, L1 14,15,16,17,17 | COTAB's, L1 12,13,14,15,15 (L1-8) | 19 | port earlier; rules1e |
| Paladin, ranger attacks | POR | fighter only: 3 from L7 | paladin L7, ranger L8 also 3 (COTAB) | 19 | port earlier; rules1e |
| Races: half-elf ranger, cleric/ranger; human paladin, ranger | POR | not offered | offered; half-elf ranger takes the fighter's age | 19 | port earlier; rules1e |
| Ranger, cleric/ranger alignments | POR | any | good only (COTAB) | 19 | port earlier; rules1e |
| Paladin XP, L2-8 | POR | -1, no progression | COTAB's 2751 .. 175001 | 19 | port earlier; rules1e |
| Ranger XP, L2-7 | POR | -1, no progression | COTAB's 2251 .. 90001 | 19 | port earlier; rules1e |
| Money | all four | coins per character, with weight | one party purse, copper; no coin weight (decision page 13: a minor unification) | 11, page 13 | POR port earlier; others decided |
| Money: NPC share of the coins | POR | NpcTreasureShare cuts each of the five coin types and leaves each type's remainder | one copper pile is cut; equal whenever every pile divides (`TestNpcShareAgainstPascal`) | page 1: a minor unification | por-go hiredShare |
| Money: WHO skips the purse | POR | EclWho may pick by the character's coins | the party purse is no character's, so `purseWho` skips it | page 5 | por-go purseWho |
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
| Combat spell list | POR | every memorized spell, a noncombat one (whenCast 0: Friends, Read Magic, Find Traps, Knock, Strength, Cure Disease) cast nothing | only spells with whenCast <> 0 | U4 | por-go 54c4549 |
| Poison death | POR | "is killed", status Dead | the poison death message, status "Poisoned" for dead-with-poison | U6 | port earlier (affects1.go, extras.go) |
| Empty treasure screen | POR | shown with nothing on it | skipped; no rule effect | U7 | port earlier (aftercombat.go) |
| Bracers | POR | own slot; backstab needs no or leather armour | bracers in the armour slot (readySlot) and count as no armour for backstab | U8 | port earlier (combatturn2.go, charcalc.go) |
| Camp "Rest until fixed" | POR | none | taken from the later Gold Box games' Fix, always offered: clerics cast Cure Light Wounds on the most wounded, memorize it in every first-level slot and rest to learn it while needed, then relearn their planned spells; natural healing does the rest (`healParty`) | Felix, healParty ruling | port earlier (camp.go) |
| Diogenes' camp (area 19) | POR | on camp: `RestEvery = 0`, then Save 1 to 6B32 (ECL6 block 19, A592 `09 00 01 01 32 6B`): memorized slot 19 of whoever is selected | a semantic remap by the decompiler: the write is read as `RestChance = 1` (6DD3, a 1% chance the rest is interrupted), the camp setting beside RestEvery (patches/area19.patch) | Felix d4 | por-go 2b93868 |
| Off-hand weapon | POR | none | 4e rule set only; under 1e ApplyItems takes an off-hand weapon off | U9 | por-go 4b24740 |
| Heavy items when robbed | POR | RobItems cut the chance for items over 24 and 255 weight | no weight cut: items weigh nothing (ItemSave) | money/encumbrance unification | port earlier |

## Options (Felix's additions, not changes to the default)

Critical hits (ruling 4), party inventory (ruling 12), path search (page 3: Gameplay "Path search", default off, the computer walks by aiPathDir). Quick search (U2: Gameplay "Quick search", default off: a sweeping step or a Look takes 3 minutes instead of a turn; por-go 3be40ab). Day sky by clock (U3: Gameplay "Day sky by clock", default off: day 6:00-20:00 instead of the scripts' hour < 14; por-go 36339c3). Rest until healed (page 7, a convenience: camp rests until everyone is healed, `stopHealed`).
