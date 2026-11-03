# gobox/ui API (window phase 2, first cut)

Imports only gobox/gpu, gobox/input, gobox/gpu/vector. Tests need no window.

## Widgets (in -> out)
- ScrollBar{R,N,Rows,Top}: Thumb(), To(y)->top, Draw(fill,BarColors), DrawOn(dst,BarColors).
  MenuBar(rowRects,n,rows,*top,*sel,k)->BarHit; Bars: Reset/Add per frame, Click(x,y), Hover(y,down) -> taken.
- MenuRow{Text,Key,Do}; RowKeys, RowsOff, DigitKeys, HotText, VisLen, TitleLines, WrapWords, Ink* marks.
  MenuNav(act,key,*sel,keys)->(pick,back,ok). Act comes from the game's Table.
- Table{Rows []Binding,Order,Combat}: Act(k), CombatAct(k), KeyOf(act), PadKeys(combat), HelpLines().
  Keypad(digit)->Act. The key table itself (codes, pad buttons, help text) is the game's data.
- Pager{P,Off}: Turn(PageSource,d)->closed, Key(act,src)->(closed,turned).
  PageSource.Page(p,off)->(page,pages,maxOff,ok). HelpPages(lines,rows) cuts help into pages.
- FullPageCells(bounds,minW,minH)->(k,cw,ch): geometry of a whole-window page.
- MapImage(*MapSquares,cell,ink[16],doorColour,glyph)->*image.RGBA; DrawText(img,glyph,s,x,y,c).
- AreaMap.Draw(dst,pic,tile,party,marker,k0); AreaGeom(...) is the pure geometry.
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
