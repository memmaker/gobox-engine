# pod-go — Pools of Darkness: data and rules in Go

Source of truth: the byte-perfect Turbo Pascal reconstruction in ~/Projects/pod-re (GAME.EXE + GAME.OVR
rebuilt byte-identical). Base: sotsb-go's `rules` (Silver Blades is PoD's direct predecessor); every
function is checked against its PoD counterpart and changed to it. Scope: extracted data and rules, and (M9) the gobox front
end over them (cmd/pod). Rules: ~/notes/idiomatic-go.md.

Porting aids: `docs/proc_map.tsv` pairs each PoD procedure with its SotSB counterpart (682 of 933, 181
identical in shape; sotsb-re tools/procmatch.py); `docs/ds_map.tsv` maps PoD data addresses to SotSB
(tools/dsmap.py). Game data: `gog/` (extracted from the GOG installer, not committed) or $POD_DATA.

- [x] M0 repo: sotsb-go rules copied, module `pod`, builds without a front end (ui/text only)
- [x] M1 tables: every data table from PoD's GAME.EXE (DS paragraph 2535), named, by gobox tablegen — 56 mapped, 1 added (expSteps), 9 dropped; sky colours and item templates kept as Silver Blades literals in names.go until M4/M5
- [x] M2 drawing and blocking input out of rules: decisions in, State/Events/Pending out — ega/ and drawing gone; pictures, portraits, sprites as EvPicture/EvPortrait/EvMissile asset refs, EvQuit on exit, unconverted screens panic via blocked
- [x] M3 characters: creation, PoD record, saves, import from Silver Blades, training
  510-byte record (.pc/.thg/.efx, .sav in saves), Silver Blades .who/.stf/.sfx and
  savgam.dat party import (O06E:10BA, 203E), creation (O075:06DF, 21EE, 5478, 38C5),
  training (O075:4638, O0EC:38FA), hit points (O0AA:36BA, 3975, O075:441E), class
  change (O102:0E0C), savgamX.pty load/save (O06E:203E, 2989), POOL4.CFG. The game
  block p87F8 is e.game[0:1024] with PoD offsets (party count 31, training 47).
- [x] M4 world: ECL VM (PoD opcode set), adventuring, area data
- [x] M5 camp, resting, shops, temples, vault, treasure: camp/rest/Fix from O06B/O0B2 (no-study rests, Fix by
  known spells, rest encounters and difficulty in the game block), after combat from O04F (game bytes 20/28/29/30,
  stack merge, PoD exp values, no hireling share), money from O0B5 (Pool, platinum appraisal, gem/jewel items),
  temple from O04C (free Heal list with Restoration, Repair), shop from O052 (price factor game byte 39, platinum
  prices, pool pays the rest), vault file in PoD's layout, sheet/item names/paladin powers from O0AA/O0EC.
  ECL TREASURE's generator is PoD's random scroll (O055:1ABE); M11 dropped SotSB's item generator.
- [x] M6 spells and affects: PoD spell ids 1..126 and table (O0BB:6DFB), menus, casting,
  targeting and every O0BB handler incl. the new high-level spells and overlay hooks (652D..6B64);
  PoD affect numbers 1..127 with all O07C:32F6 handlers; O0E2 rules (saves, to-hit, damage,
  affect groups, clouds in inCloud); spell slots in spellslots.go (O102:0000, 0594, 06D7) from
  PoD's per-level total tables. Still SotSB: the cast
  animations (breaths, gaze 124 and summoning finished in M11; 117 is O0A1:4C33's)
- [x] M7 combat: setup, turns, AI, map — O05F (loop, turn order, old age, steps,
  attackTarget, quick), O062 (computer turns: turn undead, spells, moves, attack loop,
  flight, weapons), O0A1 (spell targets incl. area aim 4334, attack targets 3B4F,
  backstab, attack animation, morale, ally question, aim, god intervene "Helm"),
  O065 (DungCom/WildCom maps incl. river/trees/scatter, combat blocks, placement by
  monster sides DS:AEE8), O130/O135 (line of sight, near list, six-square footprints,
  bodies, 70 ground types). Sweeps removed.
- [x] M8 tests green against the GOG data: the whole-fight test's player walks round allies (it stepped into one
  forever), the camp test frees save A's full slots, Modify recalculates every slot (O075:332A), SETUP MONSTER
  sets the monster sides, vorpal/weapon sequence read attacksMade and rv.critical, PoD's one cloud list, dead
  Silver Blades leftovers removed (staticcheck U1000-clean).
