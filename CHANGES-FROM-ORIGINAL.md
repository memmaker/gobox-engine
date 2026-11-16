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
| POR script, journal references | POR | the text named an entry of the printed Adventurer's Journal ("UNDER ENTRY 48", "PROCLAMATIONS LXIV, ...") for the player to look up | the pager shows the entry (Journal), the reference left out of the text (47 sites; rules/scriptfixes.go) | por-go 38e0ef5 |
| POR script, quest handovers | POR | the original has no quest journal | each text that hands a quest over is followed by Quest(name), which notes it in the quest journal (30 sites; rules/scriptfixes.go); no printed text is matched | por-go (this) |
| COTAB/SotSB script, journal references | COTAB, SotSB | the text named an entry of the printed Adventurer's Journal ("RECORD IT IN JOURNAL ENTRY 31", the number often a PRINT of its own) | the import's journalFixes (rules/journalfixes.go: COTAB 55, SotSB 61 sites) follow each reference with Journal(entry), the text as the text window showed it; the port no longer matches printed text (journal.Ref, Printed gone) | cotab-go, sotsb-go (this) |
| POR script, area 0 InsideTown | POR | nothing at the site | fix: the arms shop's south door (wall 3, not 9) was never locked (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 0 Rolf's history of Phlan | POR | nothing at the site | (extra) Rolf tells the history of Phlan (2 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 0 shop and temple greetings | POR | goto SeekHealingAsk; goto SeekHealingAsk | the temple asks nothing first (8679e68): the jump to SeekHealingAsk and its 'DO YOU SEEK HEALING?' print go, the healing menu (TempleHealing) opens directly (6 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| Temple healing menu | all four | services for the selected member, then "Pay for cure?" | services for the whole party (green: helps anyone, grey: no one); hovering marks the cards it would help; picked, "<Service> for whom?" takes members one after the other (click, F1-F8, arrows+Enter), each paid and cured at once, no pay question; Esc/right click/Exit end it, as does no one left | Felix 2026-11-16 | this commit |
| POR script, area 0 TavernTale | POR | Text("YOU OVERHEAR TAVERN TALE "); Scratch6 += 1; Print(Scratch6); WaitReturnPressed() | (extra) the tale itself in the pager; the original printed "YOU OVERHEAR TAVERN TALE n" for the player to look up (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 0 tavern wager in gold | POR | Print(" PP."); HOW MANY PLATINUM PIECES WILL YOU WAGER? (HOUSE LIMIT IS 1000); &Wager > 1000 | gold: " GP.", GOLD PIECES, house limit 5000 (1000 platinum in gold), one `TavernGame()` command over the party purse in 64-bit gold (importer/por_scriptfixes.go fix 58); DEFAULT (Felix may override) (1 site; importpor/scriptfixes.go) | por-go Df |
| POR script, area 1 InsideSlavePen | POR | nothing at the site | fixed: the boy was offered again after he was bought, rescued or lost (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 2 on camp | POR | nothing at the site | fix: no encounter is left (RandomEncounter exits), the rest was still broken off (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 2 fled encounters | POR | if CombatEnd > PartyDied { goto EjectToStreet } | running away goes to RandomFled (ChasedOut's fix) (6 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 2 Restal's seal once | POR | nothing at the site | fix: once, not every visit (2 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 5 ThroneRoom | POR | if Scratch10 >= 85 { goto ArrogantKnave } | fixed: was >= 85, so a weak abusive party got the confession and a strong one the fight (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 7 VoteNext | POR | TowerEndFlags = GenheerisFlags \| DragonDone | fixed: the original stored GenheerisFlags\|DragonDone in TowerEndFlags (4A72, never read), so the dragon fight and the ending could be replayed (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 9 coins typed | POR | TakeCoins:; Scratch10 = 195; Scratch10 = 193; Scratch10 = 191; Scratch10 = 189; Scratch10 = 187 | typed: the coins are SelCoins[Scratch10] (platinum 8 .. copper 0); was SelName[187 + Scratch10] with its high byte in V_6B01, Scratch7 (3 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 9 AlarmActive | POR | AlarmEndHour = 1 + AuxCounter | fixed: was 1 + AuxCounter (a loop temp); the castle areas use AlarmHour (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 9 GateTowerSquare | POR | if Scratch9 = GateFlags & WestGarrisonSlain; Scratch9 == 0 { Scratch2 \|= 2 }; if GateFlags lacks EastGarrisonSlain { Scratch2 \|= 1 }; goto TowersCharge; ; EastTowerCheck: | fixed: the two sides tested the wrong tower as the near one (cell 11 is below the east tower, TowerExitCellTable), so a slain garrison charged again; cell 10, below the west tower (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 9 FleeToArea6 | POR | NewECL(ValjevoCastle6) | fix: the original left SymbolSet at 2, loading area 6 from the wrong data set (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 9 bugbear bluff | POR | BugbearWarnings = Suspicious | fixed: BugbearWarnings = Suspicious here made every later visit an impostor accusation; it belongs to ImpostorsAccused (2 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 14 WeaponPileAmbush | POR | nothing at the site | fix: 13-15 killed made 5 - 6..8 thieves, a byte gone below 0 (253-255) (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 14 AmbushManyMasters | POR | nothing at the site | fix: the original left the thief count from earlier (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 16 FindPoolCellLoop | POR | if GiantLizardsKilled < 10 { Scratch8 = 10 - PoolTreasureTaken } | fixed: the original subtracted PoolTreasureTaken (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 16 YesNoMenu | POR | Scratch9 = Menu("PRESS <RETURN> OR BUTTON TO CONTINUE") | fixed: was Scratch9, which holds the building's fight (BuildingEncounter) across a wrong friend-word (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 18 on camp | POR | if PatrolsBeaten18 >= 64 { RestSafe() } | fix: the original ran on and set RestEvery 2, RestChance 16 over these (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 18 PatrolFight | POR | CountPatrolBeaten() | fix: running away counted as beating the patrol (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 18 FreeForAllMember | POR | nothing at the site | fix: with a one-member party 8+Random(44) can pass the last of the 50 patrons and leave the member selected (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 20 fled fights | POR | nothing at the site | fix: running away counted as beating it (12 sites; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 24 on move | POR | Scratch1 = Table99A8[Scratch10] | fix: no area that way (the south openings): the party bumps instead of loading map 255 (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 25 KeepRiders | POR | goto [FightKeepRiders, ExitEvent][Scratch1] | fix: LEAVE exited with the castle picture still up (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 25 PassageOutside | POR | Scratch1 = Menu(exit="LEAVE CAVE", "STAY IN") | the menu's Esc row (exit=, 5c9b793) was the decompiler's guess LEAVE CAVE; Esc stays in, as the cave's other menus (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 28 CellDispatch | POR | nothing at the site | fixed: the open west side of the guard post (4,0) skipped the commandant's attack at its south door (4,1) (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 28 DwarfAndCommandant | POR | LoadMonster(AlHyamDazid, count=1, icon=AlHyamDazid) | fixed: was AlHyamDazid, met alive later (area5); block 82 (a mage) is not in MON6CHA and aborted the game; the disk has no other mage (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |
| POR script, area 28 RelaxOrWalk | POR | nothing at the site | fixed: no dinner invitation after fighting the guards (1 site; rules/scriptfixes.go) | por-go aece709 or earlier |

## Unifications

| What | Games | Original | New | Ruling | Commit |
|---|---|---|---|---|---|
| Script DAMAGE | POR, SotSB, POD | POR: no 0x10 bit (flags&0x1F the save bonus only); SotSB: damage re-rolled per random member; POD: 0x10 = a made save halves, bonus mask 15 | COTAB's (O03C:2958) for all: flags 0x80 with 0x10 = a made save still hurts, bonus mask 31, one roll for the flagged forms; rules1e `scriptDamage` (core.ScriptDamage). POR: no constant DAMAGE operand in ECL*.DAX sets 0x80 and 0x10 together, so its scripts play as before (variable operands V_6E7D unverified); TestAgainstPascal `damage` documents it | 3 (S-diffs) | rules1e, por-go (this) |
| Paladin saves | POR | own rows, L1 14,15,16,17,17 | COTAB's, L1 12,13,14,15,15 (L1-8) | 19 | port earlier; rules1e |
| Paladin, ranger attacks | POR | fighter only: 3 from L7 | paladin L7, ranger L8 also 3 (COTAB) | 19 | port earlier; rules1e |
| Races: half-elf ranger, cleric/ranger; human paladin, ranger | POR | not offered | offered; half-elf ranger takes the fighter's age | 19 | port earlier; rules1e |
| Ranger, cleric/ranger alignments | POR | any | good only (COTAB) | 19 | port earlier; rules1e |
| Paladin XP, L2-8 | POR | -1, no progression | COTAB's 2751 .. 175001 | 19 | port earlier; rules1e |
| Ranger XP, L2-7 | POR | -1, no progression | COTAB's 2251 .. 90001 | 19 | port earlier; rules1e |
| Monk saves | POR | the class rows | also the old-level row (COTAB oldClassSaveData, O0FE:07FB): a monk with no old level takes row 0, 11,10,10,14,11; no POR career is a monk | Felix 18:10 (saves) | rules1e |
| Paladin spells, L9+ | POR | none | COTAB's cleric slots (paladinSlotData, O0FE:000F); POR's paladins stop at 8 | Felix 18:10 (slots) | rules1e |
| Ranger spells, L8+ | POR | none | COTAB's druid and magic-user slots (rangerSlotData); POR's rangers stop at 7 | Felix 18:10 (slots) | rules1e |
| Thief skills: items of thievery, free action | POR | none (POR's items have no powers) | COTAB's (O0FE:0AEA): levels lifted to 5/7 or +5 kept across skills; free action lifts to 4 or +10 | Felix 18:10 (thief) | rules1e |
| Magic-user slots: item of wizardry | POR | none (no item powers) | COTAB's: doubled | Felix 18:10 (slots) | rules1e |
| Constitution save bonus | POR | dwarf/halfling affect 0x5A adds 1-5 (Con 4-18) to the paralysis roll | COTAB's (O0FE:07FB): subtracted from the stored paralysis save (the bonus helps; COTAB's original adds it to the target, making the save harder) for a dwarf, halfling or save-bonus item, plus 1-4 for everyone at Con 19-25, in all games; affect 0x5A is no longer given and is dropped from loaded records (no double count) | Felix (FOLD-PLAN T; subtract ruling) | rules1e |
| XP past a class's table | all four | POR reads other data at L10-13 (thief L10: 0, no level), 0 past 13; COTAB reads on into the next class's row; SotSB 0 (no level) | POD's table rows first (Tables.XPPast, levels to 11), then POD's step per level (cleric 225000, druid 750000, fighter 250000, paladin 350000, ranger 325000, magic-user 375000, thief 220000) up to level 40; a table ending at -1 stays -1. POR's thief L10 = 160001, L11 220001, L12 440001 (TestAgainstPascal) | Felix 20:05 (T4), pod-rows-first ruling (T5) | rules1e |
| Hit dice on level-up | COTAB, SotSB, POD | first-level dice at L1 only, then 1 die; fixed hit points from classMaxHitDice | POR's: the first-level dice count every level, no cap (Tables.MaxHitDice deleted) | Felix 20:05 (T4) | rules1e (the games after E2) |
| Race ability limits | SotSB, POD | by the character's sex | the men's limits for everyone (POR, COTAB) | Felix 20:05 (T4) | rules1e (the games after E2) |
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
| Cursed bow in the computer's turn | POR | a cursed bow stays in the weapon hand: next to a foe the holder neither shoots nor strikes | the bow is readied in the Launcher slot and takes no hand (365b9eb4), so the holder strikes bare-handed as with a plain bow; the Pascal's cursed `aibow` vectors are checked against the plain ones (engine/decide_pascal_test.go). | Felix 2026-11-15 | 365b9eb4 |
| Camp "Rest until fixed" | POR | none | taken from the later Gold Box games' Fix, always offered: clerics cast Cure Light Wounds on the most wounded, memorize it in every first-level slot and rest to learn it while needed, then relearn their planned spells; natural healing does the rest (`healParty`) | Felix, healParty ruling | port earlier (camp.go) |
| Cabinets' CAST KNOCK (area 14) | POR | the Spell command looked at the first member only and never at slot 0, kept the answer addresses in bytes so the script never saw it, and the script cleared the slot after the Knock | one script command placed at import (`Scratch6 = CastKnock()`, fix 30): the first okay member with Knock memorized casts it the door's way (`spendKnock`), the slot spent; the Spell command is gone; the cabinets stay script events | Felix d3 | por-go 32639f9 |
| Diogenes' camp (area 19) | POR | on camp: `RestEvery = 0`, then Save 1 to 6B32 (ECL6 block 19, A592 `09 00 01 01 32 6B`): memorized slot 19 of whoever is selected | a semantic remap by the decompiler: the write is read as `RestChance = 1` (6DD3, a 1% chance the rest is interrupted), the camp setting beside RestEvery (patches/area19.patch) | Felix d4 | por-go 2b93868 |
| River in area 26 | POR | the river script kept its drawn/opened flag in 4A9A, the game word CouncilMeeting (area 8), so the river clobbered it | its own area field `RiverState` (layout area26.txt, `replaces`); patches/area26.patch gone; not saved until saves by field (1b-f) | 1b-e | por-go 1c766dc |
| Off-hand weapon | POR | none | 4e rule set only; under 1e ApplyItems takes an off-hand weapon off | U9 | por-go 4b24740 |
| Heavy items when robbed | POR | RobItems cut the chance for items over 24 and 255 weight | no weight cut: items weigh nothing (ItemSave) | money/encumbrance unification | port earlier |

## Options (Felix's additions, not changes to the default)

Critical hits (ruling 4), party inventory (ruling 12), path search (page 3: Gameplay "Path search", default off, the computer walks by aiPathDir). Quick search (U2: Gameplay "Quick search", default off: a sweeping step or a Look takes 3 minutes instead of a turn; por-go 3be40ab). Day sky by clock (U3: Gameplay "Day sky by clock", default off: day 6:00-20:00 instead of the scripts' hour < 14; por-go 36339c3). Rest until healed (page 7, a convenience: camp rests until everyone is healed, `stopHealed`).
