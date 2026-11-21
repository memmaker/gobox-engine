# br-default: Buck Rogers is a data-folder game on the shared engine

## What changed for a user
- `gobox import br1 <originals>` and `gobox import br2 <originals>` write the folder; `gobox br1` / `gobox br2` play it
  (the launcher lists it). Same as the Pool games: the `hosts` table picks the host by the manifest's ruleset.
- Folders imported before this change are stale (BR import version 23): import again.
- The old BR test starts (`-teststart`, `-block`, `-credits`, `-hp`, `-space`, `-rest`, `-fightn`, `-shop`) are gone;
  `-explore`, `-battle`, `-dev` and the engine's dev scenes are the starts.

## What was removed
- Flag `-shared` (run and import), `Game.ImportShared`, `importer.RunWith`, the old `ImportBR` (one `ImportBR` now:
  the former `ImportBREngine`), `start.Shared`, `cmd/gobox/br.go` (the `players` table, `playOther`, `brFlags`).
- Import outputs only the old engines read: `data/MON0.json`, `data/ITEM0.json`, `data/ITEMS.bin/.txt`,
  `data/START.bin/.txt`, `data/OVR.bin/.txt`, `data/8X8D1.bin/.txt`, `data/SHIPS.bin/.txt`, the originals' `saves/CHRDATA*`
  copies (written, then deleted by the roster step) and the partial script layout (written, then replaced).
  Kept: GEO1..6, `tables.json`, `characters/`, `pictures.json`, `art/`, `journal/`, `gamedata.json`, `gametables.json`,
  `sound.json` (silent: the BR sound driver is a later track), the full layout, `saves/` roster.
- `go.mod`: `buckrogers`, `buckrogers2`, `ecl-br` requires and replaces. `ui.Game.CreateCharacter/Roll/Rename`
  (only the old repos' front tests called them). Sibling-repo paths in guards, `foldwt.sh`, `gate.sh`, `scripts/`.
- Goldens stay under `rulesxxvc/testdata/oldengine/` (files; their README says how they were made).

## Absent from a BR folder against `data/por`, and why
- `COVER.PNG`: nothing in this repo reads it; its writer decodes EGA, BR's TITLE is VGA.
- `docs/`, `extras/`, `sources/`: kept by the import (the player's and hand-written), not written by it.
- `sound/`, `soundtrack/`, `speech/`: the Pool's sound sources; BR has none yet.
- BR2 `journal/`: no source (Open item in REFACTOR.md).

## The old repositories
`buckrogers-go`, `buckrogers2-go` and `buckrogers-core` are deleted (2026-11-21). Their full histories, including
`buckrogers-go` branch `golden/one-rng` where the goldens were made, are git bundles in `~/Projects/re-archive/buckrogers/`.
The BR1 and BR2 remaster renders moved to `data/br1/extras/remaster` and `data/br2/extras/remaster`.
`rulesxxvc-history.bundle` is in `~/Projects/re-archive/`.
