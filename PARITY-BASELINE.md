# Parity with por-go baseline c70cbee (Track P)

**Goal.** The one gobox engine keeps every player-visible feature and flow por-go had at c70cbee (2026-10-31), plus the decisions since (ARCHITECTURE.md). This lists them, where each lives today, and the title-sequence history per game for the declarative manifest sequences.

**How it was checked.** Static reading only: c70cbee's PORT_CHECKLIST.md "Extras"/"Idiomatic Go" and MODERNIZED.md section 7, then a grep per feature over gobox-engine and por-go HEAD (cmd/por, rules). No launches, no tests run for this list. "Present" means the code exists; it does not mean it was played.

Status:
- **G**: in gobox (front, front/present, front/classic, blobber, combat, ui, app, game). Done.
- **P**: only in por-go (rules or cmd/por). It works today, but it is **lost when cmd/por and por-go/rules go** (Track R) unless it moves to rules1e (rules) or gobox (presentation). Handoff to B3.
- **C**: changed by a later decision (commit given).
- **L**: lost.

## Title, main menu
| Feature (c70cbee) | Status | Today |
|---|---|---|
| Amiga painted title (DISK2/TITLE.PNG) under the main menu, lettering and functions fade in | G+P | gobox amigaimport/amigaimport.go, front/present/mainmenu.go (TitleBackdrop, fade); por-go rules/title.go |
| Key/click during the title's hold ends it, input goes no further (f260f61) | G | front/latch.go, front/title.go IntroStep |
| Credits page over 30 rows, remaster line at the bottom | G | Data.CreditBoxes (gobox 260eb02), RemasterCredit |
| Title music: piano recording, else PC-98 / Famicom (7d1c79b, 9120b05) | G | front/present/music.go, soundtrack.go |
| Windows exe icon (Amiga title) | P | por-go cmd/por/icon.ico, rsrc_windows_amd64.syso: move to cmd/gobox |
| Browser build (goplay, GOPLAY_DATA) | P / partly L | por-go rules/volume_js.go only; GOPLAY_DATA no longer read anywhere. Unverified whether a js build still compiles |

