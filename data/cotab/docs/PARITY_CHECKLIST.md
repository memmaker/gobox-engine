# Function parity checklist

Routine-by-routine check that each Go port behaves like the original, not
just that it exists. Tick a box only after comparing the Go code with the
listing (`python3 decomp/tools/view.py ovrNN LO HI`), using the Ghidra
decompile as a guide to its structure (see `.claude/skills/ghidra/SKILL.md`). Check:
branches and conditions (signed vs unsigned), loop bounds, every far call
(target and argument order), DS globals read and written, return value, and
quirks kept on purpose.

Rule routines of cotab's engine (generated from the `// name = SEG:OFF` doc comments;
the engine keeps no drawing, so these are the rules).

## O03C
- [x] O03C:0050 `cmdExit` (eclcmds.go)
- [x] O03C:02E9 `cmdLoadCharacter` (eclcmds.go)
- [x] O03C:03C9 `cmdSetupMonster` (eclcmds3.go)
- [x] O03C:0465 `cmdLoadMonster` (eclcmds2.go)
- [x] O03C:0835 `cmdApproach` (eclcmds3.go)
- [x] O03C:0873 `cmdPicture` (eclcmds3.go)
- [x] O03C:095E `cmdInputNumber` (eclcmds.go)
- [x] O03C:09A4 `cmdInputText` (eclcmds.go)
- [x] O03C:0A1C `cmdPrint` (eclcmds.go)
- [x] O03C:0BE7 `cmdNewECL` (eclcmds.go)
- [x] O03C:0C41 `cmdLoadFiles` (eclcmds3.go)
- [x] O03C:0EE9 `cmdVerticalMenu` (eclcmds2.go)
- [x] O03C:10B4 `cmdHorizontalMenu` (eclcmds.go)
- [x] O03C:1240 `clearTreasure` (adventure.go)
- [x] O03C:1240 `cmdClearMonsters` (eclcmds.go)
- [x] O03C:12A9 `cmdPartyStrength` (eclcmds.go)
- [x] O03C:1454 `cmdCheckParty` (eclcmds2.go) — fixed: 4th var is the affect-found flag (census leaves it 0, not the count); affect check leaves lowest 0
- [x] O03C:167E `cmdPartySurprise` (eclcmds2.go)
- [x] O03C:171E `cmdSurprise` (eclcmds2.go) — fixed: operands in the frame's order (party arg4+2-arg1, monsters arg2+2-arg3)
- [x] O03C:17E4 `cmdCombat` (eclcmds2.go)
- [x] O03C:1B9D `cmdTreasure` (treasure.go)
- [x] O03C:1F76 `cmdRob` (eclcmds4.go)
- [x] O03C:20BD `cmdEncounter` (eclcmds2.go) — a menu digit past the table is ignored (asm reads past v6a)
- [x] O03C:27B7 `cmdParlay` (eclcmds2.go) — prompt is " " (asm passes a char local as a string, junk length)
- [x] O03C:2856 `cmdFindItem` (eclcmds2.go)
- [x] O03C:290B `cmdDelay` (eclcmds2.go)
- [x] O03C:2958 `cmdDamage` (damage.go)
- [x] O03C:2CB6 `cmdSpriteOff` (eclcmds2.go)
- [x] O03C:2CDA `cmdClock` (eclcmds4.go)
- [x] O03C:2D0F `cmdPrintReturn` (eclcmds2.go)
- [x] O03C:2D38 `cmdClearBox` (eclcmds4.go)
- [x] O03C:2D7F `cmdWho` (eclcmds4.go)
- [x] O03C:2DCA `cmdAddNPC` (eclcmds4.go)
- [x] O03C:2E33 `cmdSpell` (eclcmds2.go) — fixed: slot n is record byte 30+n (was Memorized[n-1]; scan runs past Memorized to 131)
- [x] O03C:2F23 `cmdCall` (eclcmds3.go) — animation frame counter DS:7247 of call 26627 not kept (drawing)
- [x] O03C:3094 `campEvents` (adventure.go)
- [x] O03C:30FE `cmdProgram` (eclcmds2.go)
- [x] O03C:323F `cmdProtection` (eclcmds4.go) — code wheel question dropped
- [x] O03C:3271 `cmdDump` (eclcmds2.go)
- [x] O03C:32A5 `cmdFindSpecial` (eclcmds2.go)
- [x] O03C:32F9 `cmdDestroyItems` (eclcmds2.go)
- [x] O03C:3607 `runEvent` (eclrun.go)
- [x] O03C:3677 `scriptRestart` (eclrun.go)
- [x] O03C:3758 `Adventure` (eclrun.go) — fixed: b4FDB cleared after the redraw test (it gated it, not keepScript); keepScript (b47FB) no longer cleared

## O042
- [x] O042:0045 `castCureAnyway` (temple.go)
- [x] O042:00F1 `payForCure` (temple.go)
- [x] O042:027E `templeCureBlindness` (temple.go)
- [x] O042:0322 `templeCureDisease` (temple.go)
- [x] O042:043B `templeCureWounds` (temple.go)
- [x] O042:0636 `templeRaiseDead` (temple.go)
- [x] O042:088B `templeNeutralizePoison` (temple.go)
- [x] O042:0974 `templeRemoveCurse` (temple.go)
- [x] O042:0A62 `templeStoneToFlesh` (temple.go)
- [x] O042:0B11 `templeHeal` (temple.go)
- [x] O042:0DD9 `templeMenu` (temple.go)

## O045
- [x] O045:0037 `computeExp` (aftercombat.go) — Go skips the divide when no one earns (the original divides by zero)
- [x] O045:036A `giveExp` (aftercombat.go)
- [x] O045:0556 `combatResult` (aftercombat.go)
- [x] O045:0ABC `showResult` (aftercombat.go)
- [x] O045:0D2B `lootList` (aftercombat.go)
- [x] O045:0DFC `lootTake` (aftercombat.go)
- [x] O045:0F2E `takeTreasure` (aftercombat.go)
- [x] O045:10C3 `lootMenu` (aftercombat.go)
- [x] O045:13C7 `cleanupCombatants` (aftercombat.go)
- [x] O045:150E `hiredShare` (aftercombat.go)
- [x] O045:17A2 `AfterCombat` (aftercombat.go)

## O048
- [x] O048:004E `shopList` (shop.go) — price factor 16 (unset in original buy code) uses the plain price
- [x] O048:034C `takeItemCopy` (shop.go)
- [x] O048:0474 `shopBuy` (shop.go)
- [x] O048:06E7 `shopMenu` (shop.go)