- [x] M9 the gobox shape (gobox-engine DESIGN.md "The shape of every game", sotsb-go/por-go as templates):
  rules.NewContract(engine) is the game.Rules (rules/contract.go) that gobox's classic.Game builds every screen
  from; pod has no front code (the old rules/front*.go layouts are deleted); `presentdata` (Theme, Series "pool", Entries with
  "Add from Secret" = Silver Blades import, Rests, Bindings, box positions); `cmd/pod` wiring only on app.Flags
  (-data, -saves, -load, -ecltrace, -setup writes COVER.PNG); `import.toml` for gameimport; run.sh (gog/ or
  $POD_DATA) and build_n_run.sh. Screens still on `blocked` refuse with ErrNotConverted ("not now") instead of
  panicking. Not wired yet: walls in the 3D view (no relocated WALLDEF symbols in rules), the PoD border
  (BORDERS.DAX), sound/music, Modify and SetStat on the sheet, roll-to-minimums, the seen map, deleting a save,
  the full map legend, journal, icon editor, inventory/shops/temple calls (`errNotNow`). gobox gaps: GOBOX_GAPS.md.
- [x] M10 the front wired end to end: PoD's VGA DAX pictures decoded (pod/pic, EGA stand-ins); 3D view walls from
  the WALLDEF1 symbols (gobox draws them); the BORDERS.DAX frame (presentdata);
  seen squares and the full map legend from the player's marks, inventory, loot, shop, temple, camp and the
  creation sheet as contract states and decisions (rules/contract.go). Still refused: journal and clue book (PoD ships them only as PDFs), icon editor,
  DeleteSlot, WalkTo, LootDetect, UsePower, trade/halve/join (no front action), sound.
