# gobox/fixture

Builds a save fixture (the save, named by `State.File`: `SAVGAMA.DAT` or a canonical save, + party `.SAV/.SPC/.ITM`) for a trace scenario.
Needs nothing but the stdlib. Never writes to the game's data, only to `-out`.

- `fixture.Run(g Game, args []string) error`: flags `-out -data -slot -block -walls -geo -x -y -dir -set FIELD=VAL -strong -decisions FILE`.
- `Game` callbacks: `Root(data)` data dir; `Open(root, slot)` returns a `State`; `ECL(root, block)` disk + script (no 2-byte header);
  `Member(root, name, ext)` file bytes; `Strong([]byte)`; optional `ParseDecision(line)` (the game's `engine.ParseDecision`,
  used to validate `-decisions`); `NeedSlot`.
- `State`: `EnterArea`, `SetPos`, `SetVar`, `Bytes`, `Party`. por's `*rules.SaveEditor` fits; `SetVar` takes a field's layout name.
- `RawSave` / `OpenRawSave(sv, RawLayout)` / `MakeStrong`: the byte-level State; the game passes its SAVGAMA layout (offsets) as a `RawLayout`, with its script layout (`Vars`) naming the fields `-set` sets.

cotab/sotsb `cmd/fixture/main.go`:

    sv, err := os.ReadFile("../SAVE/SAVGAMA.DAT") // then fixture.OpenRawSave(sv, layout)
    fixture.Run(fixture.Game{Root: ..., Open: ..., ECL: <dax.NewLoader(root).Block("ECLn", b)>, Member: <read ../SAVE/name+ext>,
        Strong: fixture.MakeStrong, ParseDecision: ...}, os.Args[1:])

Test: `go test ./...` (temp dirs only; the fixture has no randomness, so no seed).

## Migrate
- `OpenRawSave(sv)` became `OpenRawSave(sv, fixture.RawLayout{...})`. cotab and sotsb pass their layout (the former constants):
  `Size 0x3351, Disk 0, Area 1, Game 0x801, ECL 0x1401, ECLSize 0x1E00, Pos 0x3201, Count 0x3208, Names 0x3209, NameSize 0x29,
  GameSymSet 0x624, AreaBlock 0x1E4, AreaGeo 0x18A, AreaWallBlock 0x3F4, AreaWallSlot 0x3FA`.