## O04C
- [x] O04C:0168 `arg` (eclvm.go)
- [x] O04C:01E8 `initScript` (eclrun.go)
- [x] O04C:036C `loadScript` (eclrun.go)
- [x] O04C:04B4 `freeSteps` (eclcmds3.go)
- [x] O04C:0543 `showPortrait` (eclcmds3.go) — the O110 portrait drawing is left to the front-end
- [x] O04C:0580 `showEncounter` (eclcmds3.go) — Show3DSprite and the pO110_000A draw are left to the front-end
- [x] O04C:07E4 `charValue` (eclvars.go) — for value 269 and for charisma outside 3..25 the original returns whatever was in the result; Go returns 0
- [x] O04C:0B4F `setCharValue` (eclvars.go)
- [x] O04C:0D52 `setVar` (eclvars.go)
- [x] O04C:0F16 `getVar` (eclvars.go)
- [x] O04C:105D `setTextVar` (eclvm.go)
- [x] O04C:1421 `readTextVar` (eclvm.go)
- [x] O04C:1693 `menuHotkeys` (eclcmds.go)
- [x] O04C:17AA `scriptMenu` (eclcmds.go)
- [x] O04C:18AE `ListMenu` (listmenu.go) — the ECL wrapper; its "~" strip goes to a dead local
- [x] O04C:1B01 `stepForward` (eclvm.go)
- [x] O04C:1B9A `scriptDuel` (eclcmds2.go)
- [x] O04C:1DEF `robMoney` (eclcmds4.go)
- [x] O04C:1F1C `robItems` (eclcmds4.go)
- [x] O04C:1FD2 `skipCommand` (eclrun.go)
- [x] O04C:20E5 `partySpeeds` (eclcmds2.go) — haste doubles, slow halves (the original has them swapped)
- [x] O04C:2200 `damageWithMessage` (damage.go)

## O057
- [x] O057:004B `combatEnd` (combat.go)
- [x] O057:0100 `Combat` (combat.go)
- [x] O057:01BC `nextActor` (combat.go)
- [x] O057:0281 `takeTurn` (combat.go)
- [x] O057:03EC `playerTurn` (combat.go)
- [x] O057:0788 `combatMenu` (combat.go)
- [x] O057:09B3 `endRound` (combat.go)
- [x] O057:0B26 `moveAttack` (combat.go)
- [x] O057:0B26 `stepCost` (combat.go) — O08D:0748 multiplies without the 255 case; moveAttack's check is moveCost
- [x] O057:0F03 `attackTarget` (combat.go)
- [x] O057:1003 `otherMenu` (combat.go)
- [x] O057:11FA `speedMenu` (combat.go)
- [x] O057:132F `quickOne` (combat.go)

## O05A
- [x] O05A:004B `o10_000F` (combatai.go)
- [x] O05A:024B `aiTryCast` (combatai.go)
- [x] O05A:03B1 `canCast` (combatai.go)
- [x] O05A:04AA `aiCastFromItem` (combatai.go)
- [x] O05A:060B `aiCastKnown` (combatai.go)
- [x] O05A:073B `aiCanStep` (combatai.go)
- [x] O05A:09D1 `aiMoveStep` (combatai.go)
- [x] O05A:0DB1 `aiAttackLoop` (combatai.go)
- [x] O05A:11F7 `aiGiveUp` (combatai.go)
- [x] O05A:1269 `aiKeys` (combatai.go)
- [x] O05A:137F `aiFleeCheck` (combatai.go)
- [x] O05A:1535 `weaponScore` (combatai.go)
- [x] O05A:1673 `aiPickWeapon` (combatai.go)

## O05D
- [x] O05D:00D3 `dungeonRubble` (combatsetup.go)
- [x] O05D:0306 `cellSide` (combatsetup.go)
- [x] O05D:0388 `cellEdge` (combatsetup.go)
- [x] O05D:08CD `dungeonMap` (combatsetup.go)
- [x] O05D:0991 `wildTerrain` (combatsetup.go) — past the table the original reads the following DS bytes; Go gives 0
- [x] O05D:0A00 `wildCliff` (combatsetup.go)
- [x] O05D:0B0B `wildTrees` (combatsetup.go)
- [x] O05D:0E4A `wildScatter` (combatsetup.go)
- [x] O05D:0FC8 `wildMap` (combatsetup.go)
- [x] O05D:1030 `combatMapSetup` (combatsetup.go)
- [x] O05D:10E0 `combatBlocks` (combatsetup.go)
- [x] O05D:1202 `formationEdge` (combatsetup.go)
- [x] O05D:1233 `placeAt` (combatsetup.go)
- [x] O05D:1380 `placeOne` (combatsetup.go)
- [x] O05D:17FE `placeCombatants` (combatsetup.go)
- [x] O05D:1C3B `combatSetup` (combatsetup.go) — the screen, icon and text-delay resets (b4FE2, b7289/728E/72A6, b72AA/7288/728D) are drawing only

## O060
- [x] O060:0054 `Init` (init.go)
- [x] O060:06E5 `newGameState` (init.go)

