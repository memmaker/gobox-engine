# Track C: character creation (one flow)

## Decisions to copy into ONE-BINARY.md
- Creation, Roll (Felix 2026-11-11): `CreationRules.Roll(d *Draft, k RollKind, r *RNG) error`; a minimum it cannot meet is a typed `*core.MinimumError` (Ability, Want/Best with strength percent, Tries for rolls spent), never matched by text; the draft stays as it was; the creation screen shows it (POR's text; POR's special "STR 18/xx is for fighters only" now reads "STR 18/xx impossible here, max 18"). Implementers: gobox/creation's fake, por porCreation.
- coordinator (2026-11-11, for C6): a new character goes to the roster, not the party (baseline c70cbee: Keep only SaveCharacter's). Typed `core.RosterID`, a stable ID stored in the roster entry, never derived from the name (rule 7); Make of a new character returns it (Make has no callers outside creation). Quick Start joins through a rules call by RosterID (e.g. `Join(RosterID) (CharID, error)`); POR's roster files map names to IDs only when the import converts them.
- coordinator (2026-11-11, for C6): interim, until A8's DAX picture number -> pool lookup: POR's charOf carries PortHead/PortBody/IconHead/IconBody from the creation sheet's record (ponytail: in rules/porcreation.go charOf).

## C6 handover (agent C6, 2026-11-11)
Done (por-go rules, gobox core+creation; see the commit lines in agentC.log):
- Roll(d *Draft, k, r) error with `*core.MinimumError` (Felix). por rollMinimal returns it; setMinStats split out.
- POR Engine wired: CreateCharacter (creation.New), ModifyCharacter (creation.Modify), Roll, Rename, SetStat (creation.SetScore: out of range is now refused with porCreation.ErrOutOfRange instead of clamped; "18/5" one-digit percent refused by ParseScore), SetField, EndCreation (Keep = porCreation.Make) delegate to gobox/creation; creationEdit takes the draft from the record first (icon editor, PNG drops, DAX pickers and play.SetChoice pool picks still write the record: ponytail, goes when they write the draft), then rebuilds the record by charOf. setStatTyped deleted.
Open, in order:
1. QuickStart over creation.QuickParty: needs RosterID + Join (above); quickCreate and quickPick.good still the old path.
2. RosterID: core type, POR roster entry field (save format: the roster file gains it; the import converts old files), Make returns it.
3. When A9's core.PicID is in HEAD: retype Draft.Head/Body, core.Icon Head/Body, Character.Head/Body to core.PicID; the DAX lookup (game, DAX file, block) -> PicID replaces the charOf DAX-bytes carry. Pickers to the draft: PickHead/PickBody (pickgrid.go, B5 has uncommitted hunks there) and play.go SetChoice setPoolID to write cr.d.D.Head/Body/Icon (pool IDs only), icon editor to Draft.Icon; then creationEdit's draftOf pre-sync goes. Needs A8's lookup for DAX-only art.
4. done: copperPerGold reads goldbox.GoldBoxRates (gobox a5ab1d8).
5. cotab, sotsb, pod, BR1, BR2 (BR via rulesxxvc: message main, X4 owns it); then Creation() into CharacterRules.
Unverified: no full suite run; only rules tests matching Creat|Minimal|Roll|Quick|Modify|Sheet|Pick|Icon|Keep|Train, gobox core and creation.

## C7 handover (agent C7, 2026-11-11)
Done:
- [x] 1 baseline c70cbee restored (gobox 137d41a, por-go 1455421): SetScore clamps (ErrOutOfRange gone), "18/5" = 18/05, text that is no score keeps it silently, "STR 18/xx is for fighters only, max 18". TestSetScoreBaseline.
- [x] 2 one copy: pickers (PickHead/PickBody), PNG drops (icon, portrait), icon editor (EndIcon in creation), play.SetChoice write cr.d.D; creationEdit's and creationDone's draftOf(cr.c) copies gone (gobox 582b8ac, por-go f878174).
- [x] 3 PicID: Draft.Head/Body, Template Head/Body/IconHead/IconBody, Icon Head/Body, POR poolIDs are core.PicID. rules/porpics.go: PortHead/PortBody/IconHead/IconBody <-> PicID through e.pics (sprites.Pictures); charOf's DAX carry and quickPortHead/Body gone. A POR picture picked from the pool is stored as POR's own number now (same picture).
- [x] 4 RosterID (coordinator-approved): core.RosterID/NoRoster; .cha trailer key "rosterid"; save dir "rosternext"; BuildCharList numbers old .cha once (migrateRoster); Make -> (CharID, RosterID, error); Join(RosterID) on CreationRules (por joinRoster, linear scan); QuickStart = creation.QuickParty; quickCreate/quickPick gone. TestRosterIDs.
Open:
- [ ] 5 port cotab (rules/creation.go: its own sheet, saveDraft writes at once with no keep step: behaviour differs from POR, ask main), sotsb, pod; BR1/BR2 via rulesxxvc (X5 owns, message main).
- [ ] 6 Creation() into CharacterRules once all implement it.
Unverified: QuickStart's dice now go through creation.begin/Roll; the party stats for a fixed seed may differ from quickCreate's (not compared). Icon size from the draft overrides the race default in charOf (Enforce re-derives it on a race change). No full suite; gate passed (por/rules, gobox core/creation, games).

## C8 handover (agent C8, 2026-11-11)
Done (no commits, no code changes):
- [x] 1a QuickStart seed 7: full charRecord of all six members identical between c70cbee and f878174 (scratch exports; only the field name HitDice -> Thac0Base differs).
- [x] 1b A foreign PicID (no POR number) is kept: setPic puts it in poolIDs, the .cha PID2 trailer saves it, pickPic prefers it on read. Not a bug (code read, no round-trip test).
Open (asked main, waiting):
- [ ] 2 cotab/sotsb/pod port. Decisions needed: PointBuy (cotab, sotsb; not in CreationRules), cotab keepNew name prompt, sotsb joinDraft (party vs roster), sotsb defaultPortraits by sex, pod has no sheet (blocking picker plus starting spells/equipment). Text minimums ("STR INT ...") become typed during the port.
- [ ] 3 BR1/BR2 via rulesxxvc (X7). [ ] 4 Creation() into CharacterRules.

## C9 handover (agent C9, 2026-11-11)
Coordinator lines (main approved, 2026-11-11):
- coordinator: point-buy is creation's third roll kind: core.RollPointBuy, Draft.Buy map[Ability]int32 (the bought bases), core/pointbuy.go the one table (PointBuyBudget/Min/Max, PointBuyCost, PointBuySpent); gobox/creation Draft.PointBuy(bases) checks it (ErrPointBuy) and calls Roll(RollPointBuy); game/pointbuy.go forwards to core. B8 adds porCreation.Roll's RollPointBuy case (as Engine.PointBuy: setBaseStats from d.Buy, rollHP).
- coordinator: (a) the draft carries the training's dice as typed Draft fields (cotab/sotsb/pod roll trains to the starting experience: rollHitPoints -> trainSilently, per-level hit dice), so charOf replays the training without fresh dice.
- coordinator: cotab gets RosterID as POR (C7): id in the .cha trailer, old files numbered once on load; PoolIDs [4]string become core.PicID.
Done:
- [x] point-buy in core + gobox/creation + game forwarder (TestPointBuy in creation). Commit: see agentC.log (pending main's gate.sh PROJECTS fix when this was written).
Open, in order:
- [ ] cotab port. Design note: what to carry for (a). cotab's training is a pure function of the RNG state, the kind and the rolled stats: create.go makeCharacter -> rollAbilities, rollHitPoints (rollHitPoints2 at level 1, then TrainCharacter's Dice pairs per level, train.go:403-409). HitRolls []int32 needs a dice-injection seam in e.Dice. A typed `Draft.RollSeed` (the RNG state before the roll), with charOf replaying the roll and then applying the sheet's edits, carries the same thing with no seam. That is a different field from the approved one, so ask main which. Then: cotabCreation (Fields/Fits/Enforce over kind.enforce and creation.options, Roll with typed *core.MinimumError replacing Roll(minimal, mins string), SetScore over SetStat's clamps, Make = keepNew/modifyPlayer's Keep, Join by RosterID, Draft), state calls delegate to gobox/creation as por create.go does, EndCreation's empty-name keep follows c70cbee, saveDraft goes (POR's keep step). The iconedit.go, quickstart.go, slots.go, party.go and verbs.go readers of e.creation.c move to the draft.
- [ ] sotsb (roster + Join, default portraits by sex in Enforce), pod (POR's sheet in place of the list picker; starting spells and equipment into Make).
- [ ] BR1/BR2 via rulesxxvc (X7). [ ] Creation() into CharacterRules. [ ] delete the old paths and game/pointbuy.go.
Unverified: the full suites; point-buy has no game-level test yet, because no ruleset implements RollPointBuy.

## C10 (agent C10, 2026-11-11)
Done:
- [x] core.Draft.HitRolls []int32 (the starting training's hit dice, main's decision).
- [x] cotab rules/cotabcreation.go: cotabCreation implements core.CreationRules (Engine.Creation()); the hit dice of rollHitPoints/rollHitPoints2 go through state.hitDie, recorded into Draft.HitRolls by Enforce/Roll and replayed by charOf (no fresh dice); fresh = kindRecord + rollAge; SetField = setField(cr, ..), SetStat = setStat(c, i, st, x); Roll(minimal) = rollMinimal, a typed *core.MinimumError (the text matched cotab's); RollPointBuy over makeCharacter(bases); Minimums = d.Lows (no engine state). PoolIDs [4]core.PicID. TestCotabCreation: New and SetField draw exactly CreateCharacter's/SetField's dice and charOf rebuilds the same charRecord; an impossible minimum leaves the draft.
Open, in order:
- [ ] cotab RosterID (as POR C7: .cha trailer key, old files numbered once on load; Make returns it; Join). Make returns NoRoster and Join errors until then (ponytail in cotabcreation.go).
- [ ] cotab wiring: CreateCharacter..EndCreation, PointBuy, QuickStart delegate to gobox/creation as por create.go; readers of e.creation.c (iconedit.go, quickstart.go, slots.go, party.go, verbs.go, state.go:186/238) to the draft; saveDraft/removeDraft go (POR's keep step: main approved? C7 asked; not confirmed in ONE-BINARY.md); the keepNew name prompt vs errNoName.
- [ ] cotab DAX picture bytes (PortHead/PortBody/IconHead/IconBody) <-> PicID: cotab has no lookup (POR's porpics.go over e.pics); charOf keeps a member's bytes only, a new draft gets 0 (Quick Start's quickPorts bytes are lost on that path).
- [ ] sotsb, pod (same HitRolls seam), BR1/BR2, Creation() into CharacterRules, delete old paths and game/pointbuy.go.
- [x] CON replay: TestCotabCreationCon (cotab) types CON after the roll; the draft's whole charRecord equals the old SetStat path's.
Main's answers (2026-11-11), for the next agent:
- saveDraft: delete it if it is a private copy the draft replaces (as C7's creationEdit/creationDone copies); behaviour stays, say so in the commit.
- Unnamed new character: keep cotab's keepNew name prompt (baseline); if errNoName is used, it must lead to that same prompt.
- Pictures: cotab's DAX picture <-> PicID lookup over gobox/importer/pictures exactly as por rules/porpics.go; no second mechanism; if porpics' helper is generic, move it into importer/pictures and use it from both. (porpics is POR-table-specific except the DAXRef/e.pics.ID call; cotab needs its HEAD2/BODY2 table, see quickPorts.)
Unverified: no full suites.

## C11 (agent C11, 2026-11-11)
Done (one commit per repo, gobox + por-go + cotab-go; hashes in agentC.log):
- [x] 1 cotab RosterID as POR C7: Character.Roster is the GOBXID trailer's fifth entry (decimal; old readers skip it); save file "rosternext"; loadSaves numbers the .guy files without one once (migrateRoster, sorted); writeCharacter(.guy) gives one; Make returns it (keepNew); Join = joinRoster (mayJoin, AddToParty). TestRosterIDs.
- [x] 2 cotab wired to gobox/creation: creation holds d *creation.Draft, c is charOf(d) after every edit (creationEdit; Modify puts it in the party slot as before). CreateCharacter/ModifyCharacter = flow.New/Modify; Roll (mins text -> d.Lows), PointBuy, Rename (blank changes nothing, as before), SetStat (cotab's parseStat first, then d.SetScore), SetField delegate; EndCreation keep = d.Keep, errNoName -> the keepNew prompt (main's answer). saveDraft/removeDraft deleted: they were the file mirror of the draft; the end state on disk is the same (Keep writes the .guy, discard leaves none); only a crash mid-creation no longer leaves a file. state.go's fields read d.Fields/Fits/Picks; pool picks write the draft (draftPic), during the icon designer its copy (kept by EndIcon via iconOf). QuickStart = creation.QuickParty.
- [x] 3 pictures: POR's pickPic/setPic are gobox importer/pictures Pick/Set (Set takes the game's "none" number: POR 0, cotab noPicture since block 0 is a picture). cotab rules/cotabpics.go: PortHead/PortBody = HEAD2/BODY2 blocks, IconHead/IconBody = CHEAD/CBODY (iconBlock) via e.pics.ID; e.pics = pictures.GamePictures("cotab") (ponytail: ReadPictures once cotab runs from its folder). quickPorts gone from the code (the test keeps it as the expected numbers). Unlike POR, the draft reads only PoolIDs (the number is not read back as a pick; a new character's PortHead 0 is HEAD2 block 0, a real picture, and stating it would change what an unpicked character shows) and setPics keeps the pick in PoolIDs too, since cotab states only PoolIDs (memberOf).
- gate fixes: imports_test allows gobox/creation and gobox/importer/pictures; bugclass59 and TestStateVerbs expect ROSTERNEXT and no file before Keep.
Open, in order:
- [ ] sotsb port (as C10+C11 for cotab: sotsbCreation, HitRolls seam over train.go:319 and create.go:316, RosterID, wiring). Main decisions needed first (gameplay): setDraftField auto-joins on the class pick (joinDraft) vs POR's keep-to-roster; setStat refuses out-of-range (error) where POR/cotab clamp; Modify refuses race/class changes; "NO-ONE" placeholder name; defaultPortraits by sex.
- [ ] pod (no sheet; starting spells/equipment into Make), BR1/BR2 (X owns rulesxxvc), Creation() into CharacterRules.
- [ ] delete game/pointbuy.go (sotsb creation.go:224 and pod still call game.PointBuySpent) and front/present/pointbuy.go's use (switch it to core.PointBuySpent over a map[Ability]int32).
Unverified: Quick Start's stats for a fixed seed differ from the old path (fields enforced in sheet order). No full cotab suite outside the gate.

## C12 (agent C12, 2026-11-11)
Felix's decision (sotsb, via main): sotsb's creation is unified to POR's behaviour. A class pick no longer joins the party at once (setDraftField/joinDraft go): the character is kept to the roster, as POR. Out-of-range stats clamp (setStat no longer refuses). Modify allows race and class changes as POR. No "NO-ONE" placeholder name. Default portraits as POR (none: defaultPortraits by sex go). sotsb's creation is ported with cotab's hit-dice design (Draft.HitRolls recorded/replayed).
Done:
- [x] C11's three items committed (gate GATE_ALL=1 green): gobox f4c1e44, por a21965c, cotab c44f165.
- [x] (a) cotab Quick Start, seed 3: the new path drew other dice (sheet-order Enforce built the template kind at once). Enforce's first build now replays CreateCharacter (human fighter, rolled) then setField race, class, alignment, sex, the old order: the whole party's records and the final RandSeed equal the old path's (compared against an export of cotab c44f165^). TestQuickStart pins the seed (0x2ebd5e3c). Only difference left: the reroll cap is 2000 rolls (gobox quickTries), the old path's 1999.
- [x] (b, behaviour) sotsb as POR: setDraftField/joinDraft/defaultPortraits gone (a class pick is a plain setField; Begin is refused while the sheet is up, the draft is kept to the roster); setStat clamps; Modify's race/class change makes the member again in its slot (name, pool ids, items kept; discard restores the original pointer, creationSheet.orig); no NO-ONE (the name prompt keys on PName empty, creationSheet.named gone). TestSheetAsPOR; TestGameDraft/TestNewCharacterSheet updated.
Open, in order:
- [ ] sotsb port (Felix's decision above): sotsbCreation (as cotabcreation.go; the creation sits in a temp party slot, pickTemp/dropTemp, which charOf must keep for trainSilently), hit-dice seam over e.Dice during startingHitPoints (gold 2d6 and levelHitPoints both draw there), RosterID (sotsb has no trailer: the .ids sidecar's fifth line is the natural place), wiring of sheetCall/partyDecide/nameKept/play.go's Begin-joins-draft, game.PointBuySpent -> core.PointBuySpent.
- [ ] pod, BR1/BR2, Creation() into CharacterRules, delete the old paths, game/pointbuy.go, front/present/pointbuy.go -> core.PointBuySpent.

## C13 (agent C13, 2026-11-11)
- [x] 1 quickTries: the originals differ (all rolls counted): POR c70cbee 1+2000 (quickPick.good is true on its 2001st call), cotab c44f165^ 1+1999, pod maxQuickRolls 1+1999, sotsb quickTries 2000 in its loop (= 1+1999 over the draft's first roll). core.CreationRules.QuickRerolls() int (the rerolls after the first roll); gobox's quickTries is gone. POR 2000, cotab 1999, sotsb 1999. gobox de1aacb, por b2cc27a, cotab 4224c99.
- [x] 2 sotsb port (Felix's decision ad9c237, cotab's design): rules/sotsbcreation.go sotsbCreation (Engine.Creation()); the hit-dice seam is e.Dice while startingHitPoints runs (hitDice.on: the thief's 2d6 gold and levelHitPoints, in trainSilently too), recorded into Draft.HitRolls by Enforce/Roll and replayed by charOf; newCharacter = kindRecord (no dice) + startingAge; setField/setStat/roll's cores are remake/clampStat/rollMinimal (typed *core.MinimumError). RosterID: Character.Roster, the .ids sidecar's fifth line (roster.go; rosternext; Join = joinRoster over the .who files' .ids). Sheet wired: creationSheet holds d *flow.Draft, c = charOf after every edit (sheetEdit/showDraft), Modify's Adjust (sheetStep) and the sheet's pool picks (sheetPic) edit the draft; nameKept renames through the draft; quickStart = creation.QuickParty. Icon picks: kept drafts put a SotSB CHEAD/CBODY picture into IconHead/IconBody with no pool id (keepPics, pictures.Set over GamePictures("sotsb")); on the sheet they are PoolIDs only.
  Proof: TestQuickStartDice pins seeds 3 and 0x12345678 from the old quickStart (ad9c237): RandSeed after and the roster .who/.stf/.sfx bytes equal. Differences from the old path: the members join as loaded from the roster (Add's path, derived values worked out); the .ids hold the RosterID; a user's sheet pick of a SotSB icon becomes the number (as POR's pictures.Set) instead of a pool id.
  ponytail: no migration of older .who files to RosterIDs (only Join takes them, only Make gives them).
Open, in order:
- [ ] pod (rules/draft.go: rollDraft, buyDraft, setDraftStat, sheetVerb, modifyField, classPick; quickstart.go maxQuickRolls 2000 = QuickRerolls 1999; startingMageSpells/startingEquipment into Make; Need has "hp" and str00 keys).
- [ ] BR1/BR2 (buckrogers-go, buckrogers2-go rules/create.go sheetPointBuy over game.PointBuySpent; rulesxxvc).
- [ ] Creation() into CharacterRules (POR's is B11's, R5).
- [ ] 4 delete game/pointbuy.go: users left are front/present/pointbuy.go, pod draft.go:197, buckrogers(2)-go create.go:422; play.PointBuy.Scores []int (stat order) to a typed per-row core.Ability (main approved); sotsb's boughtOf goes with it.
Unverified: sotsb full rules suite ran once mid-work (66 s; the 3 failures it showed are fixed and their tests pass); the gate.
