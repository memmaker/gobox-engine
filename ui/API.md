# gobox/ui API (window phase 2, first cut)

Imports only gobox/gpu, gobox/input, gobox/gpu/vector. Tests need no window.

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
- MapImage(*MapSquares,cell,ink[16],doorColour,glyph)->*image.RGBA; DrawText(img,glyph,s,x,y,c).
- AreaMap.Draw(dst,pic,tile,party,marker,k0); AreaMap.DrawIn(dst,rect,pic,...) the same clipped into rect of dst (centred as in a window of rect's size); AreaGeom(...) is the pure geometry.
- BIOSKey(gpu.Key)->(code,ok): the INT 16h table (Enter 13, Esc 27, Bksp 8, Tab 9, F1-F10 0x3B00.., arrows/KP 2468 0x48/50/4B/4D00, Home/End/PgUp/PgDn/KP1379, KP5 0x4C00, Delete 0x5300).
- input.Repeat(held,delay,rate)->bool (gobox/input, no gpu): held = ticks down (1 = first tick); true on 1, at delay, then every rate. Host sets delay/rate.
- ResultsRuns(Results,cw,ch)->[]Run{Text,X,Y,Role}; ResultsTitle(Results). Roles: RunGreen/Cyan/Red.

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
No API change. ui holds no colours, fonts or borders (callers pass them: `BarColors`, `ink`, `glyph`); only comments were reworded.
