# gobox/front API (first cut)

The shared front end: main menu, party setup, character creation, save/load
slots. Pool of Radiance is the template. Imports nothing but the stdlib so far.
The module owns the screens; the game implements `Game` (rule calls) and
supplies Theme and Data; the module never owns game state.

## In the module now
- `CellRect{X1,Y1,X2,Y2}`, `Pixels()`, `Framed()`: 8x8 text-cell rectangles.
- `Creation`, `Field`, `Option`, `FieldSex/Race/Align/Class`: the creation sheet's parts, filled by the game's engine (por: `engine.CellRect = front.CellRect` etc. are aliases, goldens unchanged).
- `Game`: the rule calls. Each returns a refusal error or nil (por's `*engine.Engine` satisfies it, asserted in engine/snapshot.go):
  - party menu: `Begin, Lead(n), RemoveMember(n), DropMember(n), NewClass(n, class)` (refused with a pick-a-class error), `QuickStart(), QuitGame()`
  - creation: `CreateCharacter(), ModifyCharacter(n), Roll(minimal, minimums), Rename, SetStat(i, typed), SetField(field, option), EndCreation(keep), CreationView() *Creation`

- `ActKind`, `Action` and the `Act*` kinds; `Intro`; Game's title calls (`TitlePicture, TitleCredits, TitleDone`).
- `Host`: `Call(func(Game))`, `Edit(func(Game))`, the rule-call queue a screen takes (por's Presenter implements it). Screens written against it in por: present/remaster/creation.go (createAct), title.go (introStep).

## Planned (not code yet; the screens still live in por's presenter)
Game gains the slots calls (`SlotsView(loadOnly)`, `Save`, `DeleteSlot`, `Load`) and the picker/icon-editor calls
(`OpenPicker, PickerView, Pick, ClosePicker, EditIcon, IconEditorView, IconSize, IconColour, IconRemove, EndIcon`),
once Slots/Picker/IconEditor views move into front.

```go
// Theme: the game's look, filled by por with today's exact values.
type Theme struct {
    Colours  map[string]color.Color // ink, hilite, fade backdrop dim ...
    Fonts    map[string]string       // role -> gobox/ui/fonts pool name
    Border   [..]ui.Frame            // border symbol ids (ui/text.Theme)
    TitleHold, TitleFade time.Duration // timing is data
    TitleDim float64
}
// Data: the game's content.
type Data struct {
    Logo     *front.Picture // the title picture's Pix, the lettering cut by present.Data.Logo (LogoCut)
    Backdrop image.Image
    Entries []Entry // extra party-menu entries, after por's default list
}
```

Menu entries: por's list (engine partyFunctions) is the default; a game appends entries via Data.Entries.
Screens expose a stable name (`main`, `creation`, `picker`, `icon`, `slots`) for the game's input latch; the latch stays in the game (por: cmd/por/latch.go, screenOf).

## Migrate
cotab and sotsb first need por-style rule calls in place of their blocking ListMenu/DisplayInput loops:
1. Replace each blocking party-menu / creation loop with synchronous, non-blocking rule calls (`Begin, RemoveMember, Roll, SetStat, SetField, EndCreation ...`) that return a refusal error or nil, and a `CreationView()` snapshot; no goroutines, no input reads in the engine.
2. Make the engine satisfy `front.Game` (add `var _ front.Game = (*Engine)(nil)`); use por's single in-place creation sheet, the rules answering which choices fit.
3. Fill Theme (colours, fonts by pool name, borders, timing) and Data (logo, backdrop, extra entries such as character import) from the game's own data.
4. Drop the game's own menu/creation drawing and input; keep its latch, fed with the front screen name.
5. Check: the game's golden dumps do not print package paths (aliases keep them identical).
