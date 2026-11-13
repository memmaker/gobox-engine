# tools/remaster

Generic wall/floor texture pipeline (cotab/sotsb/por). Calls the OpenAI image API through the game's
`remaster-walls.py` (`cfg.rw`, default `tools/remaster-walls.py` in the game dir); only `make_*` calls it.
Game data lives in `GAME/remaster/`: `remaster.json`, `areas_data.py` (the AREAS table, run by `areas.py`
with `area()/src()/origin()` in scope), `maps.txt` (optional: per-map walls + `index.txt`), `viewdump/`, `raw/`.

`remaster.json` (paths relative to GAME): `out` textures dir; `dump` command writing the wall pictures
(`{data}` `{out}` filled in); `data` original files dir; `walldump` command for maps.txt (optional);
`reuse_dir` (another game's textures, for `reuse`); `index` the game's map table path (default `out/index.txt`; set it when `out` is the gobox/walls pool); `trim` [r,g,b]; `alias` {map wall: texture};
`context_replace` [[old,new]] edits of the prompt CONTEXT; `floor` {ref,prompt,raw,out} or a list of them, each optionally with `crop` [x0,y0,x1,y1], `sky` [w,h] (an untiled sky_* picture) and `pixel` (the pixel-style copy's path).

    python3 tools/remaster/make_textures.py ../cotab-go [AREA] [finish | VALUE...]   # finish = no API call
    python3 tools/remaster/make_floor.py ../cotab-go [RAW...]   # RAW: only those entries; a raw output already there is reused, no API call

Both refuse to replace an existing, different texture unless given `--force` (`out` may be the shared pool).
    por-go:   same with ../por-go (remaster.json: dump ["go","-C","../gobox-engine","run","./ui/cmd/viewdump","-transparent","{data}","{out}"], trim [0,170,170], no maps.txt)
    sotsb-go: same with ../sotsb-go

Moving a game over: split its `areas.py` into `areas_data.py` (the AREAS entries only; helpers live here) and
write `remaster.json`. `groups.json` is not read by these scripts.

## Migrate

In `areas_data.py`, pass `reused={value: texture}` to `area()` instead of `por=`. The old `por=` key is still read as a fallback.
