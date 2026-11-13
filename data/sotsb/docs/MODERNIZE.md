# MODERNIZE.md: from a faithful port to a modern Silver Blades

End goal: the game of `../por-go/MODERNIZED.md`, reached the way
`../por-go/docs/engine-redesign.md` ended up ("Rules, not screens"), but
taken there directly instead of through the screen stack.

## Keep (parity with the original matters only here)

- Texts, lore, story, characters (ECL scripts, journal, NPCs).
- Graphics as assets: pictures, sprites, combat icons, tiles, wall art,
  decorations, read from the DAX files.
- Maps, encounters, NPCs, monsters, items (GEO, ECL, MON*, ITEMS).
- The core rules: dice, THAC0, saves, damage, spells and affects,
  treasure, experience, training, resting, movement, ECL semantics.
- Music and sound effects (R7E4 sound driver).

## Replace

- UI/UX and controls, every screen layout, the resolution.
- The 3D exploration view (real 3D, Blobber renderer from por-go).
- Combat presentation (3D tabletop, from por-go `present/battle`).
- Ebiten, `ega`, the trace/snap harness, goplay.

## Shape

```
engine/   rules and game state only. No drawing, no keys, no timers,
          no goroutines. One entry point:
              res := eng.Apply(decision)   // synchronous
          res.State   the typed state (party, area, position, time,
                      combat, Pending: what the game waits for)
          res.Events  what happened, for the front end to animate
          res.Err     a refusal with a reason
present/  the front end on gpu/ (wgpu + glfw, copied from por-go): owns
          every screen, highlight, menu, pacing, animation, camera.
gpu/      none: the shared ../gobox-engine modules (gpu, input, blobber, combat).
cmd/sotsb the native binary: main loop on the main thread, calls Apply.
```

Pending is how the game content asks: a script menu / text / typed line /
member pick, a shop, temple or training hall a script opened, a fight's
player turn. The next decision answers it. Time only moves when a
decision says so (Step for combat rounds, Rest, Walk).

Parity is checked headless by Go tests that drive `Apply` with fixed dice
(the Turbo Pascal LCG and seed stay), plus the existing rule tests.

## Checklist

- [x] D0 plan and checklist (this file); goplay gone (volume_js, games.js
      rule, GOPLAY_* env, README/CHECKLIST lines)
- [x] D1 engine API: `Decision`, `Event`, `State`, `Pending`, `Apply`,
      refusals; headless driver `engine.Step` for tests
- [x] D2 no blocking input: ReadKey / KeyPressed / Delay / InputString /
      DisplayInput / ListMenu / YesNo / SelectCharacter callers become
      Pending + continuation; the engine goroutine goes. Done for the VM,
      adventuring, loot/vault, camp, party/creation/saves/training,
      shops/temples/sheet, appraise, Use, Detect, Modify Character, camp
      Alter/Display; combat left to gobox/combat
