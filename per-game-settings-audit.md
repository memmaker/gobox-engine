# gobox per-game settings audit

Rule under test: gobox has no per-game switches. It holds no bools, enums or modes that games set differently to pick behaviour, no game ids, and no branches on which game is running. Game data passes through. Behaviour switches do not.

Callers checked: por-go, cotab-go, sotsb-go, pod-go, buckrogers-go (BR1), buckrogers2-go (BR2).
Findings are ranked by size (how much gobox code the switch drives, and how many games disagree).

## Violations

### 1. `game.CreationFlow` enum (`CreationADnD1` / `CreationADnD2BR`)
- Defined at `game/game.go:817-829`, field `Data.CreationFlow` at `game/game.go:1235`.
- Branches on it: `front/classic/creation.go:27`, `:41`, `front/classic/sheet.go:81`. It is passed in at `front/present/data.go:284` and `front/classic/game.go:51`.
- Who sets it: only BR1 sets `CreationADnD2BR` (`buckrogers-go/presentdata/data.go:16`). BR2 does not set it, so BR2 gets the Pool layout. Either that is a BR2 port bug, or the enum is not needed.
- This is a textbook mode enum: it is named after a game and selects a layout.
- Fix: derive the layout from the data the game already passes. Fields carry the order, and Stats/Derived say whether 18/xx and power picks exist. In practice: when the game's `Fields` say so, draw one field per row, and offer power picks only if the rules offer them. Check BR2's sheet against `buckrogers2-re` (creation screen) to confirm it matches BR1.

### 2. `Data.FreeAim` / `Data.NoFreeAim` (combat bar Aim)
- `game/game.go:1229-1230`, used in `front/playrules/playrules.go:837-838`.
- `FreeAim = true`: por, sotsb, pod. `NoFreeAim = true`: BR1. Neither: cotab (it gets Aim through `VerbKeys["Aim"]`) and BR2.
- These are two opposing bools that toggle one behaviour per game.
- Fix: unify. Offer gobox's free Aim exactly when the game's `VerbOrder` or `VerbKeys` names "Aim". That is data every Pool game already supplies. BR1 and BR2 name no Aim, so they get none. Then delete both fields. Check: in BR2's turn menu in `buckrogers2-re`, confirm there is no Aim entry, as in BR1.

### 3. `brVerbs`: Buck Rogers turn verbs and hotkeys hard-coded in gobox
- `front/playrules/playrules.go:932-940` (table). Used at `:810`, `:828` and `:958` (the Answer dispatch). The action types are commented "Buck Rogers' turn menu" at `game/play/play.go:594`.
- The keys T, V, I, A, S, D, J are BR's hotkeys, baked into gobox. The play actions are generic vocabulary (fine), but the label and key table is one series' data.
- Verdict: violation (game data living in gobox).
- Fix: move the labels and keys into BR1 and BR2's `Data.VerbKeys` and `Data.Texts`. gobox keeps only the actions, plus a generic fallback key from the label.

### 4. `rules.Config{TrapSkipStoned, TrapRollEach}`
- `rules/rules.go:33-36`, branches at `:248` and `:255`. The package doc at `rules/rules.go:1-4` even says "What differs per game is a Config flag". That doc line is itself against the rule.
- Who sets it: sotsb only (`sotsb-go/rules/rules.go:350`, both true). Everyone else passes the zero value.
- Fix: first check pod. POD is sotsb's engine lineage, so look in `pod-re` (or the pod decomp) at its trap damage routine for the stoned skip and the per-member roll. If pod matches sotsb and por/cotab do not, the games really differ: move `TrapHits`' damage loop into sotsb and pod (or give the shared function the per-member roll as data the caller computes). If por and cotab also do it, it is a port bug in them: make it unconditional. Then delete `Config` and rewrite the package doc.

### 5. `Data.CreditFrame` string mode (`""` vs `"edges"`)
- `game/game.go` (CreditFrame field). Branches at `front/present/pages.go:279` and `front/present/remaster.go:419`.
- `"edges"`: cotab, sotsb, BR1, BR2. Default `""`: por, pod.
- This is a string-valued mode switch.
- Fix: unify on one credits layout. The placed `CreditLines` (cells, colours) are already data. If por and pod show their credits differently in the originals, express that as placed lines and border data, not as a mode. Check por (`por-go/decomp`) and pod (`pod-re`) title credits screens. If they also draw one frame on the page edges, the default path is a port bug.

### 6. `Data.SaveNoConfirm` / `Data.SaveUnnamed`
- `game/game.go:1243-1244`, branches at `front/present/screens.go:139` and `:161`.
- Who sets them: cotab and BR2 (both true). por, sotsb, pod and BR1 leave them false.
- All Gold Box originals save to lettered slots A-J. Overwrite confirm and slot naming are almost certainly the same engine routine series-wide, so the disagreement looks like port drift.
- Fix: unify. Check the save routine in `por-go/decomp`, `sotsb`/`pod-re` and `buckrogers-re` for an overwrite prompt and a name prompt. Pick the original behaviour, make it gobox's only behaviour, and delete both fields.

### 7. `Data.NoHotkeys`
- `game/game.go:1240`, branch at `front/playrules/playrules.go:298-305`.
- Who sets it: BR1 and BR2. The Pool games do not.
- This is a series switch: "this game's menus have no hotkeys".
- Fix: move the information into the data the game publishes. `play.Question.Keys` already carries per-reply keys. Let a game that wants arrow-only lists say so per question (for example a `Keys` entry meaning "no key"), or have gobox never invent keys for lists the game marks `List`. Then drop the global flag.

