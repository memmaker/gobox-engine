# game.Rules contract gaps

`rules/contract.go` implements gobox's `game.Rules` over pod's engine. This
lists only what still does not fit. `game.Decision` has an unexported
`decision()` method, so pod cannot add its own verbs: anything below stays
unreachable until gobox/game carries it.

## Notes on how pod fills the newer fields

- Shop.Offers: PoD rolls a gem's or jewel's value only when it appraises, so
  each offer (Item 0 the party's gems, 1 its jewelry, Member the selected one)
  has Value 0; the value arrives as the appraisal's EvText.
- Temple service prices and Training.Cost are 0: PoD heals and trains for free.
- Picture.Clear is left unset: the contract sends art refs, never Pixels.
- No pod wait offers an extra move beside its answer, so no Ask carries extra
  Options.
- SpellRules (`rules.GameData`): PoD has no tip texts, so every spell's tip is
  written from its spell table record (class and level, range, targets or area,
  amount, saving throw, camp or combat only, casting time). "Amount" is the
  record's duration-or-effect number; the record does not say which.
- Combat.Aims: Range is for the turn's caster; a hit dice spell (Sleep, Faerie
  Fire ...) states Targets as its budget's top (Faerie Fire the caster's level,
  else the dice's): PoD takes the squares sent while the rolled budget lasts.
- Item Text is the base record's damage, armour class bonus, range and hands;
  Kind and Slot follow the base record's slot (weapon, shield, armor, gauntlets,
  helm, belt, robe, cloak, boots, ring), scrolls and ammo.
- Doors: Kind is the wall kind, State the door: DoorAt 1 is DoorOpen, 2 and 3 are both DoorLocked. PoD's code (O12A:048A's callers in O065/O068) only tests the bits for 0 and 1, so it never tells a wizard lock from a lock; DoorWizardLocked and DoorClosed are unused.
- SpellRef.Range 0 is touch (the table's 255); the tips say "touch".
- Item.Scribe: a magic-user in action who does not know the scroll's
  magic-user spell (campScribe's tests, without "already scribing").
- Creation: a name (Create "name") keeps the draft in one decision; a field
  change on a rolled draft discards it and makes it again (PoD's way).
- TakeLoot.All applies TakeCoins, then TakeItem for each item until the pool is
  empty or an item is refused (the refusal comes back as an EvRefused event; what
  was taken stays taken).

## Decisions still with no clean mapping

- `game.Call`: refused. Pod's `Call` is a front hook (a function), not a rule.
- `game.Fix{Member}`, `game.Bandage{Member}`, `game.Repair{Member, Item}`,
  `game.PoolCoins{Member}`: PoD's are party-wide (the casters heal everyone, the
  first dying member is bandaged, the temple mends everyone, the purse is
  shared), so the member is only checked to be one (-1: none); Repair's Item is
  ignored.
- `game.Forget{Member, Spell}`: pod forgets all the member's pending spells; Spell
  is ignored.
- `game.LayOnHands` and `game.Cure`: the paladin is the selected member;
  Cure's Service is ignored.
- `game.RemoveCharacter{Roster}`: pod's removes party member N; Roster is read as
  that index.
- `game.Adjust` on the creation draft: refused. PoD rolls a draft and a field
  change rolls it again; Adjust changes the selected member (ModifyCharacter).
- `game.Cancel` where pod has no backing out (a question with no Exit/Cancel/No
  /Done/Quit option, exploring, a fight): refused.
- `AddChar.From` and `Load.From` take only "sotsb" (PoD's Silver Blades import)
  or "".
- `game.Save{Name}`: PoD names its saved games by letter, so a Name must be a
  letter A..J (it picks that slot); any other name is refused.

- Refused calls (PoD lacks the feature): Speak, SetPolicy, UsePower, PickPower,
  LootShare (the pool's money goes to the one purse), LootDetect (a Cast of the
  detection spell), StopRest, TryDoor (doors are ECL questions), WalkTo (no
  auto-walk), IconPNG and IconRemove (no PNG icons).
- DeleteSlot{Name}: Name is the letter A..J; it removes that saved game's file.

## State still not carried

- TempleName and Shop.Name: PoD's temple and shops have no names in the rules
  or its data; left empty.
- Automap Known, Marked and Notes are not saved: PoD's saved game has no map
  (O06E:203E).
- Automap Labels: PoD's maps have no labels of their own; left empty.
- Automap Terrain: the rules keep no overland terrain ids per square (the
  wilderness is the GEO map's walls); left empty.
- Bestiary: PoD keeps no record of the monsters met; left empty.
- Creation: Step is only "race" (no draft yet) or "name" (a rolled draft, to keep
  or roll again): PoD makes a character from all four choices at once.
- The intro animation (INTRO.ACF/BMA through VACE.EXE, with its SMS songs) and the ending (FINAL.ACF/BMA) are not stated: game.Intro carries only TITLE.DAX pictures 1..3, and gobox has no ACF/BMA player.

## Ask shapes

- AskMenu (horizontal) and AskList (vertical) both become Choice.
- Option hotkeys are the label's first letter or digit, not the original's
  per-menu keys.
- The front's own asks (move mode, confirm, item and aim pickers) are front
  state, not rules questions, and are not stated.
- Shop.Identified and Identifies stay empty: PoD neither remembers the last item it identified nor counts identifications.
- PartyCap is 0: PoD caps items per member (ItemCap), not for the party.
- Line.Hides is never set: pod states no PickItem lists.
- Service.Not holds the temple's own words ("is not blind."); the cures of wounds say nothing, so a member at full hit points has no entry.
- Item.Words leaves plurals alone (ItemName's plural endings are not applied); Sale is stated in a shop only, in gold.
- Member.Powers: a paladin's Lay on Hands and Cure Disease, aimed at himself until the decision names another member.
- Item.Scroll lists every spell on a scroll, with Scribing set by the top bit that marks a spell being scribed. Item.Spell is its first spell.
- Automap.NoteKinds stays empty: PoD's player labels have no kind.
- Beast.Danger and Icon, along with Automap.Terrain and the terrain/marker art, stay empty: PoD has no bestiary and no overland terrain ids.
- QuickParty portraits do not set the pod combat icon (pod has no HEAD/BODY portraits).
- Member.PortraitHeadID..IconBodyID are "" for old saves, imports and quick-start members. gobox falls back to the DAX refs (Icon numbers).