- [x] D3 script VM resumable: ECL suspends on menus, text, input, WHO,
      parlay, delay, combat, shop, temple, training, vault; resumes on the
      answer (scriptAsk/scriptSub, rules/eclrun.go scriptOn; text and delay
      are EvText/EvPause events the front paces, not waits. Dead leftover:
      ReadKey's blocked() reached only from aiKeys/flushKey, never live)
- [x] D4 adventuring as decisions: walk, turn, camp, search, look, area
      change, time passing, encounters (rules/eclrun.go exploreDecision; the
      full-map Walk in rules/walk.go, one real Step or Turn per paced step)
- [x] D5 camp, rest, memorize, scribe, cast out of combat, party order
      as rule calls
- [x] D6 party menu, creation, modify, add/remove, load/save (named slots,
      original file format), import from Curse, as rule calls
      (gobox front.Game: rules/frontgame.go over Call decisions; por's
      single creation sheet rules/creation.go, named slots rules/slots.go;
      Curse import is the "Add from Curse" entry; cmd/blades on front/present)
- [x] D7 sheet and inventory (ready, use, drop, trade, vault), shops,
      temples, training as rule calls (rules/frontitems.go on gobox's shared
      item, shop and temple screens; itemWords from START's table, checked
      against ITEM*.DAX; vault deposit (Pool O13C:0581, Drop); trimmed names)
- [x] D8 combat as a state machine: turn order, player turn decisions
      (move, attack, cast, use, guard, delay, flee, bandage, quick), AI
      turns stepped, typed combat events (engine/combatflow.go; left: View,
      Speed, the magic toggle, god_intervene, Move's Esc undo; a monster's
      delayed blast takes delay 0 where the original asked the player)
- [x] D9 drawing ripped out (text, menus, party/chargen screens go to gobox
      text renderer and gobox/front, por as template; no ui widget swaps before): ega, page/text cells, frames, boxes, menus,
      fonts as engine drawing, cmdbar, nav, mouse, iconeditor drawing;
      pictures and tiles are published as asset refs (file, block, frame)
      (208d6ad: present/, ega/ and the rules drawing deleted; indexed Pix)
- [x] D10 Ebiten, cmd/snap, cmd/trace and the trace hooks gone; go.mod
      on gpu/ (wgpu fork, gogpu, oto; only oto's ebitengine packages remain)
- [x] D11 front end skeleton on gobox/gpu + gobox/input (not copied from
      por-go; `replace gobox/<m> => ../gobox-engine/<m>`, see
      ../gobox-engine/DESIGN.md; menus our own until its window manager lands): window, input, text, menus from
      State, title, party menu, creation, explore with the 2D fallback
      (done: present/ + cmd/blades; explore is 3D. Left: the title picture,
      camp spells (memorize, scribe, cast), shop identify/appraise, portraits)
- [x] D12 3D exploration: gobox/blobber (textures and maps from WALLDEF/GEO), MAPS.md for SotSB, areas.py,
      make_map, wall textures (image model calls: ask before spending),
      compare sheets per area (View.Walls with each WALLDEF's own fronts,
      doors among them; floor and ceiling the original's colours; minimap and
      full map. Left: per-area compare sheets, remaster textures)
- [x] D13 3D combat tabletop: gobox/combat (a combat.Board from the combat
      State, Dungeon=true, WallTiles; its combat.Decision mapped onto Apply)
      (View.Battle and Turn from rules/frontviews.go; a won fight, its log
      and treasure seen in -shot captures. Left: multi-target aims, Ready)
- [x] D14 sound and music through gobox/audio from the engine's sound events
      (View.Events EvSound through present's driver; every id renders non-silent
      offline, accepted as checked. Title music: SOUND.DAX AdLib song on DOS,
      SECRET.MX song 0 on the Amiga (EvSong 2, rules.ForPlatform; not listened to).
      Amiga songs 1..6 at the Amiga program's moments (combat start, treasure
      after a fight, training, the party destroyed, big pictures 0x75 and 0x76;
      TestSongMoments). Open: the Amiga waits for song 1 before the fight
      (Sound(2)), not done; the Amiga effect mapping is inferred)
- [x] D15 por-go's UI extras where they fit SotSB (minimap, full map,
      party purse (rules.Purse), journal, slots, options) (minimap, full
      map with MAPNOTES labels and Walk, purse, slots; the journal's 72
      entries proofread from sotsb-re/gog/Adventurers Journal.pdf into
      rules/journal/JOURNAL.TXT, text only; showing them is gobox's)
- [x] D16 docs: README rewritten for SotSB, CHECKLIST/PORT_CHECKLIST
      folded in (see Origins)
- [ ] D17 rules implement gobox/game/play directly (no Decision, Event,
      EventKind, Mode, Result, no game.Rules adapter):
      [x] private key layer gone (input.go, aiKeys, flushKey, ReadKey);
      [x] dead code gone (CreateCharacter, ModifyCharacter, A..J Load/Save,
      Move{Dir}, auto-walk, View/ModeSheet, title steps, camp Drop/Speed);
      [x] passive ECL VM (frames + pending question); [x] passive flows (screen stack, no func)
      (creation draft is data; Call{F func} stays in the adapter until step 4);
      [ ] events, questions and waits on play types; [x] State/Do on the engine (Engine is play.Game; still built over
      the game.Rules adapter inside); [x] explore/camp/ask verbs, [x] combat/shop/loot verbs are play actions in Do;
      [ ] party and creation verbs, [ ] adapters deleted; [ ] tests ported

## gobox-engine

Guiding principle (user): a game is only a data holder and a rules engine.
Presentation, screens, input, drawing, sound playback, save/load and the app
shell belong in gobox; nothing new of that kind is built here (present/ and
the EGA-cell drawing get replaced by gobox front/ui). The game's package is
rules/ (was engine/); gobox/rules is imported as grules. Coins are one party
purse (grules.Purse).

Shared modules in ../gobox-engine (replace lines in go.mod): adf, amiga,
audio, char, combat, dax, ecl, ega, exe (tables.json), fixture, geo, gpu,
input, rules (signed saves, no encumbrance: user decisions), ui (fonts,
Flow), blobber. Pending there: the text renderer and gobox/front (main menu,
party, chargen). Map-edge wrapping stays ours (geo SolidEdges).

## Origins (from the old CHECKLIST.md)

The rules were ported procedure by procedure from the byte-perfect Turbo
Pascal reconstruction in ../sotsb-re (894/894 procedures, START.EXE and
GAME.OVR rebuilt byte-identical), over cotab-go's (and so por-go's) engine
shape. Each Go function names its SotSB address (Onnn:xxxx / Rnnn:xxxx).
Porting aids in sotsb-re: tools/procmatch.py (SotSB to CotAB pairs),
tools/dsmap.py (data addresses), tools/xref.py (work list and uncovered
procedures)..
Style: ~/notes/c-to-go.md (idiomatic Go, no emulation, no magic numbers).