### 8. `Data.Unmarked`
- `game/game.go:1228`, branch at `front/playrules/playrules.go:323`.
- Who sets it: sotsb and pod. por, cotab, BR1 and BR2 do not.
- This is a look toggle per game.
- Fix: check whether por and cotab mark the selected card in the originals (`por-go/decomp`, `cotab-re` party list draw). If all originals agree, unify. If they really differ, it becomes a theme/data property of the card art, not a rules bool.

### 9. `dax.Loader.OneDisk`
- `dax/dax.go:116-117`. It already carries a ponytail note saying it is to go.
- Who sets it: sotsb, pod, BR1 and BR2 in their engines. cotab sets it only in its cmd tools (fixture, geodump, walldump), not in its rules engine. por never sets it.
- Fix (as the note says): make "missing stays missing" the default, have por opt into the other-disk fallback only if its multi-disk layout really needs it, and remove the field. Also check cotab's engine: is its missing `OneDisk` a port bug? Compare with its own cmd tools, which say "as cotab always had it".

### 10. `Data.PointBuy`
- `game/game.go:1236`, read at `front/present/data.go:284` (`classic.PointBuy`) and `front/types.go:32`.
- `true`: por, cotab, pod, BR2. Unset: sotsb, BR1. Yet every one of the six games handles `play.PointBuy` / `game.PointBuy` in its rules.
- This is a bool toggling a gobox flow, and it is duplicated: the rules already answer the question.
- Fix: derive it from the rules. Offer Roll/Point Buy when the draft's `Allowed` contains `play.PointBuy`, and delete the field. Then look into why sotsb and BR1 handle the action but do not set the flag (likely dead code or a port bug).

## Borderline

### 11. `rules.Hooks.Hit` ("por's rules4e Hits")
- `rules/rules.go:29`, used at `:86`.
- The comment names a game, but the hook is a generic extra veto with the same meaning for every caller. `Begin`, `Groups` and `Group` are generic side-effect points (fine).
- Fix: reword the comment (drop "por's"). Keep the hook.

### 12. `classic.GameID` keys the shared portrait origins
- `front/classic/poolicon.go:98-115`, `front/classic/amigatwin.go:45`, set at `front/present/data.go:277` and `front/classic/game.go:51`.
- gobox looks up `origins.json[GameID]` and `[GameID+"-amiga"]`. The id selects data in a shared collection, not behaviour, but gobox does know the id string.
- Fix (optional): have each game pass its own origins map as data (`Data.IconOrigins`) and drop `GameID` from classic.

### 13. `Data.Series` / `present.Data.Series` ("pool", "buck-rogers")
- `front/present/data.go:36`, `:112`, `:279-282`.
- This is a folder name for shared asset pools, so it is data. Keep it, as long as nothing branches on the value. None was found.

### 14. `game.ScreenKind` (ScreenSpaceFight, ScreenShipSheet) and `play.ShipFight`
- `game/game.go:1396-1500`, `game/play/play.go:370-388`, `front/present/spacefight.go`, `front/present/shipsheet.go`.
- Only BR1 and BR2 use them. They are screens the game publishes and gobox draws (UI must live in gobox), and they are not toggles.
- Verdict: fine as UI. Strip the "Buck Rogers-style" wording from the type docs (`game/game.go:1396`, `play.go:370`) so they read as generic screens.

## Fine (checked)

- `play.Question.NoPick`, `NoEsc`, `Keys`, `List`, `Continue` (`game/play/play.go:127-132`) and `game.Ask.NoPick`/`NoEsc`/`Yes` (`game/game.go:899-901`): these are per-question semantics. por, cotab, sotsb, pod and BR2 all set them per question from their originals' ask calls. No game sets them globally.
- `play.Draft` (`game/play/play.go:179`): generic. Step and Ask are data.
- `Data.VerbKeys`, `VerbChords`, `VerbOrder`, `PartyMenu`, `Confirms`, `Texts`, `DaxMask`, `ReadyMark`, `MapCell`, `LabelInks`, `DoorInk`/`SelectInk`, `CreditGaps`, `TitleBackdrop`, `Border`, `Font`: all are tables or values the game loads (data).
- `geo.Map.SolidEdges` (`geo/geo.go:19`): each game computes it per area (por always true, the others by area block). It is per-map data. Only the comment names games. Reword it.
- `ecl.Default()` (`ecl/ops.go:27`): no game calls it or clears entries (grep finds no `ecl.Default()` in the six games). There is no override problem, but the table may be dead. Reword the "cotab and sotsb" comment, or delete it if unused.
- `combat` / `combat/grid`: no switches. The comments that name games cite decomp addresses where "all four agree" (fine).
- `blobber.Options`, `front/present.Options`, `app.Config`: player or host settings, not per game.

## Comment-only cleanups (game names next to generic code)

`rules/rules.go:1-4` (doc says "Config flag" per game), `rules/purse.go:15,46`, `front/view.go:587`, `front/present/*` (temple.go:21, sheet.go:88/124/127/173/181, invdraw.go:336, mainmenu.go:165, fight.go:214), `front/classic/*` (game.go:67/87/526/529/534, loot.go:24, party.go:40, partymenu.go:29, combat.go:74), `front/playrules/playrules.go:217,616,769,975,985`, `ui/text/list.go:5`. All of these describe data-driven behaviour, with the game named only as an example. They are not violations, but they invite the next per-game branch. Reword them to "a game whose ..." where it helps.
