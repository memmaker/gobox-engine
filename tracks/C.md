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
