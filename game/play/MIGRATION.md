# Moving to the typed contract (gobox/game/play)

Goal: the game is a typed, purely reactive interface. gobox reads `State()`,
switches over its type, offers the state's `Choices`, sends one `Action` to
`Do`, plays the `Outcome`, reads `State()` again. The game never acts alone.

## What goes away from game.Rules / game.State

| Old | New |
|---|---|
| `Step()` (game runs time, AI, scripts) | `EndTurn` outcome carries the monsters' moves; `AdvanceTimeBy` for rest/waiting; scripts advance on `Continue`/`Reply` |
| `Pending() *Ask`, `State.Pending`, `Ask`, `Answer`, `Cancel` | the state's `Choices`; gobox builds the dialog |
| `Refusal{Ask}` (door question, LootLeave "sure?") | plain error from `Do`; the follow-up actions are in `Choices` |
| `Mode`, `Window`, `Holding`, `Screen`, `Pacing`, `Flags` | the State's type (gobox picks screen and pacing) |
| `Read`, `Page.Prompt`, `PaneRows`, `Intro`, `Sequence` | `Reading` / `Talking` states; paging and timing are gobox's |
| walking | unchanged (user, 2026-11-05: gobox walks): `ExploreMap.Walk` is the game's `game.Map`; gobox sends `RunEvent{Before}`, `Arrive`, `RunEvent` and `Face` |

