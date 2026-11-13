# In-app capture check (bin/pod, offscreen -shot)

Run one at a time, in the background (never taking focus): `GOBOX_BACKGROUND=1 bin/pod -data gog -saves <scratch copy of gog/__support/save/SAVE> -shot OUT.png -keys '...' [-load A]`.
T = `\w\r\w\r\w\r\w\w` (the three TITLE pictures cut short).
P = `\w\w\w\w\D\w\r\w\w\w\r\w\w\w\r\w\w\w` with `-load A` (answer No to Rolf, two more texts: adventuring).
`\F` is F10 (cmd/pod's escape); a real tab character moves camp's focus to a member's menu.

| Screen | File | Keys | Verdict |
|---|---|---|---|
| Title | 01_title.png | `\w` | OK: TITLE.DAX picture 1 over the whole window (then 2 and 3). |
| Main menu | 02_main.png | T | OK: Load Game, Party, Quit to DOS. |
| Slots page (Load Game) | 03_slots_load.png | T `\r\w\w` | OK: lists saved game A (fixed: it said "No saved games yet"). |
| Party menu (no party) | 04_party_entries.png | T `\D\r\w\w` | OK: Add, Create, Add from Secret, Exit. |
| Creation sheet | 05_create_sheet.png | T `\D\r\wc\w\w` | OK, readable, no grey square. PoD has no portraits. |
| Creation after reroll | 06_create_reroll.png | ... `r\w\w` | OK: abilities, HP and damage rerolled. |
| Load from the slots | 07_load_from_slots.png | T `\r\w\w\r\w\w\w\w` | OK: game A loaded, six cards, Begin Adventuring. |
| Walking | 08_walk.png | P `\U\w\w\U\w\w` | OK: walls are PoD's sandstone symbols (fixed by M12's VGA decoding). Remaster view: no frame or compass round it (gobox design). |
| Turn | 09_turn.png | P `\R\w\w` | OK. |
| Area map | 10_map.png | P `m\w\w` | Map, party marker, 11,2 OK. LABELS heading whole (gobox ce75e55). |
| Character sheet | 11_sheet.png | P `v\w\w` | OK (fixed: v did nothing). Damage was 0D0: ITEMS was read past a false 2-byte header (fixed in M14, now 1d8+6). |
| Items | 12_items.png | P `i\w\w` | OK (fixed): items with Ready/Use/Trade/Halve/Join/Drop. |
| Camp | 13_camp.png | P `a\w\w` | OK (fixed): cards and each member's View/Memorize/Scribe/Drop beside Rest/Exit. The member menus overlap the text box. |
| Camp member menu | 14_camp_member_menu.png | P `a\w\w<Tab>\w\w` | OK: focus on a member's menu; menus no longer below the text pane, but still cover its corners (gobox gap). |
| Rest (days) | 15_rest_days.png | P `a\w\wr\w\wm\w\wm\w\w` | OK (fixed): More raises the days (2). Days then Hours is the presenter's picker. |
| Rest done | 16_rest_done.png | P `a\w\wr\w\wm\w\wd\w\wd\w\w...` | OK: rested, back at camp, the text cleared. |
| Save slots in camp | 17_save_slots.png | P `a\w\w\F\w\w` | OK (fixed): New Slot and A, Enter=Save. |
| Save in camp | 18_saved_in_camp.png | ... `\D\w\r\w\wy\w\w\w\F\w\w` | Overwrite A answered, slots page again; the save itself is checked by TestSlotsAndCampSave. |

M14 (fight, loot, training). `-fight 1 -seed 1` are debug flags (cmd/pod): three of monster 1 at the first
adventuring prompt, fixed dice; the debug fight leaves 100 gold and an item. Q = 25 times `7\w\w\w6\w\w\w`
(Quick for a caster, Quick for a non-caster).

| Screen | File | Keys | Verdict |
|---|---|---|---|
| Fight start | 19_fight_start.png | P with `-load A -fight 1 -seed 1` | OK: board, Paine's turn, bar Move Aim Use Guard Delay Quick End. Status line cut at the bottom (gobox gap). |
| Aim | 20_fight_aim.png | P `2\w\w` | OK (fixed: Aim was refused): "AIM: W A S D, TAB; ENTER, CLICK; ESC". |
| Attack | 21_fight_attack.png | P `7\w\w2\w\w\r` | OK: the attack resolves on the nearest foe. |
| After the attack | 22_fight_after_attack.png | ... | OK, no damage feedback on the board (gobox gap). |
| Continue Battle | 23_fight_continue.png | P Q | OK (fixed: the Yes/No menu was invisible over the board): Y/N confirm. |
| Fight won | 24_fight_end.png | P Q `n\w\w\w` | OK: "The party has won", 930 XP each, treasure; View Take Exit. |
| Loot taken | 25_loot_take.png | ... `t\w\w\w` | OK: Take gone once everything is taken, View Exit left. |
| Creation field by key | 27_create_race.png | T `\D\r\wc\w\w<Tab>\w\r\w\w` | OK (gobox gap 1 closed): the first field's list opens. |

Not reached:
- M15 (P now ends at 11,2 facing S: P answers No to Rolf; walk route
  `\U\w\R\w\U\w\U\w\U\w\U\w\U\w\R\w\U\w\w\w`, one `\w` a step, reaches the hall):
  26_training_ask.png "WOULD YOU LIKE TO TRAIN?"; 28_training_hall.png the hall menu (`y\w\w\w`);
  29_training_train.png Train for Binky: "NOT ENOUGH EXPERIENCE" (`\r\w\w\w`).
  30_fight_cast_list.png Paine's CAST list (`-load A -fight 1 -seed 1`, P `\w\wc\w\w`; bar keys F C I G P U + M).
  31/32 (Down, Enter on the list) close the list without the aim prompt: the cast itself is not captured.
  33_fight_dungeon_walls.png stone wall cubes (DungeonTiles, Battle.Dungeon).
  34_fight_use_inventory.png the bar's Items key is `i` (the font draws I like 1; digits pick by position):
  P `\w\wi\w\w` opens the fight's item pick on gobox's View.Inventory. OK.
- Load while adventuring: PoD loads only from the main/party menu (F10 while exploring lists the games, saving off).
