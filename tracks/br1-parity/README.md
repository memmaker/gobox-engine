# BR1 parity: the old engine's tests against the shared engine

Audit of every `func Test...` of `buckrogers-go/rules/*_test.go` (62 files; `TestMain` in `main_test.go` is the
harness, not a test) against the shared engine, for the fold's end state (REFACTOR item 9: the old BR engines go).
`tests.csv` has one row per old test: file, test, what it asserts, status, the shared-engine tests that assert it,
a note. `engine/br1parity_test.go` (`TestBR1ParityTable`) checks that every `covered` and `port-done` row names test
functions that exist in this repository, and that every other row says why, so the table cannot rot.

Statuses

- `covered`: a shared test asserts the same behaviour (named in the row; a note says where it is only partly so).
- `port-done`: the behaviour was untested on the shared engine; a test was ported (or a fix made) in this track.
- `port-open`: the game has the behaviour, the shared engine lacks it or cannot be tested yet; the note says what it needs.
- `n/a`: the old engine's internals (its memory, decoder, front contract, own files) with nothing to port.
- `differs`: the shared engine deliberately behaves differently (a ruling or the Pascal); the note cites it.

Counts (125 rows): covered 65, port-done 22, port-open 14, n/a 17, differs 7.

## Ported, with the fixes it found

| old test | shared test | found and fixed |
|---|---|---|
| adv TestSidestep | TestBR1SidestepTwin, TestBR1WallBump | engine/br1par_walk_test.go; also BR1 only (BR2 not run) |
| adv TestSidestepTwin | TestBR1SidestepTwin |  |
| adv TestItemSlots | TestBR1ItemSlots, TestBR2ItemSlots | FIXED rulesxxvc/session.go itemCard: Slot was empty for every BR item |
| aid TestAid | TestAidOffered, TestAidBandages, TestAidByTheAI | FIXED rulesxxvc/ai.go: a computer-played team member treats patients first (combatai.pas AiMoveAttack:471) |
| aimed TestClassAimedSquares | TestBR1AimKeepsFighting |  |
| aimed TestClassSubAskKeepsState | TestBR1AimKeepsFighting |  |
| aimed TestClassStateStable | TestBR1NoSilentDo, TestBR2NoSilentDo | 20 builds compared in exploring, a fight turn and a shop |
| aimed TestClassSheetEverywhere | TestBR1SheetEverywhere, TestBR2SheetEverywhere |  |
| bigpic TestBigPicture | TestBR1BigPicture |  |
| jetpack TestJetpack | TestJetpackOfferedOnlyWhen, TestJetpackFlight | the rule was right; new rulesxxvc/par_jetpack_test.go |
| journal TestLogBook | TestBR1JournalEntries, TestBR1NewBlock | save/load of the read entries rides the game save (TestBR1SaveLoadExploring does not assert it): unverified |
| music TestAreaName | TestBRPlaceNames, TestBR1AreaName | FIXED: BR1 map showed "Map"; places carry Name (adventure.Place.Name, BR import version 18). BR2 still has no names |
| quiet TestClassQuietAnswer | TestBR1QuietAnswer |  |
| quit TestQuit | TestBR1TitleThenPartyMenu, TestBR2TitleThenPartyMenu |  |
| routed TestBugClass45Routed | TestBR1NoSilentDo, TestBR2NoSilentDo | FOUND+FIXED: Unready was a silent no-op (Engine.Unready sent rules.Ready, which only readies in BR), Ready/Drop/Unready in a fight panicked (rules.SetSpeed unanswered, combat.pas 273), Camp offered and panicked (CampOpen unanswered) |
| rules TestTitleThenMenu | TestBR1TitleThenPartyMenu, TestBR2TitleThenPartyMenu | the title is the manifest Start sequence; its pictures may be empty (compared equal) |
| spaceaid TestSpaceAid | TestShipAid | rulesxxvc/par_aid_test.go; the line is "is bandaged" (viewchar.pas 1004) |
| spacefight_e2e TestSpaceNotices | TestShipDyingCount | the texts and counts are ported; the old engine showed them as held status lines, the shared engine as narrated lines |
| sweep TestClassNoSilentAccept | TestBR1NoSilentDo, TestBR2NoSilentDo |  |
| throw TestThrow | TestBR1Throw, TestBR2Throw | FIXED engine/combatturn2_engine.go aimConfirm: Attack{Target: NoChar} for a bare square panicked; now Attack.At |
| title_import TestStartIsBR1Title | TestBR1TitleThenPartyMenu |  |
| wallbump TestClassWallBump | TestBR1WallBump |  |

## Not ported: what each needs

| old test | needs |
|---|---|
| adv TestRemove | needs the Change menu (explore.pas CampAlterMenu P = chargen.pas DropCharacterMenu) and the party menu Remove: a new screen and a ruleset roster verb |
| aid TestTrainHall | training not ported (rulesxxvc has no TrainFact/Train; chargen.pas TrainingMenu 1151): needs level-up skill points per level, a hall flag |
| aid TestTrainHallLeave | as TestTrainHall |
| bugclass TestClassItemActs | Trade/Halve/Join: the engine offers none of Give/Halve/JoinStack for BR (play.go items()) and rulesxxvc answers no GiveItem/Halve/JoinStack (viewchar.pas UseItemMenu): a ruleset verb set and an engine offer; Ready/Unready/Drop are covered by TestBR1NoSilentDo |
| deskew TestDeskewWestColumn | rules.Board.Straight is not built by rulesxxvc (rulesxxvc/ground.go): needs the unskewed dungeon arena of cmap.go in the BR arena builder, an engine option (Deskew) |
| exit TestClassEdgeFlagBeforeStep | the shared engine does it (engine/game_engine.go checkEdge) but a test needs a block with an open map edge: DevAt(81) in the shared import has walls on every edge square (Map walls 5; the old test read an open (15,0)); find which geo the script loads and the Pascal's edge squares |
| exit TestClassJoinedName | needs JoinStack/Halve for BR (see TestClassItemActs) |
| exitstep TestClassExitOneStep | as TestClassEdgeFlagBeforeStep |
| pack TestClassPartyPack | a ruling of the old engine (as POR pool); rulesxxvc has no PoolItems/PartyCap: needs a ruling for BR and a ruleset pack |
| savegame TestLoadSlot | the shared engine does not read an original BR save (SAVGAMx.DAT: the area/party/aux variables + CHRDATx files); needs a BR SaveLayout (as rules1e original.go) and an importer saveimport step; the old engine read and wrote the original format |
| spacefight_e2e TestSpaceMenuEsc | REFACTOR p2 Open: Esc in the Target submenu; needs the engine to answer Esc from Offer.Group |
| spacefight_e2e TestLootStatus | postcbt.pas:762: a Looting verb for the ship sheet and its text; a new screen state |
| view TestPurge | as TestRemove (the Change menu / DropCharacterMenu) |
| view TestTrain | as TestTrainHall |

Test names that also run for BR2 are the ones named `TestBR2...` beside their `TestBR1...` twin in the rows.
Unverified: the `covered` rows were matched by reading the old test and the named shared test's comment and name;
several are partial (the note says so).