## O064
- [x] O064:0019 `blockAffect` (affects.go)
- [x] O064:0038 `addAffectNow` (affects.go)
- [x] O064:0071 `loseTurnHandler` (affects.go)
- [x] O064:0087 `saveAndHitUp` (affects.go)
- [x] O064:0096 `blessHandler` (affects.go)
- [x] O064:00A6 `curseHandler` (affects.go)
- [x] O064:00DC `snakesHandler` (affects.go)
- [x] O064:015F `wardEnchantedHandler` (affects.go)
- [x] O064:017A `kindBaneHandler` (affects.go)
- [x] O064:01DF `armourBonusHandler` (affects.go)
- [x] O064:0224 `protFromEvilHandler` (affects.go)
- [x] O064:0259 `protFromGoodHandler` (affects.go)
- [x] O064:028E `resistColdHandler` (affects.go)
- [x] O064:02AD `charmHandler` (affects.go)
- [x] O064:0379 `suffocateHandler` (affects.go)
- [x] O064:03BC `poisonDamageHandler` (affects.go)
- [x] O064:041F `shieldHandler` (affects.go)
- [x] O064:044A `vsGoblinoidHandler` (affects.go)
- [x] O064:0480 `resistFireHandler` (affects.go)
- [x] O064:04B1 `silencedHandler` (affects.go)
- [x] O064:0517 `slowPoisonHandler` (affects.go)
- [x] O064:0583 `spiritualHammerHandler` (affects.go)
- [x] O064:06C6 `invisibleHandler` (affects.go)
- [x] O064:06FB `vsOrcHandler` (affects.go)
- [x] O064:073F `mirrorImageHandler` (affects.go)
- [x] O064:07C0 `quarterLessHandler` (affects.go)
- [x] O064:07E8 `coughingHandler` (affects.go)
- [x] O064:089E `collapseHandler` (affects.go)
- [x] O064:0951 `dazzledHandler` (affects.go)
- [x] O064:0974 `weakAndWastingHandler` (affects.go)
- [x] O064:09D9 `confusedHandler` (affects.go)
- [x] O064:0B6F `hamperedHandler` (affects.go)
- [x] O064:0B80 `stumblingHandler` (affects.go)
- [x] O064:0BA5 `hasteHandler` (affects.go)
- [x] O064:0C1D `clearCloud` (clouds.go)
- [x] O064:0C1D `stinkingCloudEndHandler` (clouds.go)
- [x] O064:0F06 `avoidDistantAttack` (affects.go)
- [x] O064:0F77 `readiedWeapon` (affects.go)
- [x] O064:0FE0 `avoidPlainWeaponsHandler` (affects.go)
- [x] O064:101B `slowHandler` (affects.go)
- [x] O064:1038 `weakenHandler` (affects.go)
- [x] O064:10C2 `wastingHandler` (affects.go)
- [x] O064:1153 `vsGiantsHandler` (affects.go)
- [x] O064:11A2 `gnomeExtraHandler` (affects.go)
- [x] O064:11C9 `prayerHandler` (affects.go)
- [x] O064:1212 `fireBodyHandler` (affects.go)
- [x] O064:1243 `iceBodyHandler` (affects.go) — deviation: tests the saving throw, the original DS:5D0B
- [x] O064:127B `flickerHandler` (affects.go)
- [x] O064:129A `engulfedHandler` (affects.go)
- [x] O064:12BA `regeneratingHandler` (affects.go)
- [x] O064:12D8 `magicWeaponsOnlyHandler` (affects.go)
- [x] O064:132B `fireDamageReducedHandler` (affects.go)
- [x] O064:1386 `healingOverTimeHandler` (affects.go)
- [x] O064:13CA `minorGlobeHandler` (affects.go)
- [x] O064:1407 `poisonAttack` (affects.go)
- [x] O064:14AE `paralyseTarget` (affects.go)
- [x] O064:1520 `poisonStingHandler` (affects.go)
- [x] O064:1534 `poisonSting4Handler` (affects.go)
- [x] O064:1548 `poisonSting2Handler` (affects.go)
- [x] O064:155C `paralysingTouchHandler` (affects.go)
- [x] O064:1595 `feeblemindHandler` (affects.go)
- [x] O064:1636 `hiddenFromKind19Handler` (affects.go)
- [x] O064:1671 `poisonStingMinus2Handler` (affects.go) — the asm passes byte 254, i.e. -2
- [x] O064:1685 `blurHandler` (affects.go)
- [x] O064:1696 `flicker95Handler` (affects.go)
- [x] O064:16D2 `unaffectedByType32Handler` (affects.go)
- [x] O064:171A `vsKind3Handler` (affects.go)
- [x] O064:1772 `vsKind8Handler` (affects.go)
- [x] O064:17B3 `berserkHandler` (affects.go)
- [x] O064:18D9 `recoveringHandler` (affects.go)
- [x] O064:1919 `burningTouchHandler` (affects.go)
- [x] O064:194C `acidTouchHandler` (affects.go)
- [x] O064:197F `halfDamageHandler` (affects.go)
- [x] O064:1990 `saveVsElementsHandler` (affects.go)
- [x] O064:19E1 `electricityHealsHandler` (affects.go)
- [x] O064:1A14 `arrowsScratchHandler` (affects.go)
- [x] O064:1A55 `firstStrikeHandler` (affects.go)
- [x] O064:1AB9 `cloudkillEndHandler` (clouds.go)
- [x] O064:1D98 `fireHalvedHandler` (affects.go)
- [x] O064:1DB2 `rangedHalvedHandler` (affects.go)
- [x] O064:1E06 `fallsDeadHandler` (affects.go)
- [x] O064:1E42 `conSaveHandler` (affects.go) — deviation: below constitution 4 adds 0, the original an unset local
- [x] O064:1EB8 `regrowthHandler` (affects.go)
- [x] O064:1EE8 `backsUpHandler` (affects.go)
- [x] O064:1F91 `regrowsHandler` (affects.go)
- [x] O064:1FCC `regeneratesHandler` (affects.go)
- [x] O064:201E `regrowingHandler` (affects.go)
- [x] O064:205D `burnsHandler` (affects.go)
- [x] O064:20DA `avoidRanged60Handler` (affects.go)
- [x] O064:20EE `resistRoll` (affects.go)
- [x] O064:214F `magicResist50Handler` (affects.go)
- [x] O064:215D `magicResist15Handler` (affects.go)
- [x] O064:216B `elfResistHandler` (affects.go)
- [x] O064:218F `noCharmHoldHandler` (affects.go)
- [x] O064:21A4 `noParalysisHandler` (affects.go)
- [x] O064:21B2 `coldImmuneHandler` (affects.go)
- [x] O064:21C9 `noPoisonParalysisHandler` (affects.go)
- [x] O064:21EA `fireImmuneHandler` (affects.go)
- [x] O064:2201 `fireReducedHandler` (affects.go)
- [x] O064:2246 `electricityHalvedHandler` (affects.go)
- [x] O064:2260 `edgedHalvedHandler` (affects.go)
- [x] O064:22BF `magicWeaponsHalvedHandler` (affects.go)
- [x] O064:22F9 `slingStonesHandler` (affects.go)
- [x] O064:233C `coldHalvedHandler` (affects.go)
- [x] O064:2356 `needsMagicWeaponHandler` (affects.go)
- [x] O064:23A2 `avoidsBowsHandler` (affects.go)
- [x] O064:23F6 `spitsAcidHandler` (affects.go)
- [x] O064:2523 `paralysingGazeHandler` (affects.go)
- [x] O064:259D `fireTouchHandler` (affects.go)
- [x] O064:25D0 `halfElfResistHandler` (affects.go)
- [x] O064:25F4 `noCharmHoldSleepHandler` (affects.go)
- [x] O064:2623 `magicImmuneHandler` (affects.go)
- [x] O064:2643 `slainByArrowHandler` (affects.go)
- [x] O064:26D3 `itemHandler` (affects.go)
- [x] O064:2750 `noMindAffectsHandler` (affects.go)
- [x] O064:277C `rangerVsGiantsHandler` (affects.go)
- [x] O064:27B5 `electricityImmuneHandler` (affects.go)
- [x] O064:27CC `rootedHandler` (affects.go)
- [x] O064:27E0 `berserkFoeHandler` (affects.go) — deviation: the nearest-target search runs in combat only (the original also outside, on a stale map)
- [x] O064:28D0 `alwaysInvisibleHandler` (affects.go)
- [x] O064:28EF `paladinCuresHandler` (affects.go)
- [x] O064:2932 `runningAwayHandler` (affects.go)
- [x] O064:2975 `zapsHandler` (affects.go)
- [x] O064:2A19 `dispelEvilHandler` (affects.go)
- [x] O064:2AD8 `setupAffectHandlers` (affects.go)

