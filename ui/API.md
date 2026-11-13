# gobox/ui API (window phase 2, first cut)

Imports only gobox/gpu. Tests need no window.

## Widgets (in -> out)
- ScrollBar{R,N,Rows,Top}: Thumb(), To(y)->top, Draw(fill,BarColors), DrawOn(dst,BarColors).
  MenuBar(rowRects,n,rows,*top,*sel,k)->BarHit; Bars: Reset/Add per frame, Click(x,y), Hover(y,down) -> taken.
- MenuRow{Text,Key,Do}; RowKeys, RowsOff, DigitKeys, HotText, VisLen, TitleLines, WrapWords, Ink* marks.
  MenuNav(act,key,*sel,keys)->(pick,back,ok). Act comes from the game's Table.
  MenuNavH: the same with Left=Up, Right=Down, for horizontal option bars (MenuNav itself unchanged).
- Table{Rows []Binding,Order,Combat}: Act(k), CombatAct(k), KeyOf(act), PadKeys(combat), HelpLines().
  Keypad(digit)->Act. The key table itself (codes, pad buttons, help text) is the game's data.
- Pager{P,Off}: Turn(PageSource,d)->closed, Key(act,src)->(closed,turned).
  PageSource.Page(p,off)->(page,pages,maxOff,ok). HelpPages(lines,rows) cuts help into pages.
