# What gobox lacks for pod-go

Written down instead of patched (gobox-engine is not edited from this repo).

- Sound: closed by gobox a4f9bd6 (`present.Data.AdLib`); presentdata sets MUSIC.DAX/ZOUNDS.DAX and the
  rules' EvMusic goes out as front.EvSong.
- Border: closed by gobox df88b7f (`Theme.Border.Pics`); presentdata sets BORDERS.DAX block 0's frames.
- VGA pictures: closed by gobox 33bde9d (`dax.ParseVGA`, `Picture.Palette`); pod/pic is gone.
- Inventory: closed by gobox 33bde9d (`Inventory.Trade/Halve/Join`, `InvItem.Readied`, `Data.ReadyMark`).
- Save slots: front.Slots is named slots; PoD has ten lettered games. pod maps names to letters A..J and refuses
  any other name; DeleteSlot has no counterpart and is refused.
- Creation picks by keyboard: closed by gobox's focus spots (Tab/arrows move, Enter/Space opens a field);
  capture/27_create_race.png: T `\D\r\wc\w\w<Tab>\w\r\w\w` opens the Male/Female list.
- Missing remastered map logged: closed by gobox 8239250.
- Full map heading: closed by gobox ce75e55 (capture/10_map.png shows LABELS whole).
- Camp member menus over the text pane: closed by gobox 2219a51 (menus stop above it). Left over: in camp the
  text pane sits mid-window and the members' menus still cover its corners and last line
  (capture/14_camp_member_menu.png).
- No compass and no frame round the 3D view: by design. The remaster draws the view over the whole window for
  every game (the BORDERS.DAX frame is used for its boxes).
- Combat bar letters: gobox's (bda7c87 unique hotkeys; Turn.Labels/Keys removed in bee4ac5); pod publishes Turn.Can.
- Combat feedback, aim cursor, status line, camp menus, log flood: fixed by gobox (8ca1bb2, ca29a04 and the
  rest); pod publishes Battle.Strike and EvDamage with the victim.
- The -keys walking that "did not register" was pod-side: the route assumed the 3,6 start, P starts at 11,2.
- Fight Use key: `1` went straight to the aim prompt in 34_fight_use_inventory.png, not via View.Inventory.