## Esc menu and options
| Feature | Status | Today |
|---|---|---|
| Esc menu: Resume, Save/Load, Gameplay, Graphics, Audio, Back to Main Menu, Quit (e78e2ba) | G | blobber/blobber.go (pages), app/menu.go |
| Graphics: Fullscreen, Dithered darkness, Light maps, Torch intensity | G | blobber/blobber.go:603ff, blobber/view.go Settings |
| Gameplay: Smooth / Fluid movement | G | blobber/view.go Settings.Fluid |
| Gameplay: Day sky by clock (6-20h, was always on; now option, ruling U3, default off) | C, P | por-go rules/eclops.go:166, rules/engine.go DaySky; host flag cmd/por/host.go:19 |
| Audio: Voice-over, DASP Sfx, Step sounds, Music/Effects/Voice volume | P | por-go cmd/por/host.go:13-26 (HostOptions); mixing gobox audio/driver.go (Pack) |
| Options kept in SETTINGS.JSON | G+P | gobox app/prefs.go; por-go cmd/por/host.go |
| Fonts Ctrl-F cycle, FONT.CFG, FONTS/*.f8 | G+P | gobox ui/fonts, front/present/fontmenu.go; por-go rules/fonts.go |
| UI cell size K, every screen fills the window, 16:10 window, fullscreen bars only for original pictures | G | gobox front/present (invdraw.go, uiscale_test.go), gpu/run.go |

## Exploration 2D/3D remaster
| Feature | Status | Today |
|---|---|---|
| Blobber 3D view, night sky, celestial, lights | G | blobber/ |
| Down steps backwards, facing kept; tap = one move, hold flows | G | blobber/view.go |
| Context-sensitive command bar, Cast only when able; Sweep ahead (W) / Search here (S) | G | front/present/remaster.go, cards.go; game/game.go |
| Card click opens View; card right-click menu (View, Items, Spells, Magic, Heal, Cure) | G | front/present/cards.go, dispatch.go |
| One member-pick screen for "whom?" (View.Choosing) | G | front/present/remaster.go |
| Party overview (P / pad X) | G | front/present/explore.go, input/input.go |
| Journal pager with the printed journal's pictures, voiced paragraphs | G | game/journal.go, front/present/journal.go; refs as script commands (por-go 38e0ef5) — C |
| Clue book | G | front/present/cluebook.go |
| Voice-over (Kokoro, VOICED.GOB, once per save) | P | por-go rules/voice.go, rules/story.go storySpeak |
| Bestiary (camp, BESTIARY.GOB) | G | front/present/bestiary.go, front/view.go Bestiary |
| Messages: notes flash, confirmations boxed, refusals as notes; toasts | G | front/present/fullpage.go, remaster.go |

## Minimap and full map
| Feature | Status | Today |
|---|---|---|
| Minimap (Tab / pad Y), unseen dim, light-brown doors | G | blobber/scene.go, blobber.go |
| Full map (M / minimap click): area name top left, cursor, legend, Mark/Remove, Walk | G | front/view.go FullMap.Name, front/present/fullmap.go |
| Labels kept per slot (LABELS.TXT), New Phlan's shops always labelled | P | por-go rules/fullmap.go |
| Explored squares per slot and per map (MAPS.GOB) | P | por-go rules/slots.go |
| Walk on the full map, stops at text/monsters/key | G | front/classic/walk.go, front/present/fullmap.go |

## Inventory, shop, loot, party
| Feature | Status | Today |
|---|---|---|
| One party purse | G | core Purse (gobox d995a8a), front/classic/party.go |
| Party inventory (60 slots, +10 bag) | P | por-go rules (pool) |
| Inventory screen: EoB paper doll, item atlas icons, tabs, tooltips, cursor stacks | G | front/present/doll.go, inventory.go, invinput.go, invdraw.go |
| Item names "(1d8)" "[7]" | P | por-go rules/items.go:85 itemStats |
| Shop Look (item rules box) | P | por-go rules/itemlook.go |
| Buy list: first usable item, buyer named, F1-F8, unusable grey | G+P | game/game.go Dim; presentation front/present/shop.go |
| Loot = inventory screen with the treasure as wares; F1-F8 pick taker | G | front/present/inventory.go, front/classic/loot.go |
| F1-F8 open member n's inventory; P toggles inventory/View | G | front/present/bindings.go |
| Quickstart six members | P | por-go rules/quickstart.go (agent C owns chargen) |
| Forget Saved Character | G | front/act.go ActForget, front/present/mainmenu.go |
| Roster: member ready to level in purple | P | por-go rules/train.go levelReady |

## Character creation (agent C owns; listed only)
Names keep case; minimal stats reroll with a count or refuse; paladins and rangers (CotAB rules); portraits: Shift-click next head, PNG drop becomes portrait; Quickstart. All in por-go rules/create.go + gobox front/creation.go, front/picker.go. Status: see Track C.

## Combat
| Feature | Status | Today |
|---|---|---|
| Classic full-width 12x7 view, status row, battle log (Log) | G+P | front/classic/combat.go, front/present/fight.go; log text por-go rules/combatflow.go |
| Move reach shown, click walks cheapest path | G+P | combat/battle.go Paths; por-go rules/combat.go moveReach |
| Free-attack (provokes) red squares in the walk preview | G+P | combat/types.go; por-go rules/combat.go provokes |
| Area-of-effect / single-target / bolt preview while aiming | P | por-go rules/aoe.go |
| Hit chance in the hover text | P | por-go rules/combat.go hitChance |
| Mouse aiming follows the pointer, a click on any target takes it (KeyHover, aimTake) | G | front/act.go ActPoint, front/ask.go |
| Remaster: Strike lunge, red flash/sparks, weapon-hit sound | G | combat/battle.go, front/present/combat.go |
| Remaster camera watches the turn, Tab/Shift-Tab jump to the target | G | combat/control.go |
| Bar tooltips with the readied weapon | G | combat/bar.go |
| Combat is action-driven (typed actions) | G | front/act.go, combat/ |
| Throw refused with "No target in range" | P | por-go rules (throwTargets) |
| 4e Combat option | G+P | core/rules.go, por-go rules4e |

## Camp, rest, training
| Feature | Status | Today |
|---|---|---|
| Camp: roster replaces minimap, F1-F8 member, bar lists only what they can do | G | front/present/camp.go |
| Rest: Heal (Fix), Memorize, Wait 1h, Night/Day | G+P | front/present/camp.go:176-181; rules por-go rules/camp.go |
| Memorize as one screen with a saved plan (NAME.mem); "Relearn the spells used?" | C | the plan replaced Relearn (por-go f4b2ff9 redesign M2) |
| Camp fire animates while waiting | unverified | |
| Party Order by dragging a card | G | front/act.go ActOrder, front/present/dispatch.go:478 |
| Training hall | G | front/present/train.go, front/classic/training.go |

## Saves
| Feature | Status | Today |
|---|---|---|
| Named slots on F9 anywhere, newest first, New Slot, load unwinds, atomic, no save in combat | G+P | front/slots.go, front/present slot screens; files por-go rules/slots.go, savegame.go |
| Party wiped returns to the party menu | unverified | |

## Sound and music
| Feature | Status | Today |
|---|---|---|
| Amiga music and effects (Paula) | G | amiga/sound.go |
| Extra music: song 6 on the overland map, song 4 entering Phlan | unverified | audio/driver.go plays songs; trigger not found by grep |
| DASP sound pack | G+P | audio/driver.go Pack; switch in por-go cmd/por/host.go |
| Player's own music (data/music/<trigger>) | G (new) | front/present/music.go |
| Sound data from START.EXE | Track A | por-go c383aa1 |

## Input
| Feature | Status | Today |
|---|---|---|
| One nav model on every menu; Enter confirms; Esc = Exit; Up/Down like Left/Right | G | ui/, front/present/menu.go |
| Mouse everywhere: hover, click, twice = Enter, wheel, right = Esc | G | gpu/input.go, front/present/dispatch.go |
| Gamepad: d-pad walks/moves, Y minimap, X overview | G | gpu/pad_*.go, ui/bindings.go |
| Latch: no input falls through a screen change | G | front/latch.go, TestLatch |
| Keys released on focus loss (OnFocus) | G | gpu/input.go |
| Keyboard help (?) grouped, paged | G | ui/pager.go HelpPages, front/present/pager.go:156 |
| 120 Hz input poll, same-frame result, held-key repeat | unverified | |
| Scripted captures POR_SHOT/POR_KEYS/POR_POINTER | C | flags (por-go 9767281: app.Flags -dev -seed -open -fullscreen), \C/\M clicks (gobox 5327723) |
| POR_DEV dev views | C | `-dev` flag (9767281); por-go rules/devviews.go |

## Lost
No feature was found missing from today's running por-go by this static pass. The real risk is the **P** rows: they live in por-go/rules or cmd/por and vanish with them. Gobox has nothing to restore on its own today; each P row needs a home:
- rules side (rules1e, B3): voice-over keys, labels/MAPS.GOB persistence, party pool, itemStats, itemLook, aoe, hitChance, moveReach, provokes, throwTargets, levelReady, day sky, battle log text, quickstart (C).
- host side (gobox app/cmd/gobox): HostOptions (Voice, DASP, Step sounds, volumes, DaySky) from cmd/por/host.go; exe icon; browser build.

## Title sequences per game (for the manifests' declarative sequences)
Shared gobox mechanics: play.SceneTitle with Steps map[int]play.TitleStep{Layers, Song} (game/play/play.go:286), front/title.go IntroStep, front/sequence.go; Data.TitleBackdrop under the main menu with zoom-in (aa97da4, a9678d1); credits as Data.CreditLines in the edges frame or Data.CreditBoxes (11080f6, 17a9a1b, 260eb02); NoCredits skips the step (29a10f9); first step's picture before the intro ends (38ce3a1); last picture stays under the credits (340daf5); click steps as Enter (dbe1d54); a VGA still is not an animation (59973e9, c2eed10).

- **POR** (por-go): Amiga title DISK2/TITLE.PNG under the menu, fade in (b77c46e, 3d86d7b); key/click ends the hold (f260f61); title paced by the presenter over TitlePicture/TitleFlash/TitleCredits/TitleDone (c13038e); logo cut from TITLE 2 (92db570); scene carries TITLE pictures 1 and 2 (bb7ab15; the first was black before); credits' two boxes as data ovr02:0206 (a513e76); title music: piano recording else PC-98 title (7d1c79b, 9120b05, d3dc927, 86675c6). Sequence: TITLE 1 (lightning flashes), TITLE 2, credits (two boxes, remaster line), Amiga painting + lettering under the main menu, piano or PC-98 tune.
- **COTAB** (cotab-go): SSI logo, credits, painting with the Curse logo, its Realms banner; banner goes, painting with logo stays under the menu (25bbfb9, f0632dd, 220aec2, acd5d01, d08d24a); credits on the in-game page with remaster lines (1c59bde), lower box under "based on" (9a1d1cb); full intro only when the config asks, else SSI logo only (3265a2d); title tune on the Tandy chip under Amiga effects (3aea597, gobox 1f11bc7).
- **SotSB** (sotsb-go): SSI (TITLE 1), credits page, painting (TITLE 2), painting + Secret of the Silver Blades (TITLE 2+4) as backdrop with the menu fading in (963791f, 0d9e742); Forgotten Realms banner TITLE 3 over the painting after it (467813a, reverses 963791f's drop); credits end on remaster lines (93c02b7), screen frame only (53a6487); title song EvSong 2 on the first step (11cbe1e, 543f2d0); Amiga music in its place (11cbe1e).
- **POD** (pod-go): intro animation (INTRO, P1) first; ending FINAL P3 on the finale's save-and-quit (ecf1324, a686b3b); then SSI screen, credits (TITLE 3), Forgotten Realms (1), painting (2), painting as backdrop with the menu (b4ff09d); TITLE pictures as whole-screen page layers (83a088e).
- **BR1** (buckrogers-go): title O030:0640 as steps of TITLE pictures where pR267_159A puts them, credits page pO030_02E1 (1729c44, b1d1989, 25fef1a); TITLE 4 behind the team menu as TitleBackdrop (3ef3a38, 7bbc8dd); gobox steps it (6ab56dd).
- **BR2** (buckrogers2-go): steps from pO025_0C36, both credits pages and TITLE 4 in its order (61d0301, 0ebc210); picture 4 as menu backdrop (47e9206); TitleDone to the team menu.

## Handover
Phase 1 done. Phase 2: nothing found lost in gobox's own scope; the P rows are moves for Track R (B3) and the host-options move to cmd/gobox. Next checks worth a targeted test: song 6/4 triggers, camp fire, wiped party to party menu, 120 Hz poll, js build.
