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