Old decisions with no action, because the choice is gobox's: `Answer`,
`Cancel`, `Select` (menus), `Items`, `Maneuver` (open a menu), `Wait`,
`Ranged`, `Aim` (now `Shoot`/`Cast` with the square), `Rest`/`Camp`/`LeaveCamp` (now `AdvanceTimeBy`, `MakeCamp`,
`BreakCamp`), `Save`/`Load` (now `Game.Save`/`Load`), `SetSpeed` (pacing is
gobox's), `SetIcon`/`Create` (now `SetChoice`), `KeepDraft` (`KeepCharacter`),
`AddChar`/`DropChar` (`AddCharacter`/`DropMember`), `Trade` (`Give`),
`TakeLoot` (`Take`), `Cure` (`Heal`).

## Built into the contract (2026-11-05 review)

1. **Own view types.** `play` has its own `Member`, `Battlefield`, `Unit`,
   `Walls`, `Draft` and the trading and looting fields, trimmed of
   everything that steered the screen (`Panel`, `Marks`, `Cursor`, `View`,
   `Log`, `Turn.Options`, `Member.Card`/`Powers`/`Chosen`/`Colour`,
   `Walls.Key`, dropped PNGs, ...). Only pure data types are reused from
   `game` (Picture, Item, SpellRef, Lore, Quest, Automap, Ground, Reach,
   Offer, Service, Training, Field, Where).
2. **Allowed per action.** `Common.Allowed` replaces the flat choice list:
   `Plain` holds every action with its parameters filled in, `Aimed` the
   square-taking ones with the squares, target count and area, `Moves` the
   reachable squares. No labels or hotkeys: those are gobox's.
3. **Scene kind.** `Scene.Kind` says text, talk or handout; text is plain
   strings (colours are gobox's).
4. **Fact-only outcomes.** `Outcome.Events` are `play.Event` types
   (`Attacked`, `Walked`, `Said`, `TimePassed`, `Refused`, ...) with no
   `Wait`, `Ms`, `Page`, `Pose` or `Screen`: gobox paces and shows them.

## Order

There is no bridge: every action and state is one explicit type in `play`,
and a game moves over whole.

1. gobox: a front path driven by `play.Game`, beside the old one; a game
   runs on one or the other.
2. COTAB first: its rules package implements `play.Game` completely, tests
   pass, its cmd switches to the new path.
3. SOTSB, POD, BR1, BR2 the same way, one game per change.
4. POR last and only if the user lifts its freeze; until then gobox keeps the
   old path for it.
5. When no game uses the old path, delete game.Rules, Decision and the
   flow fields from game.go.

## Open points

- A verb only one game has still gets its own type in `play`.
- Sound and music cues ride in `Outcome.Events`; where the music is set by
  the place, `Where.Music` stays state.

## Old fields and decisions, one by one (2026-11-05 gap pass)

What front/playrules maps each to, so a game on play loses nothing.

| Old State field | play |
|---|---|
| `Mode` | the state's type; Trading by `Kind` (shop, temple, training, vault) |
| `Pending` | built by gobox: Scene replies / typed / Continue, `SceneHeading` → Direction with Done (`Leave`), `ShipFight.Menu` → Choice, maneuver pick |
| `Door` | `Exploring.Door`; `TryDoor{Way}` |
| `Vault` | `Trading{Kind: TradeVault, Vault}` |
| `Shop.IdentifyPrice/Unknown/Identified` | `Trading.Identify/Unknown/Known`; `Reloads` from Allowed `Reload` |
| `Loot.Status` | `Looting.Status`; `LootDetect`, `LootShare`, `LootStatus` actions |
| `Intro` | `Scene{Kind: SceneTitle, Intro}` (the Amiga card is one of its Pictures steps) |
| `Sequence` | `Scene{Kind: SceneSequence, Sequence}` |
| `Results` | `Scene{Kind: SceneResults, Results}` |
| `Screen` | `Scene{Kind: SceneScreen, Screen}`; the space fight is the `ShipFight` state (`Fight`, `Sheet`, `Menu`) |
| `Page.Layers`, `Page.View` | `Scene.Layers`, `Scene.Placed` (Tone, gobox colours) |
| `Flags.Full` | `Scene.Full` |
| `Flags.Day`, `Walls.Sun/Mountains` | `Exploring.Day` (the game's say), `Walls.Outside`, `Walls.Mountains` |
| `Flags.IconDrop`, `IconEditor`, `Picker`, `PowerPick` | `Common.IconEditor/Picker/PowerPick` |
| `Sheet`, `SheetPage` | `Member.Powers/Reactions/Sheet`; which sheet is up is gobox's |
| `Creation.Step/Choices` | `Draft.Step/Choices` |
| `Member.ClassText/LevelText/SpellsReady/Memo` | the same on `play.Member` |
| `Combat.Turn.Moves/Casting/Weapon/Missile/Maneuvers`, `Unit.Pose` | `Battlefield.Moves/Casting/Weapon/Missile/Maneuvers`, `Unit.Pose` |
| `LastText` (Over) | `Over.Text` |
| `Where.Music`, `Map.NoMap` | unchanged (Common.Where, ExploreMap.Automap) |
| `Pacing` | gobox's: rest chunks (`AdvanceTimeBy` per Step, stopped by `StopRest` or anything happening), the auto-walk a square per Step, anim when a layer has frames; fight moments by the events' Ms/Pose |
| `Ink` | dropped: text colours are gobox's only |
| `Saves` | gobox's slot files over `Game.Save/Load` |

| Old decision | play action |
|---|---|
| `QuickStart`, `Quit`, `Modify`, `Rename` | the same |
| `Roll{Minimal}` | `Reroll{Minimal}` |
| `OpenPicker`, `Pick`, `ClosePicker`, `EditIcon`, `IconSize`, `IconColour`, `IconRemove`, `IconPNG`, `EndIcon` | the same |
| `UsePower` (`Sure` is gobox's ask), `PickPower`, `SetPolicy` | the same |
| `TryDoor`, `Mark`, `Unmark` | the same |
| `Plan`, `CampRest` | the same |
| `OpenTraining`, `CloseTraining`, `Learn` | the same |
| `LootDetect`, `LootShare`, `LootStatus` | the same |
| BR turn menu `Answer{N: 'T' 'V' 'I' 'A' 'S' 'D' 'J'}` | `Target`, `ViewTarget`, `Intimidate`, `Aid`, `Sprint`, `Dodge`, `Jetpack` |
| space fight crew menu `Answer{N}` | `Crew{N}` |
| overland/rocket Direction `Answer{N}`, Done | `Heading{Dir}`, `Leave` |
| `TitleDone`, `SequenceDone` | `Continue` |
| `WalkTo` | none: gobox walks the path (RunEvent{Before}, Arrive, RunEvent) |
| `StopRest`, `Rest` in camp | none: gobox's rest chunks |
| `Speak` | none: gobox plays the clips |
| `OpenSheet`, `CloseSheet`, `ShowMember`, `OpenItems`, `CloseItems`, `Items`, `Select`, `Browse`, `ViewMember`, `Resize`, `SetSpeed`, `TitlePicture/Credits/Flash`, `LootLook` | none: the screen's |
| `LootTake`, `Value`, `KeepAppraised` | none: `Take` and `Appraise{Keep}` say it |

New events: `SaveOffered` (old `EvSaves`), `ShipShown` (a ship fight moment, old `Event{Screen, Ms}`).

## Parity pass (COTAB, BR1, BR2, SOTSB diffs, 2026-11-05)

Which play field a game fills for each old State field, and what is gobox's.

| Old field | the game fills | gobox's |
|---|---|---|
| `Pending` over explore/combat/loot | `Common.Ask` (`Question`: text, replies, yes/no, typed, number with `Default`/`Max`); answered by `Reply` | the Ask's kind, keys |
| `Pending` party menu | `Assembling` with its `Allowed` (`NewCharacter`, `AddCharacter`, `RemoveCharacter`, `Modify`, `OpenTraining`, `ChangeClass`, `DropMember`, `QuickStart`, `Begin`, `Quit`) | the menu, labels (a series renames one in `Data.Texts`, e.g. "Remove Character": "Purge Character"), View Character, Load Saved Game |
| `Pending` creation steps | `Draft.Step` with `Choices`, or `Typed`/`Number` and `Default`/`Max`; answered by `SetChoice{Field: Step}` | the ask |
| `Picture` | `Common.Picture` (any state; exploring: `WindowPicture`) | the window |
| `Page.Pane`, `LastText` | `Common.Text` (the message lines; the last is LastText); Scene's `Text` adds to it | paging, colours |
| `Page.Prompt` | `Common.Status` (plain text) | colours |
| `Note` | nothing | the toast: refusals and `Refused` |
| `Map` (automap) | `Common.Automap` in every state (`ExploreMap.Automap` still taken while exploring) | |
| `Walls` | `ExploreMap.Walls` with `Side` set; no `Side`: no 3D view | `Key` |
| `Walls.Sun/Mountains`, `Flags.Day` | `Walls.Sun`, `Walls.Mountains`, `Exploring.Day` | |
| `Window` | `Common.Picture`, `ExploreMap.Overland` | the rest |
| `Flags` | `Where.Searching` (Sweeping), `Member.CanCast` | Adventure, Ordering (camp with `Order` allowed), IconDrop, Credits |
| `Holding` | nothing | from the state (door, shop, temple, loot, fight, story) |
| `Intro` steps, credits | `Scene{SceneTitle, Intro, Steps}` (each step's layers and tune) | stepping, the credits pages (Data), the flash |
| `Combat.Log`, `CursorOn`, `View` | the fight's `Narrated`/`Said` events | the log, cursor, view |
| BR turn verb keys | nothing | `Data.VerbKeys` ("Aid": 'H') |
| `Saves` | optionally `play.Slots` (the original's slots, listed after gobox's) | gobox's slot files |
| door ways | `TryDoor{Way: DoorExit/DoorBash/DoorPick/DoorKnock}` | |
| rest | allow `AdvanceTimeBy{D}`: any rest up to D (0: any) | the chunks |

## Round 4

| Need | play |
|---|---|
| Continue-only ask with Status shown | `Question.Continue` |
| Facing ask | `Question.Heading` (answer is `Heading{Dir}`) |
| Reroll yes/no with prompt | `Draft.YesNo`, `Draft.Prompt` |
| Team menu order and keys | `game.Data.PartyMenu` (default unchanged) |
| No hotkeys | `game.Data.NoHotkeys` (creation, reply, party menus) |
| Layered picture | `Common.Layers` (Window from Picture or Layers in every state) |
| Day sky outside exploring | `Common.Day` (`Exploring.Day` still read) |
| Add Character with empty roster | `AddCharacter{Roster: -1}`; gobox opens the roster and sends the pick |

gobox owns combat labels (Move from `Allowed.Moves`, Attack, Cast, ...), relabels via
`Data.Texts`, keys via `Data.VerbKeys`. A question over exploring is ModeAsk with
WindowScene, HoldNone, Adventure, and walls kept. When Continue is the only allowed
action, gobox asks Continue.

## POD pass

| Old | play |
|---|---|
| creation sheet figures (Age, ...) | `Draft.Member.Sheet` (adapter publishes it as SheetPage) |
| default rest "until study is done" | `Camping.Study` (gobox's default rest; 8h when 0) |
| save/load only where the rules allow | `Common.NoSave`, `Common.NoLoad` (zero: allowed); published as `Flags.NoSave/NoLoad`, honoured by the party menu and the slot screen (camp, F10) |
| save directory | `game.Data.SaveDir` (gobox still owns the files) |

Numeric sound/song names ("10") resolve as on the old path (front/classic turns them into
numbers). A free combat cursor (Aim with nothing pending) is gobox's; the game is not told.

NoHotkeys covers vertical list picks only (creation choices, the roster); bar menus keep their keys. gobox never invents a key twice in one menu (the later one gets none). A Scene list pick shows Common.Status as its prompt.

## COTAB extended parity

| Old | play |
|---|---|
| Story portrait (HEAD2 3 + BODY2 3), WindowPicture | `Common.Portrait{Head, Body}` (or pool `HeadID/BodyID`); gobox composes it |
| Attack with no adjacent foe (the core asks for a square) | `Allowed.Aimed` with `Attack{}` and squares; the answer is `Attack{At: &sq}` |
| the combat menu's free Aim | gobox's own: Turn.Options always carries `Aim`, the game is not told |
| Pending = the combat menu | not needed: gobox's front builds the menu from Turn.Options; Pending stays nil |
| the fight's message pane | `Common.Text` (when set, it is Combat.Log; else the Narrated lines since the fight began) |
| Flags.Adventure | camp, exploring, or any state with `Where.Mode` set |
