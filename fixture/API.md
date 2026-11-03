# gobox/fixture

Builds a save fixture (`SAVGAMA.DAT` + party `.SAV/.SPC/.ITM`) for a trace scenario.
Needs nothing but the stdlib. Never writes to the game's data, only to `-out`.

- `fixture.Run(g Game, args []string) error`: flags `-out -data -slot -block -walls -geo -x -y -dir -set VAR=VAL -strong -decisions FILE`.
- `Game` callbacks: `Root(data)` data dir; `Open(root, slot)` returns a `State`; `ECL(root, block)` disk + script (no 2-byte header);
  `Member(root, name, ext)` file bytes; `Strong([]byte)`; optional `ParseDecision(line)` (the game's `engine.ParseDecision`,
  used to validate `-decisions`); `NeedSlot`.
- `State`: `EnterArea`, `SetPos`, `SetVar`, `Bytes`, `Party`. por's `*engine.Save` already fits.
- `RawSave` / `OpenRawSave` / `MakeStrong`: the byte-level State for cotab and sotsb (same SAVGAMA layout).

cotab/sotsb `cmd/fixture/main.go`:

    sv, err := os.ReadFile("../SAVE/SAVGAMA.DAT") // then fixture.OpenRawSave(sv)
    fixture.Run(fixture.Game{Root: ..., Open: ..., ECL: <dax.NewLoader(root).Block("ECLn", b)>, Member: <read ../SAVE/name+ext>,
        Strong: fixture.MakeStrong, ParseDecision: ...}, os.Args[1:])

Test: `go test ./...` (temp dirs only; the fixture has no randomness, so no seed).