- [x] M11 last Silver Blades rules: SotSB's item generator and itemTemplates gone (PoD's TREASURE only makes
  O055:1ABE scrolls); delayed blast fireball bursts at once (O0BB:34AD, PoD has no delay); O0BB:5350 hooks with
  the charm gaze at 124 (it overwrote mind blank's 123) and the SotSB breath stand-in gone; O0E2:21FC gate in
  ported; rv.damage an int (PoD's signed word A7D8).
- [x] M12 gobox audio, VGA and inventory wired: present.Data.AdLib over MUSIC.DAX/ZOUNDS.DAX (EvMusic -> EvSong);
  PoD's pictures decoded with dax.ParseVGA, Pix + one shared Palette a picture (pod/pic deleted); inventory
  Trade (asks whom), Halve and Join as the rules' decisions, Readied with ReadyMark "*" (no Yes/No prefix).
- [x] M13 in-app fixes from the capture check: walls are PoD's VGA symbols since M12 (no noise); the slots page lists
  PoD's own games (it followed Add's source, Silver Blades by default); v/i open the sheet and items while
  adventuring (rules View in the explore and overland decisions); camp is over the 3D view (View.Adventure) with
  the members' View/Memorize/Scribe/Drop menus and a working rest counter; Save in camp (PoD's camp menu has it,
  F10's slots page, gobox -keys \0); the TITLE.DAX title (pictures 1..3). gobox gaps: GOBOX_GAPS.md.
- [x] M14 fight, loot and training check: Damage 0D0 came from ITEMS read past a false 2-byte header (every
  item field shifted); COMBAT with game byte 19 opens the temple and TREASURE SET the shop; the bar's End, units'
  damage, Aim asks for a square, Continue Battle is the board's yes/no; debug -fight/-seed for captures
  (capture/19..27). Training and a spell cast not captured in-app (capture/README.md).
- [x] M15 feedback and screens: Battle.Strike and EvDamage (victim) published; Turn.Can only (bar is gobox's);
  hall Train works (OpenTraining(-1), Becomes/Why); dungeon fights set DungeonTiles 0x19 and Battle.Dungeon;
  a fight's item pick is View.Inventory (no pod list). Spell cast not captured.
- [x] M16 game.Rules contract (rules/contract.go): State, Pending, Apply, Step over the engine; gaps in
  docs/CONTRACT_GAPS.md.
- [x] M17 contract: Save.Name (a slot letter); the automap's seen squares and marks owned by the engine
  (rules/automap.go, marked after each decision, unsaved as in PoD) and stated as Map.Known and Marked.
- [x] M18 contract: spell aims, Notes, imports (sotsb), Quick.All, member checks, Str00, item facts, Allowed,
  shop/temple/training facts; rules.GameData with a tip for every spell; a name keeps the new character.

## Deliberate deviations

- Saved games: the party purse trailer after the .pty layout stays (Go-only state).
- Character files (.pc and saved-game copies): an art trailer after the 510-byte record holds the four opaque art IDs gobox picks (portrait and icon, head and body). It is written only when an ID is set, so PoD's bytes stay identical, and files without it load with empty IDs.
- Save: the "Z" autosave, the temp-dir slot and the disk-free check are not ported.
- Encumbrance (O0EC:03E2) stays cut, as in sotsb-go.
- Fight level / sweeps: PoD has neither; none is ported (CombatInfo +0x05 is unused).
- Money (M5): one party purse stays; Take takes all the pool's money (PoD asks how much of each kind), Share is
  Take, Pool moves the purse into the pool. Coins weigh nothing, so no "Overloaded. Money will be put in Pool."
- Vault (M5): one vault.dat for every party in PoD's vault<slot>.dat layout (no per-slot file, no tmpvault.dat);
  gold under a platinum piece and copper in the pool are not kept.
- Temple and paladin (M5): decisions are their own confirmation, so "cast cure anyway?" is never answered yes and
  a service that is not needed casts nothing.
- Camp (M5): the game speed stays in Engine.gameSpeed (PoD keeps it in game byte 17).
- Appraise (M5): a kept gem joins an empty item list (the original loses it).
- ECL: divide by zero in ARITHMETIC leaves the target at 0 instead of faulting.
- ECL: ROB takes from the Go party purse once, not from each record's coins.
- ECL: the music driver (R5A0:0031) and digitized sounds (R5A0:0000) become EvMusic/EvSound events.
- ECL: AdvanceClock ignores a unit outside the seven counters (the original writes past them).
- Adventuring: the overland map takes any Turn of 1..7 eighths; the dungeon only 2, 4, 6.
- Adventuring: Silver Blades' locked-door Bash/Pick/Knock is gone (PoD has none); WHO's member menu is still blocked.
- LOSE ITEM (O0EC:18EF) is simplified: the item leaves the record, no readied-item bookkeeping beyond that.
- M7: draw-only code (O111, O12A, icon/flicker drawing, Delay) is skipped; the
  "attack ally?" question is a flow Ask before the action; undead count for turning is
  computed on the fly; combat facing keeps the original +4 for
  side 1; affect group 24 (grpCombatStart) is run but M6 defines no handlers for it.
- Clouds (M8): PoD's one cloud list (O0BB:2860, O0E2:22B0/237B, O135:14A9) replaces SotSB's per-kind lists;
  body grounds are PoD's 30/31.
- M7: wilderness terrain flags (O065:0723) are the constant 4 they are in PoD.
- Spells: Monster Summoning stops at map edges; PW Blind keeps PoD's unused duration roll; Dimension Door redraws skipped.
- Affects: trimMemorized lacks the extra magic-user allowance (bAE8A); constitutionSaveBonus is 0/5 outside 4..20 (PoD reads an unset local).
- Appraise uses the party purse; the whole-fight seed pin is re-pinned in M8 (4349ea32).
- M9: the creation sheet is PoD's own creation decisions: a field change discards and makes the character again
  (rolled anew); the name is typed on the sheet and given at Keep. Save slots are PoD's letters A..J.
- M9: a decision refused at an unconverted screen (ErrNotConverted) keeps what it changed before reaching it.
- M10: walking by the full map (WalkTo) is refused; map marks and seen squares live for the run, unsaved.
- M10: roll to minimums stops after 20000 rolls instead of on a key.
- M19: Quick Start: GameData().QuickParty holds six PoD picks; while the party is empty, the party menu's Pending offers "Quick Start" (Answer 0 = the QuickStart decision), which creates, rerolls to Need (at most 2000 rolls), saves and adds each pick.
- M20: QuickParty Head/Body are por/cotab portrait names, and the members keep the default combat icons. Pool names for DAX refs are gobox's (front/present, origins.json); pod states File/Block/Frame only.
- M21: characters store gobox's four opaque art IDs (Create portraithead/portraitbody/iconhead/iconbody on the draft, kept with the character; SetIcon with those fields via SetArt; Member.PortraitHeadID..IconBodyID), persisted in the character-file art trailer.
- M22: gobox composes portraits and combat icons and runs the pickers; pod had no compositing or pickers of its own (BuildCharIcon only loads the CBODY ref, the icon editor is a nil stub), so nothing was deleted. PoD's recolour rule is documented on BuildCharIcon.
