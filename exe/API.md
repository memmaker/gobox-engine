# gobox/exe API

Unpacks an EXEPACK-compressed Gold Box START.EXE (Turbo Pascal) so a game can read
its data tables at run time.

    exe.Load(path string, ds int) (*exe.Image, error)   // ds = data-segment paragraph
    type Image struct { Mem []byte; DS int }            // Mem = unpacked image, DS = ds*16
    (*Image).Bytes(ofs, n int) []byte                   // n bytes at DS:ofs, nil if out of range

Dropped (no callers): PString, Long, LoadCurse, LoadBlades.

## Per-game switch
go.mod: `require gobox/exe v0.0.0` + `replace gobox/exe => ../gobox-engine/exe`; import
`gobox/exe`; delete the game's exe/ dir (move its exe_test.go away, the module has it).

- sotsb-go: `exe.LoadBlades(p)` -> `exe.Load(p, 3554)`  (bladesDataSegment; used in cmd/dsfind, tablegen, tablefind, dsbytes, itemdump)
- cotab-go: `exe.LoadCurse(p)` -> `exe.Load(p, 2750)`   (cotabDataSegment; cmd/tablegen)
  and sotsb-go cmd/tablegen, tablefind also use LoadCurse -> `exe.Load(p, 2750)`
- por-go:   `exe.Load(p)` -> `exe.Load(p, 0xC7C)`       (=3196, old DS 0xC7C0 bytes; cmd/sounddata only uses Mem)
- cotab-go/sotsb-go cmd/sounddata: `exe.Load(p)` -> `exe.Load(p, 0)` (only Mem is used; DS irrelevant)
