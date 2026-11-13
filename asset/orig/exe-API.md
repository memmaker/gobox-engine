# gobox/asset/orig (exe) API

Unpacks an EXEPACK-compressed Gold Box START.EXE (Turbo Pascal) so a game can read
its data tables at run time.

    exe.Load(path string, ds int) (*exe.Image, error)   // ds = data-segment paragraph
    type Image struct { Mem []byte; DS int }            // Mem = unpacked image, DS = ds*16
    (*Image).Bytes(ofs, n int) []byte                   // n bytes at DS:ofs, nil if out of range

Dropped (no callers): PString, Long, LoadCurse, LoadBlades.

## Per-game switch
go.mod: `require gobox v0.0.0`, `replace gobox => ../gobox-engine` and `replace gobox/rules => ../gobox-engine/rules` (one module: every package below imports as `gobox/<dir>`). Import
`gobox/asset/orig`; delete the game's exe/ dir (move its exe_test.go away, the package has it).

- sotsb-go: `exe.LoadBlades(p)` -> `exe.Load(p, 3554)`  (bladesDataSegment; used in cmd/tablegen)
- cotab-go: `exe.LoadCurse(p)` -> `exe.Load(p, 2750)`   (cotabDataSegment; cmd/tablegen)
  and sotsb-go cmd/tablegen also use LoadCurse -> `exe.Load(p, 2750)`
- por-go:   `exe.Load(p)` -> `exe.Load(p, 0xC7C)`       (=3196, old DS 0xC7C0 bytes; cmd/sounddata only uses Mem)
- cotab-go/sotsb-go cmd/sounddata: `exe.Load(p)` -> `exe.Load(p, 0)` (only Mem is used; DS irrelevant)

## tablespec and cmd/tablegen
`gobox/asset/orig` (tablespec.go): `Kind` (`Bytes Signed Words SignedWords Longs String(n)`, text form bytes|signed|words|swords|longs|string:N), `TableSpec{Name, At, Len, Kind, Doc, Cols, Img, Stride}` (json tags lower-case), `Generate(pkg, tables, imgs ...*exe.Image) ([]byte, error)`.
`Img` is the superset of sotsb's `Curse bool`: the index of the executable a table is read from. `Stride` (sotsb) is the row pitch when rows are padded. `Longs` and the "slot holds data, not a string" check (sotsb) are included.
`go run gobox/asset/orig/cmd/tablegen -spec tables.json -pkg engine START.EXE:DS [OTHER.EXE:DS] > tables.go`

## Migrate
- cotab-go: delete tablespec/ and cmd/tablegen/. Turn `tablespec.Tables` into tables.json (same fields; `{"kind":"bytes"}`, `"string:10"` for `String(10)`; all rows `img` 0). Run `tablegen -spec tables.json -pkg engine START.EXE:2750`. Output is unchanged.
- sotsb-go: same; a table with `Curse: true` becomes `"img": 1`; run `tablegen -spec tables.json -pkg engine BLADES-START.EXE:3554 CURSE-START.EXE:2750`.
- Either game may instead keep its list as a Go value and call `tablespec.Generate` from its own tiny main.
- Delete the game's `tablespec` import (`cotab/tablespec`, `sotsb/tablespec`); nothing at run time uses it.

A game whose own names clash with this API wraps them in a small adapter.