## O08D
- [x] O08D:0000 `turnSetup` (combatturn1.go)
- [x] O08D:0124 `turnMoveRate` (combatturn1.go)
- [x] O08D:0192 `rollDamage` (combatturn1.go)
- [x] O08D:0318 `attackText` (combatturn1.go)
- [x] O08D:065D `guardsStrike` (combatturn1.go)
- [x] O08D:0748 `stepSquare` (combatturn1.go)
- [x] O08D:0954 `leaveAttacks` (combatturn1.go)
- [x] O08D:0D1B `fleeBattle` (combatturn1.go)
- [x] O08D:0DD4 `setSpeed` (combatturn1.go)
- [x] O08D:0F0D `halfUp` (combatturn1.go)
- [x] O08D:0F3D `sweepAttack` (combatturn1.go)
- [x] O08D:1143 `mayAct` (combatturn1.go)
- [x] O08D:1228 `turnUndead` (combatturn1.go) — undead types 11-12 or a current cleric level 0 index outside the table (Go panics, the original reads neighbouring DS bytes)
- [x] O08D:1433 `weakestFoe` (combatturn1.go)
- [x] O08D:14EB `attackSlots` (combatturn1.go)
- [x] O08D:194D `countAttack` (combatturn1.go)
- [x] O08D:19DB `fullAttack` (combatturn1.go) — the original points the selected character (DS:6520) at any attacker; Go only switches partySel for party members
- [x] O08D:1CED `rangeBand` (combatturn1.go)
- [x] O08D:1DFE `healTarget` (combatturn2.go)
- [x] O08D:201C `chooseSpellTarget` (combatturn2.go)
- [x] O08D:2231 `chooseSpellTargets` (combatturn2.go)
- [x] O08D:276D `spellMenuTargets` (combatturn2.go)
- [x] O08D:28D7 `canBackstab` (combatturn2.go)
- [x] O08D:29BC `directionTo` (combatturn2.go)
- [x] O08D:2BF1 `attackAnimation` (combatturn2.go)
- [x] O08D:2E00 `monsterHPPercent` (combatturn2.go)
- [x] O08D:2E8F `maxPartyMove` (combatturn2.go)
- [x] O08D:2F1F `mayAttack` (combatturn2.go)
- [x] O08D:31D8 `aimConfirm` (combatturn2.go)
- [x] O08D:3395 `aimManual` (combatturn2.go) (deviation: the cursor walk is one square question to the UI; the checks on the square are the original's)
- [x] O08D:3B25 `aimAt` (combatturn2.go)
- [x] O08D:3E44 `pickAttackTarget` (combatturn2.go)
- [x] O08D:4032 `engulfsHandler` (affects.go)
- [x] O08D:42F2 `raysHandler` (affects.go)
- [x] O08D:45C6 `engulfingHandler` (affects.go)
- [x] O08D:46FC `huggingHandler` (affects.go)
- [x] O08D:480C `hugsHandler` (affects.go)
- [x] O08D:48D9 `godIntervene` (combatturn2.go)

## O099
- [x] O099:0148 `openDoor` (adventure.go)
- [x] O099:0275 `partyHasClassLevel` (adventure.go)
- [x] O099:02F3 `bashDoor` (adventure.go)
- [x] O099:05B6 `pickLock` (adventure.go)
- [x] O099:0683 `memorizedSlot` (adventure.go)
- [x] O099:06C6 `memorizedBy` (adventure.go)
- [x] O099:0720 `castKnock` (adventure.go)
- [x] O099:0765 `stepAhead` (adventure.go)
- [x] O099:0813 `walkForward` (adventure.go)
- [x] O099:08DF `adventureInput` (adventure.go) — fixed: View opens the sheet; Look takes ten minutes
- [x] O099:0B80 `lockedDoor` (adventure.go) — fixed: a stuck step count (>=255) resets the step word, not the encounter distance

## O09D
- [x] O09D:0032 `studyMinutes` (camp.go)
- [x] O09D:018B `forgetPending` (camp.go)
- [x] O09D:01CE `stopScribing` (camp.go)
- [x] O09D:0246 `stopStudy` (camp.go)
- [x] O09D:028E `slotsLeft` (camp.go)
- [x] O09D:03A0 `canStudy` (camp.go)
- [x] O09D:04D5 `castMenu` (camp.go)
- [x] O09D:05D4 `memorizeInfo` (camp.go)
- [x] O09D:07C2 `restMenu` (camp.go)
- [x] O09D:086F `sortSpells` (camp.go)
- [x] O09D:0955 `memorizeMenu` (camp.go)
- [x] O09D:0BAB `scribeMenu` (camp.go)
- [x] O09D:103A `displayEffects` (camp.go)
- [x] O09D:14D5 `magicMenu` (camp.go)
- [x] O09D:158D `movePrev` (camp.go)
- [x] O09D:16E5 `moveNext` (camp.go)
- [x] O09D:1864 `orderMenu` (camp.go)
- [x] O09D:19C9 `dropMember` (camp.go)
- [x] O09D:1B7B `campSpeedMenu` (camp.go)
- [x] O09D:1D38 `alterMenu` (camp.go) — Icon (O0A7:3EC6 icon builder) not ported, only the picture reload
- [x] O09D:1F22 `castCures` (camp.go)
- [x] O09D:1FDD `rollCures` (camp.go)
- [x] O09D:208F `woundsTotal` (camp.go)
- [x] O09D:20ED `fixRestTime` (camp.go) — fixed: no extra 27/34/78 minutes; the shortening uses the original uninitialised counter (starts at 14 on its stack), study minutes carry over past members out of action
- [x] O09D:2280 `applyHealing` (camp.go)
- [x] O09D:2338 `fixMenu` (camp.go)
- [x] O09D:242B `campMenu` (camp.go)

## O0A1
- [x] O0A1:001E `cString` (charimport.go) — unverified: the original sizes the second Pool scan by DS:8223, Go by poolRecSize
- [x] O0A1:0465 `rosterFiles` (charimport.go)
- [x] O0A1:0655 `saveDirReady` (charfiles.go)
- [x] O0A1:0C90 `removeCharacterFiles` (saveload.go)
- [x] O0A1:0DFC `SaveCharacter` (charfiles.go)
- [x] O0A1:13AE `nameInFiles` (charimport.go) — deviation: male strength limits for both sexes, exceptional strength clamps str00 instead of raising strength (deliberate)
- [x] O0A1:14B3 `loadPoolRecord` (charimport.go)
- [x] O0A1:1F35 `hillsfarMerge` (charimport.go)
- [x] O0A1:22A0 `importCharacter` (charimport.go)
- [x] O0A1:22A0 `loadSavedCharacter` (charadd.go)
- [x] O0A1:329C `loadMonsterRecord` (eclcmds2.go)
- [x] O0A1:357D `addNPC` (eclcmds4.go)
- [x] O0A1:360A `AddToParty` (charadd.go)
- [x] O0A1:37B5 `LoadGame` (saveload.go)
- [x] O0A1:3F5E `SaveGame` (saveload.go) — the disk-free check (savgam size + party memory vs DiskFree) and the "Put save disk in" retry are left out; a write error goes straight to "Unexpected error during save"

## O0A7
- [x] O0A7:003C `freeCharacter` (create.go)
- [x] O0A7:01CC `PartyMenu` (party.go) — fixed: "Quit anyway? N" saves and stays in the menu; added the J free-training cheat (b75A1, debug word); Go redraws after every select key, the original only when Change Class eligibility flips
- [x] O0A7:0752 `CreateCharacterMenu` (create.go) — kept quirks: racial roll adjustments test the race-list loop counter (always 7), so never apply; hit-die roll wasted; mage/thief starts at 8333 exp; deviation: female strength limits not used (8569d05)
- [x] O0A7:24EE `conBonusForClass` (train.go)
- [x] O0A7:25DA `DropCharacter` (create.go) — fixed: roster files deleted (O0A1:0C90 remove_player_file) before free_players
- [x] O0A7:25DA `DropMember` (party.go) — fixed: roster files deleted (O0A1:0C90 remove_player_file) before free_players
- [x] O0A7:26F2 `modifyDraw` (train.go) — drawing only
- [x] O0A7:2830 `ModifyCharacterMenu` (train.go) — name editing via Decision.Text instead of the key-by-key editor; deviation: strength limits not split by sex (8569d05)
- [x] O0A7:3648 `AddCharacter` (charadd.go)
- [x] O0A7:3A25 `freePlayers` (party.go) — the icon free is drawing, left to the front end
- [x] O0A7:3A25 `freeSelected` (adventure.go)
- [x] O0A7:3B7C `iconPane` (train.go) — drawing only
- [x] O0A7:45C1 `conHitPointBonusOf` (train.go)
- [x] O0A7:46BA `minHitPoints` (train.go)
- [x] O0A7:4793 `maxHitPoints` (train.go)
- [x] O0A7:49E0 `rollHitPoints2` (train.go)
- [x] O0A7:4D07 `TrainCharacter` (train.go)

## O0AC
- [x] O0AC:0B79 `playPicture` (ending.go) — drawing only, a no-op here
- [x] O0AC:111C `endGameText` (ending.go) — the bonds fade, Knights portrait and fireworks (O0AC:09F4) are drawing and not played

## O0AF
- [x] O0AF:0B70 `viewPlayer` (viewplayer.go) — Go uses e.selected() where the original reads DS:4B0C to pick which of Items/Spells/Trade/Drop appear
- [x] O0AF:0EC1 `canLoseItem` (view.go)
- [x] O0AF:1583 `ItemsMenu` (view.go)
- [x] O0AF:1B04 `itemReadyEffect` (view.go) — fixed: power 8 only for scores 0..5, as the original's case
- [x] O0AF:1ED6 `readyItem` (view.go) — "already using" for a taken ring 2 names Slots[9] (the original reads through it)
- [x] O0AF:211C `tradeItem` (view.go) — fixed: the default target is p4B10 (viewStart), shared with tradeCoins; load_pic left to the front-end
- [x] O0AF:21D5 `halveItem` (view.go)
- [x] O0AF:2285 `joinItems` (view.go) — Spells compared with themselves in the original, kept (only Spells[0] < 2 is tested)
- [x] O0AF:2478 `useItem` (view.go)
- [x] O0AF:27C2 `sellItem` (view.go) — stack price is (count*half mod 65536)/20 as the asm; the arrows/quarrels branch can never be taken; the original's garbage high byte of willOverload's result is not reproduced
- [x] O0AF:2A29 `identifyItem` (view.go)
- [x] O0AF:324B `overloadedBy` (shop.go)
- [x] O0AF:32CF `cycleSelected` (eclcmds.go) — partyKeypad (party.go, unused duplicate) removed
- [x] O0AF:347C `spellMenu2` (spellmenu.go) — SpellList (view.go) merged into it; the title is drawing only
- [x] O0AF:35F0 `canLayOnHands` (viewplayer.go)
- [x] O0AF:3655 `canCureDisease` (viewplayer.go)
- [x] O0AF:36D7 `layOnHands` (viewplayer.go)
- [x] O0AF:37EC `paladinCure` (viewplayer.go)

## O0B6
- [x] O0B6:001E `tickEffects` (game.go)
- [x] O0B6:0317 `carryClock` (camp.go)
- [x] O0B6:0317 `normalizeClock` (game.go)
- [x] O0B6:03C8 `normalizeRest` (camp.go)
- [x] O0B6:03FA `AdvanceClock` (game.go)
- [x] O0B6:049F `subtractRest` (camp.go)
- [x] O0B6:058A `twoDigits` (camp.go)
- [x] O0B6:0615 `showRest` (camp.go)
- [x] O0B6:0751 `restTimeMenu` (camp.go)
- [x] O0B6:08A1 `healTick` (camp.go)
- [x] O0B6:0960 `memorizeOne` (camp.go)
- [x] O0B6:0A22 `scribeOne` (camp.go) — fixed: goes on past the scribed spell to the next one's level
- [x] O0B6:0B4D `studyTick` (camp.go)
- [x] O0B6:0C03 `hourTick` (camp.go)
- [x] O0B6:0CEF `restFor` (camp.go) — fixed: marks every member's effects to tick (DS:471E); no key-press interrupt

## O0BA
- [x] O0BA:0F7B `treasureLeft` (coins.go)
- [x] O0BA:0FCF `itemBonus` (treasure.go)
- [x] O0BA:1007 `createItem` (treasure.go)
- [x] O0BA:1928 `appraise` (coins.go) — keeping a gem with no items adds it (the original loses it)

## O0C2
- [x] O0C2:001E `canUseSpell` (spellmenu.go)
- [x] O0C2:0211 `spellMenu` (spellmenu.go)
- [x] O0C2:03ED `addSpellLine` (spellmenu.go)
- [x] O0C2:05B9 `addSpellSorted` (spellmenu.go)
- [x] O0C2:0912 `addScrollSpells` (spellmenu.go)
- [x] O0C2:09F4 `addAllScrolls` (spellmenu.go)
- [x] O0C2:0A74 `buildSpellList` (spellmenu.go)
- [x] O0C2:0DE5 `spellRange` (spells1.go)
- [x] O0C2:0E92 `spellAmount` (spells1.go)
- [x] O0C2:0F7F `applySpell` (spells1.go) — dmg now truncated to a byte like the asm parameter
- [x] O0C2:114D `spellTargetSetup` (spells1.go)
- [x] O0C2:12E1 `CastSpell` (spells1.go) — now sets DS:7600 (showMoves) before turning the caster; "can't be cast here" text not drawn
- [x] O0C2:16B7 `clampToMap` (spells2.go)
- [x] O0C2:17CF `coneTargets` (spells2.go) — fixed: the side-line direction is taken only from squares with a combatant; with none the asm reads an uninitialised byte, Go uses 0
- [x] O0C2:1B24 `holdTargets` (spells2.go)
- [x] O0C2:1CA0 `applySpellToSide` (spells1.go)
- [x] O0C2:1D60 `fxBless` (spells2.go)
- [x] O0C2:1D91 `fxBane` (spells2.go)
- [x] O0C2:1DBC `fxCureLight` (spells2.go)
- [x] O0C2:1DF8 `fxCauseLight` (spells2.go)
- [x] O0C2:1E3A `fxAffected` (spells2.go)
- [x] O0C2:1E74 `fxProtected` (spells2.go)
- [x] O0C2:1EB3 `fxColdResistant` (spells2.go)
- [x] O0C2:1EE1 `fxBurningHands` (spells2.go)
- [x] O0C2:1F2E `fxCharmPerson` (spells2.go)
- [x] O0C2:2005 `spellStronger` (spells2.go)
- [x] O0C2:213A `spellReduce` (spells2.go)
- [x] O0C2:21DA `spellFriends` (spells2.go)
- [x] O0C2:2221 `fxMagicMissile` (spells2.go)
- [x] O0C2:2284 `fxShielded` (spells2.go)
- [x] O0C2:22B2 `fxShockingGrasp` (spells2.go)
- [x] O0C2:2306 `fxSleep` (spells2.go)
- [x] O0C2:2444 `fxHold` (spells2.go)
- [x] O0C2:24A7 `fxFireResistant` (spells2.go)
- [x] O0C2:24E0 `fxSilenced` (spells2.go)
- [x] O0C2:2519 `fxSlowPoison` (spells2.go)
- [x] O0C2:25CD `fxCharmSnakes` (spells2.go)
- [x] O0C2:2681 `fxSpiritualHammer` (spells2.go)
- [x] O0C2:26D4 `fxInvisibility` (spells2.go)
- [x] O0C2:270D `fxKnock` (spells2.go)
- [x] O0C2:2748 `fxMirrorImage` (spells2.go)
- [x] O0C2:27A6 `fxRayOfEnfeeblement` (spells2.go)
- [x] O0C2:27EB `createCloud` (clouds.go) — stinking cloud: older ground from other stinking clouds (last match wins) or from under a body; one breath per combatant
- [x] O0C2:2C5B `spellStrength` (spells2.go)
- [x] O0C2:2E19 `fxAnimateDead` (spells2.go)
- [x] O0C2:2FC4 `fxCureBlindness` (spells2.go)
- [x] O0C2:300A `fxBlindness` (spells2.go)
- [x] O0C2:3037 `cureDiseaseAffects` (spells2.go)
- [x] O0C2:30DC `fxCureDisease` (spells2.go)
- [x] O0C2:30F9 `fxDisease` (spells2.go)
- [x] O0C2:3126 `dispelsCloud` (spells2.go)
- [x] O0C2:31E4 `fxDispelMagic` (spells2.go) — global loop var DS:7761 not kept
- [x] O0C2:359A `fxPrayer` (spells2.go)
- [x] O0C2:3605 `spellRemoveCurse` (spells1.go) — fixed: magic property test was > 128, asm is > 127; callAffect gets nil + callItem where the asm passes the item as the affect
- [x] O0C2:371B `fxCursed` (spells1.go)
- [x] O0C2:3754 `fxBlinking` (spells1.go)
- [x] O0C2:3782 `fxFireball` (spells1.go)
- [x] O0C2:387B `removeAffectFromSide` (spells1.go)
- [x] O0C2:395C `fxHaste` (spells1.go)
- [x] O0C2:3986 `strikeSquare` (spells1.go) — groundInfo indexed %64
- [x] O0C2:3A44 `boltPath` (spells1.go) — who/ground reset per outer pass (asm keeps them; only differs if a bounce lands on the caster)
- [x] O0C2:3CD9 `fxLightning` (spells1.go)
- [x] O0C2:3D2E `fxSlow` (spells1.go)
- [x] O0C2:3D68 `fxRestore` (spells1.go) — bestClass 0 when nothing qualifies (asm: unset local)
- [x] O0C2:3F2C `fxSpeedy` (spells1.go)
- [x] O0C2:3F6D `fxCureSerious` (spells1.go)
- [x] O0C2:3FB7 `fxStronger` (spells1.go)
- [x] O0C2:403C `fxBolt` (spells1.go)
- [x] O0C2:408B `fxParalyzed` (spells1.go)
- [x] O0C2:40C2 `fxHeal2d4` (spells1.go)
- [x] O0C2:411C `fxInvisible` (spells1.go)
- [x] O0C2:414A `fxDamage2d4` (spells1.go)
- [x] O0C2:4185 `fxDamage2d8` (spells1.go)
- [x] O0C2:41DA `fxNeutralizePoison` (spells2.go)
- [x] O0C2:42D0 `fxPoison` (spells2.go)
- [x] O0C2:435A `fxSticksToSnakes` (spells2.go)
- [x] O0C2:43F0 `fxCureCritical` (spells2.go)
- [x] O0C2:4431 `fxCauseCritical` (spells2.go)
- [x] O0C2:4479 `fxDispelEvil` (spells2.go)
- [x] O0C2:44DA `fxFlameStrike` (spells2.go)
- [x] O0C2:451A `fxRaiseDead` (spells2.go)
- [x] O0C2:45FD `fxSlayLiving` (spells2.go)
- [x] O0C2:46BB `fxEntangle` (spells2.go) — amount of spell 136 as constants 6+3/level (DS:4060)
- [x] O0C2:4777 `fxFaerieFire` (spells2.go)
- [x] O0C2:47A4 `fxInvisibleToAnimals` (spells2.go)
- [x] O0C2:47DC `fxCharmMonsters` (spells2.go)
- [x] O0C2:486D `fxConfusion` (spells2.go)
- [x] O0C2:4931 `fxDimensionDoor` (spells2.go) — O121:0572 redraw skipped
- [x] O0C2:4AE2 `fxFear` (spells2.go)
- [x] O0C2:4C42 `fxFireShield` (spells2.go)
- [x] O0C2:4DF6 `fxFumble` (spells2.go)
- [x] O0C2:4F0B `fxIceStorm` (spells2.go)
- [x] O0C2:4F4E `fxMinorGlobe` (spells2.go)
- [x] O0C2:4F95 `createCloud` (clouds.go) — cloudkill: the first older ground that is neither cloud wins; the asm's repeat of square 9 after the loop changes nothing, not ported
- [x] O0C2:5550 `fxConeOfCold` (spells2.go)
- [x] O0C2:55F2 `fxFeeblemind` (spells2.go)
- [x] O0C2:569E `fxPlain` (spells2.go)
- [x] O0C2:56CC `fxPlain` (spells2.go)
- [x] O0C2:56FA `fxPlain` (spells2.go)
- [x] O0C2:5727 `fxSpell98` (spells2.go)
- [x] O0C2:57F1 `fxHeal2d4` (spells1.go) — cast_heal2 is the same bytes as cast_heal (40C2)
- [x] O0C2:585E `stoningGazeHandler` (spells2.go)
- [x] O0C2:5A5D `breathLightningHandler` (spells1.go)
- [x] O0C2:5BEE `spitAcidHandler` (spells2.go)
- [x] O0C2:5D24 `breathAcidHandler` (spells2.go)
- [x] O0C2:5F04 `breathFireHandler` (spells2.go)
- [x] O0C2:6091 `breathFireCloseHandler` (spells2.go)
- [x] O0C2:61B1 `throwLightningHandler` (spells2.go)
- [x] O0C2:62A4 `gazeParalyseHandler` (spells2.go)
- [x] O0C2:63B2 `isScroll` (spellmenu.go)
- [x] O0C2:63FF `scrollUsed` (spells1.go)
- [x] O0C2:648D `castMessage` (spells1.go) — drawing only
- [x] O0C2:658F `spellTableInit` (spells1.go)

## O0E3
- [x] O0E3:0014 `killWithMessage` (rules.go)
- [x] O0E3:00C7 `callAffect` (rules.go)
- [x] O0E3:010A `RemoveAffect` (rules.go)
- [x] O0E3:027A `auraAffect` (rules.go)
- [x] O0E3:027A `checkAffect` (rules.go)
- [x] O0E3:0414 `affectGroup` (rules.go)
- [x] O0E3:0D03 `cloudOwner` (clouds.go) — fixed: default is DS:6524 (list head), not the selected one
- [x] O0E3:0E47 `inPoisonCloud` (rules.go) — fixed: nausea handler gets the found affect (cough keeps the original nil)
- [x] O0E3:11DD `trapHits` (damage.go)
- [x] O0E3:1245 `AttackHits` (rules.go)
- [x] O0E3:12F1 `SavingThrow` (rules.go)
- [x] O0E3:1381 `Dice` (engine.go)
- [x] O0E3:13CC `RollDice` (rules.go)
- [x] O0E3:13F0 `AddAffect` (rules.go)
- [x] O0E3:14A7 `KillCharacter` (rules.go)
- [x] O0E3:1573 `removeInvisibility` (rules.go)
- [x] O0E3:15AB `downAffects` (rules.go)
- [x] O0E3:160D `releaseHeld` (rules.go)
- [x] O0E3:164D `CureAffect` (rules.go)
- [x] O0E3:16B0 `strengthCode` (stats.go)
- [x] O0E3:16D9 `strengthDecode` (stats.go)
- [x] O0E3:1728 `strongerThan` (stats.go)
- [x] O0E3:18D9 `statAffects` (stats.go) — fixed: ranger con bonus +1 level also when not dual-classed
- [x] O0E3:1FBA `damagePerson` (rules.go)
- [x] O0E3:22E1 `isUnaffected` (rules.go)
- [x] O0E3:23D5 `Heal` (rules.go) — fixed: wake-up handler gets flag 1 (now shares healPlayer)
- [x] O0E3:23D5 `healPlayer` (viewplayer.go)
- [x] O0E3:24C7 `Revive` (rules.go) — deliberate: O121:1356 icon redraw dropped

## O0ED
- [x] O0ED:0023 `weaponStats` (charcalc.go)
- [x] O0ED:021E `armourMove` (charcalc.go)
- [x] O0ED:02A6 `itemAC` (charcalc.go)
- [x] O0ED:0466 `ItemName` (sheet.go)
- [x] O0ED:0BDB `isHelpless` (charcalc.go)
- [x] O0ED:0C20 `ApplyItems` (charcalc.go)
- [x] O0ED:10CC `IntToStr` (char.go)
- [x] O0ED:117B `dexAC` (charcalc.go)
- [x] O0ED:11FD `dexMissile` (charcalc.go)
- [x] O0ED:127F `strIndex` (charcalc.go) — result undefined in the original above 18/100 or 25; 0 here
- [x] O0ED:1317 `strHit` (charcalc.go)
- [x] O0ED:13BA `strDamage` (charcalc.go)
- [x] O0ED:1520 `forgetSpell` (rules.go) — fixed: scans all 84 memorised bytes (was 21)
- [x] O0ED:1520 `forgetSpellOf` (affects.go)
- [x] O0ED:159C `loseItem` (eclcmds4.go)
- [x] O0ED:1690 `AddItemCopy` (items.go)
- [x] O0ED:172C `printMessage` (status.go)
- [x] O0ED:1788 `combatName` (combat.go) — drawing only, emits the text
- [x] O0ED:218A `hitMessage` (rules.go) — flashing left out
- [x] O0ED:2345 `FindAffect` (rules.go)
- [x] O0ED:23D5 `damagePlayer` (damage.go)
- [x] O0ED:24F7 `healedMessage` (rules.go)
- [x] O0ED:2564 `isParty` (combatmap.go)
- [x] O0ED:2564 `otherSide` (spells2.go)
- [x] O0ED:2588 `countSides` (rules.go)
- [x] O0ED:25E0 `nearEnemies` (combatmap.go)
- [x] O0ED:2708 `rangeTo` (combatmap.go)
- [x] O0ED:27E4 `endTurn` (affects.go)
- [x] O0ED:2833 `guard` (combat.go)
- [x] O0ED:286F `casterLevel` (combatmap.go) — classes other than 0..3 leave the result undefined in the original; 0 here
- [x] O0ED:2A29 `loadPicture` (pictures.go) — fixed: sets viewReset (b759A); drawing left out
- [x] O0ED:2E23 `SelectCharacter` (party.go) — front-end Question instead of the Home/End menu (selectMemberMenu keeps it)
- [x] O0ED:2E23 `selectMemberMenu` (eclcmds4.go)
- [x] O0ED:2FE4 `rangedWeapon` (combat.go)
- [x] O0ED:3027 `throwable` (combat.go)
- [x] O0ED:306C `ammoFor` (combat.go)
- [x] O0ED:3138 `atRaceLimitAfterRestore` (charcalc.go)
- [x] O0ED:335F `bandage` (combat.go)

## O0FE
- [x] O0FE:000F `calcSpellSlots` (classcalc.go)
- [x] O0FE:03C6 `recalcClassBonuses` (classcalc.go)
- [x] O0FE:0686 `wisdomSlots` (classcalc.go)
- [x] O0FE:07FB `savingThrows` (classcalc.go) — the old-class save row compares old monk level with monk level (0x118 < 0x110), kept as the original
- [x] O0FE:0AEA `thiefSkills` (classcalc.go) — fixed: the item-of-thievery bonus carries over from skill 2 to skills 3..8 as in the original; uninitialised bonus without that item kept at 0
- [x] O0FE:0D3E `mayTakeClass` (classchange.go)
- [x] O0FE:0EED `ChangeClass` (classchange.go)
- [x] O0FE:12AA `activeClass` (classchange.go)
- [x] O0FE:1306 `currentClass` (classcalc.go)
- [x] O0FE:1362 `firstClassLevel` (classcalc.go)
- [x] O0FE:13D1 `oldClassUsable` (classcalc.go)

## O104
- [x] O104:000F `nth` (listmenu.go)
- [x] O104:00DA `menuItemsOf` (menu.go)
- [x] O104:03D8 `DisplayInput` (menu.go) — keys via a front-end Question; the l72B0 timeout key (set only by START's title/demo) is not honoured, space returns 32 only as an option
- [x] O104:03D8 `combatInput` (combatturn2.go) — DisplayInput with no picture animation
- [x] O104:0E9F `ListMenu` (listmenu.go) — one front-end list question (no paging/cursor keys); header rows skipped as in the original
- [x] O104:11E4 `YesNo` (menu.go)
- [x] O104:1253 `newList` (listmenu.go) — GetMem nodes; n<2 gives one node as the original

## O110
- [x] O110:0103 `loadViewPicture` (pictures.go) — DAX decoding is the front-end's; spriteShown (b7292, the length byte of the name s7292) is not set true here but nothing reads it
- [x] O110:0556 `restoreBackground` (pictures.go)
- [x] O110:05D8 `loadPortrait` (pictures.go) — emits even when the head and body are unchanged (the original skips the reload)
- [x] O110:080C `bigPicture` (pictures.go)

## O116
- [x] O116:0165 `SetSkyColours` (geo.go)
- [x] O116:0573 `DoorAt` (geo.go)
- [x] O116:06A2 `WallAt` (geo.go) — directions other than 0/2/4/6 return uninitialised stack in the original; 0 here
- [x] O116:07A5 `MapCell` (geo.go)
- [x] O116:1025 `LoadWalldef` (geo.go) — a bad block stops the original game, here it is ignored; chunks counted by floor(size/780), the original copies a partial last chunk too
- [x] O116:130D `LoadGeo` (geo.go) — a wrong-size block stops the original game, here a message
- [x] O116:130D `loadGeo` (eclcmds3.go)

## O11C
- [x] O11C:0005 `sign` (combatmap.go)
- [x] O11C:0033 `nearSort` (cmap.go) — fixed: tie-break compared dirs the wrong way round (asm: dir[j] < dir[i] and dir[j]%2 <= dir[i]%2)
- [x] O11C:01A5 `newLine` (combatmap.go)
- [x] O11C:024C `step` (combatmap.go)
- [x] O11C:03F1 `lineOfSight` (cmap.go)
- [x] O11C:054A `inArc` (cmap.go)
- [x] O11C:08D8 `buildNear` (cmap.go)

## O121
- [x] O121:000F `footprint` (cmap.go)
- [x] O121:0077 `updateScreenPos` (cmap.go)
- [x] O121:0505 `mapAt` (cmap.go)
- [x] O121:0730 `combatInView` (cmap.go)
- [x] O121:0761 `combatOnScreen` (cmap.go)
- [x] O121:081B `recentre` (cmap.go)
- [x] O121:09DD `combatCentre` (cmap.go)
- [x] O121:0B3F `combatTurn` (cmap.go)
- [x] O121:0C32 `combatX` (cmap.go)
- [x] O121:0C5A `combatY` (cmap.go)
- [x] O121:0C82 `nearCombatants` (cmap.go)
- [x] O121:0CAA `combatIndex` (cmap.go)
- [x] O121:0D04 `groundAt` (cmap.go) — groundInfo indexed %64
- [x] O121:0E6F `combatRemoveFromMap` (cmap.go) — body search uses DS:70C7 for the asm's unset index; body count guarded
- [x] O121:115A `combatSquareFree` (cmap.go)

## R4CE
- [x] R4CE:0498 `DisplayTextBlock` (text.go) — word wrapping and the press-a-key pages are the front-end's (one EvText)
- [x] R4CE:0716 `InputString` (input.go) — front-end text question, filtered to 0x20..0x7A and upcased like the original
- [x] R4CE:08A7 `InputNumber` (text.go)
- [x] R4CE:093A `displayAndWait` (adventure.go) — the key wait (debug_txt) is left to the front-end
- [x] R4CE:0988 `Ticks` (engine.go) — virtual clock (Delay advances it); nothing reads it yet
- [x] R4CE:0ABF `flashMessage` (adventure.go) — fixed: emitted no text, only the delay
- [x] R4CE:0ABF `statusMessage` (status.go) — fixed: the gameDelay pause (R4CE:0B19) was missing
- [x] R4CE:0B19 `gameDelay` (adventure.go)

## R5BD
- [x] R5BD:054B `removeSubstr` (files.go)

## R670
- [x] R670:0000 `Exit` (engine.go) — fixed: had lost its exit (reseeded the RNG and returned); now ends the run with gameOver; the keyboard/file restore is the front-end's

## R699
- [x] R699:0020 `PlaySound` (sound.go) — EvSound to the front-end's driver

## R811
- [x] R811:24D1 `Randomize` (engine.go)