- FullPageCells(bounds,minW,minH)->(k,cw,ch): geometry of a whole-window page.
- DrawText(img,glyph,s,x,y,c).
- AreaMap.Draw(dst,pic,tile,party,marker,k0); AreaMap.DrawIn(dst,rect,pic,...) the same clipped into rect of dst (centred as in a window of rect's size); AreaGeom(...) is the pure geometry.
- BIOSKey(gpu.Key)->(code,ok): the INT 16h table (Enter 13, Esc 27, Bksp 8, Tab 9, F1-F10 0x3B00.., arrows/KP 2468 0x48/50/4B/4D00, Home/End/PgUp/PgDn/KP1379, KP5 0x4C00, Delete 0x5300).
- input.Repeat(held,delay,rate)->bool (gobox/ui): held = ticks down (1 = first tick); true on 1, at delay, then every rate. Host sets delay/rate.
- ResultsRuns(Results,cw,ch)->[]Run{Text,X,Y,Role}; ResultsTitle(Results). Roles: RunGreen/Cyan/Red.
- Console{Cols,Rows,Cells,Col,Row,Theme}: NewConsole(cols,rows,Theme). Put/PutSymbol/Text/TextBg/Blank, Border(x1,y1,x2,y2,style,colour), BoxTitled(...,title), TextBlock(x1,y1,x2,y2,colour,clear,s,full) (cursor = Col,Row; full = the game's "press a key"), MenuBar(x0,y,menu,items,sel,normal,hi,black,lastCol). Cell{Set,Upcase,Symbol,C,Colour,Bg,Sym}; the game renders Cells with its own font. Theme{Border(style)->Frame{Corner,Vert,Horz}} is the game's border symbol ids.
- Flow(s,x1,y1,x2,y2,&col,&row,put,full), Wrap(s,x1,x2), Punct; MenuItems(menu)->[]MenuItem{First,Last}, IsHotkey; PadLeft, PriceLine(name,price,width), Commas.
- List{Top,Index,Rows,Total,Skip(i)}: Step, SkipRows, Move(fwd), Page(fwd), HasNext(), HasPrev(leadingHdr): the cursor of the scrolling list (sotsb listmenu). Drawing, keys and prices stay the game's.

## What por's remaster must change (adapters)
- Pixel output is meant identical (same arithmetic); run pages_test goldens after switching.
- scrollbar: Presenter.scrollBars/barDrag become ui.Bars; menuBar = ui.MenuBar + DrawOn(dst, BarColors{dark,gray,white,dark}) + Bars.Add.
- menuNav(k,sel,keys) -> ui.MenuNav(actOf(k),k,sel,keys); menuRow/rowKeys/digitKeys/hotText/visLen/titleLines/ink* -> ui.* (titleLines used engine.WrapWords; same function).
- bindings: keep the por table as `var table = ui.Table{Rows: ..., Order: {ctxMenus,...}, Combat: ctxCombat}`; rename fields (ctx->Ctx, keys->Keys, codes->Codes, padKeys->PadKeys, gap->Gap); actOf=table.Act, combatAct=table.CombatAct, keyOf=table.KeyOf, PadKeys/HelpLines delegate. KeyF*, memberCodes, ctx names stay in por.
- pager: keep por's pager struct (read, left) and embed/hold a ui.Pager; p.pages(v) -> a PageSource adapter calling eng.JournalPage / helpPage and mapping engine.Pages (Page, Pages, maxOff). closePager (left / ActOK), speak, syncRead, pageKey, openHelp stay in por. pagerKey: a := actOf(k); if kp := keypad(k)... -> ui.Keypad; call Pager.Key; on closed run closePager, on turned run speak. helpPages(HelpLines()) -> ui.HelpPages(lines, helpRows).
- fullpage: only the cell geometry moved (FullPageCells(b, invW, invH)); window border, sheetCanvas, noteLine, onBorder stay.
- mapimage: copy engine.MapSquares into ui.MapSquares (Seen, Wall enums map 1:1: NoWall,Wall->WallSolid,Door,LockedDoor); pass door = pal.Identity.RGBA(engine.DoorColour). drawText -> ui.DrawText.
- areamap: Presenter.areaSrc/areaImg/areaMark* fields become one ui.AreaMap; drawAreaMap -> a.Draw(dst, m.Img, m.Tile, m.Party, m.Marker, scale(dst)).
- results: engine.Results -> ui.Results; drawResults keeps fullPage/textAt/drawPageMenu, loops ui.ResultsRuns and maps Role to inkGreen/inkCyan/EGA 12.

## Left in por
drawMenu/drawMenuIcons (Presenter buffers, pal, blobber.MenuGlyph, border tiles), fullPage drawing, uiscale_test (blobber), helpPage (engine.View/Pages), pageKey/syncRead/closePager.

## Per-game switch notes (key repeat, BIOS table, rect map)
Copies found and diffed: por cmd/por biosKey (F1-F10, arrows, KP, Home/End/PgUp/PgDn, KP5, Delete, Enter, Esc, Bksp, Tab) == ui.BIOSKey exactly.
cotab keys.go keyCodes is the same codes for a subset (no F2-F10, KP5, Delete; adds nothing), plus a per-key repeat flag (Esc, F1 no repeat) which stays the host's.
Repeat differs per copy (report): por d==1 || d>=tps/3 && (d-tps/3)%(tps*3/20)==0 -> input.Repeat(d,tps/3,tps*3/20); cotab JustPressed || d>15 && d%4==0 -> input.Repeat(d,16,4) (same ticks); blobber default Repeats d>20 && d%4==0 -> input.Repeat(d,24,4) (same ticks; it adds KeyJustPressed itself). The three rates differ by design; the host picks.
- por: delete biosKey (use ui.BIOSKey; keep remaster.KeyF* consts), repeats() -> input.Repeat with padDelay/padRate.
- cotab: delete keyCodes' code column (loop gpu keys through ui.BIOSKey, keep its repeat set), pressed/padPressed -> input.Repeat(d,16,4).
- blobber: Repeats default can stay (por already overrides it).
- Wilderness pane: a.DrawIn(dst, paneRect, ...). Horizontal bars: ui.MenuNavH instead of MenuNav.

## Migrate
Console/List additions (golden tests in console_test.go compare against copies of the games' code).
- por: text.go punct/flow/Wrap -> ui.Punct/Flow/Wrap (Engine.textBlock = ui.Console.TextBlock with full = PromptRow24+ClearKeyboard, or keep its own cells and call ui.Flow). textCell may become ui.Cell.
- sotsb: text.go punct + DisplayTextBlock loop -> ui.Console.TextBlock (keep slowText delay via the Console's put: use ui.Flow directly with a put that calls DisplayChar and Delay); DrawBorder/DrawBoxTitled -> Console.Border/BoxTitled with Theme{Border: style -> Frame{b,b+1,b+2}, b=0x114+style*3}; menu.go isHotkey/menuItemsOf -> ui.IsHotkey/MenuItems, drawMenuBar -> Console.MenuBar; listmenu.go lmStep/lmSkipHeaders/lmMoveCursor/lmPageScroll -> ui.List (listTop = List.Top); shop spaces/padLeft/commas -> ui.PriceLine/PadLeft/Commas.
- cotab: text.go punct is unused (DisplayTextBlock only emits an event); menu.go isHotkey/menuItemsOf -> ui.IsHotkey/MenuItems. Its list menu is already a Question to the front end: nothing to share.
- Left in the games: ListMenu/DisplayInput loops (engine state, input, Trace), shop/sheet/camp/viewplayer screens (prices, stats, ECL state; only the small string helpers above were shared), por's text pane regions (texts.row).
Before: No API change. ui holds no colours, fonts or borders (callers pass them: `BarColors`, `ink`, `glyph`); only comments were reworded.

## Migrate (ui/fonts)
`gobox/ui/fonts`: `All() []Font`, `Get(name) []byte`, `Load(dir) []Font` (the game's own .f8 in dir first, so it can override a pool font, then the pool), `Find(fs, name) int`. Font = {Name, Bits}. The pool holds 58 fonts: the 56 PC BIOS fonts, `Gold_Box_8x8D1` (the Gold Box text font, identical in POR, cotab and sotsb) and `Dark_Queen_of_Krynn` (from por). Origins in fonts/INDEX.md; fonts/cmd/fontexport (own module) makes a pool file from a DAX block 201.
Each game selects its font from the pool by name:
- por-go: the game font is `Gold_Box_8x8D1`; `Dark_Queen_of_Krynn` is its extra. Delete engine/fonts/ and the embed/`pcFont`/`pcFonts`/`fontFS` (engine/fonts.go:19-37); replace the dir scan in `loadFont` (fonts.go:43-55) with `e.ext.fonts = fonts.Load(dir)` (dir `data/assets/fonts`, which may drop Dark_Queen_of_Krynn.f8); `pcFont{name,bits}` -> `fonts.Font{Name,Bits}`; select with `fonts.Find(e.ext.fonts, name)` (fonts.go:61). `e.font` (init.go:17, DAX block 201) becomes `fonts.Get("Gold_Box_8x8D1")`: glyph upper(ch)%64 is the same bytes at slot ch (0x20..0x7F), so the game-font path of glyph() (fonts.go:143-148) indexes `ch*8` with a-z folded to capitals (keep the bracket stand-ins at :137).
- sotsb-go: same replacement in engine/fonts.go:19-33 (embed), :41-42 (`PCFont` -> `fonts.Get`), :61-81 (loadFont: `fonts.Load(filepath.Join(e.L.Dir,"FONTS"))`, FONT.CFG lookup via `fonts.Find`). The game font `e.font` (init.go:139-148, 8X8D5 block 201) = `fonts.Get("Gold_Box_8x8D1")`; glyph() at fonts.go:99-113 indexes it by upper(ch).
- cotab-go: no font-loading code besides `core.Engine().L.Block("8X8D1", 201)` (cmd/cotab/main.go:618, field :54, copied to combat.font at :619, used by combat.go:31 glyph). Replace with `fonts.Get("Gold_Box_8x8D1")` and read `font[ch*8:]` (a-z folded) instead of `ch%0x40`. It has no Ctrl-L cycling; it can adopt `fonts.Load`/`Find` later.
- Keep FONT.CFG, cycleFont, Glyph, Note, the Ctrl-L keys and the "Pool of Radiance" label (game text, game-side).
- Requires `gobox/ui` in go.mod with its replace (ui pulls gpu/input; fonts itself imports only the stdlib).

## Migrate (text drawing, ui/fonts + ui.Console.Draw)
Games never render text; they pick a font and call gobox. `fonts.Glyph(bits, ch, text) ([8]byte, bool)` is the one glyph lookup (text: a-z folded, `[` `]` -> `<` `>`, upper(ch)%64 into Gold_Box_8x8D1; not text: raw tile, >= 177 none; a PC font `bits` shows ch as is for 0x20..0x7E). `fonts.DrawGlyph(rgba, g, x, y, &fg, &bg)`, `fonts.DrawText(rgba, bits, s, x, y, fg, &bg)` and `Console.Draw(rgba, bits, pal, symbol)` draw it; the game uploads the RGBA to its gpu.Image. `fonts.Chooser` (`NewChooser(dir, firstName, cfgFile)`, `Step(+-1)`, `Set(i)`, `Names(label)`, `Glyph(i, ch, text)`) is the Ctrl-L list + FONT.CFG; the game supplies only the dir, the font to list first and the label for its own font. Tests: fonts/text_test.go (por's glyph() copied as oracle), console_test.go TestConsoleDraw (por's glyphAt).
- por-go: done (engine/fonts.go only wraps the Chooser; page.go glyphAt calls fonts.DrawGlyph).
- sotsb-go: delete glyph() (engine/fonts.go:99) and the embedded font loading; `e.fonts = fonts.NewChooser(filepath.Join(e.L.Dir,"FONTS"), "", cfg)`, glyph lookups -> `e.fonts.Glyph(e.fonts.Cur, ch, upcase)`; the blit loops that turn a glyph into pixels -> fonts.DrawGlyph / Console.Draw.
- cotab-go: delete glyph() (cmd/cotab/combat.go:31), `combat.font` and the Block("8X8D1",201) read (main.go:618); text blits -> fonts.DrawText(img, nil, s, x, y, fg, &bg) / fonts.Glyph(nil, ch, true). Gold_Box_8x8D1 is the same bytes as its block 201.

Pure helpers (Console, Cell, Frame, Theme, MenuItem, List, IsHotkey, MenuItems, Punct, Flow, Wrap, PadLeft, PriceLine, Commas) live in `gobox/ui/text`, free of gpu and input; `ui` re-exports them as aliases. A game whose own names clash wraps them in a small adapter.

## Host window (shell.go, shellfront.go, shellmenu.go, prefs.go; was gobox/app)
- `Play(Launch, WindowMode, Capture) error`: opens the window and runs the started game until it ends: the one host loop every game runs in (cmd/gobox calls it). Window = `Launch.W x H` times `WindowMode.Scale`, locked to that shape in windowed mode; `WindowMode.Fullscreen` takes the screen's own resolution (ORed with `Settings.Fullscreen`), `Window` forces a window. `Capture.Window` (W, H points) opens a window of that size and any shape. The game draws the whole window every frame in device pixels.
- `Launch{Front, Title, W, H, Apply, Voice, SaveLoad, MainMenu, Art, FontPrev, TPS}`: a started game; cmd/gobox builds it (pool: `playPool`; a game outside gobox: its `play<game>.Play(adv, saves, seed, Dev) (game.Game, game.Data, error)` under `NewFront(NewGame(g, d), d)`, Apply `Tune(g)`). The shell reads keyboard, mouse, pads and No Grid's walk and look for `Front`, all through the latch (`Screen()` names the screen), BIOS codes, Ctrl-letter 1..26, Ctrl-Shift-F `FontPrev`, F12 `Art.Toggle`. `Arts{Want, Toggle, Set}`: two arts; Draw cross-fades (600 ms) when Want changes, then Set (nil: the Front's own `Arts()`, if any).
- Settings: one set for every game, `HostSettings` (`HostOptions`: the presenter's `Options` and the engine's), kept in `<user config dir>/gobox/SETTINGS.JSON` beside `Settings` (blobber's), read and written only by Play. Play loads them (missing keys keep the defaults), calls `Launch.Apply(HostSettings)`, and wires OnSettings (Apply, then an atomic save), the rows (`HostOptionRows`, the voice-over row with `Launch.Voice`), Repeats at TPS, Held through the latch, OnSaveLoad, OnMainMenu. A game outside gobox takes its part through `game.Tuned` (`Tune(game.Tuning{Deskew})`).
- `Capture{Shot, Keys, Freeze, Quit, Window, GPU, Slot, Pointer, PNG, Dev}`, `Scripted()`: a run is scripted when Shot, Freeze or Quit is set (the user's input is ignored). A key every 15 frames, waiting while the game is not Idle (at most 40 x 15); after the keys plus 4 more Shot is written (`%d`: every frame, numbered) and the run exits; Quit exits; Freeze stops. GPU goes to `gpu.SetConfig`. Slot is passed once to the Front's `Start(slot)`. A run with Keys never takes focus and stays a window.
- `Key` (uint16): a game key code or the script's own `ScriptRight`, `ScriptLeft` (a click at the pointer), `ScriptNext` (the pointer on), `ScriptWait`, `ScriptHold` (left button held: a drag), `ScriptSnap` (a mid-run shot), `ScriptTimed|frames`, `ScriptAlt|c`.
- cmd/gobox's `captureFlags` registers the capture and dev switches (`-shot -keys -freeze -quit -slot -pointer -png -size -fps -passes -novsync -offscreen -glitch -skip -cpuprofile -stacks -fight -freetrain -ecltrace -saves -at`); `parseKeys` reads `-keys` (`\M \C \P \w \B \S \t<ms> \A<c>`, `\r \e \t \U \D \L \R \G \H \F`, plus cmd/gobox's escapes \1..\8 \0 F-keys). No env vars.
