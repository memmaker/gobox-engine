# tools/remaster

Generic wall/floor texture pipeline (cotab/sotsb/por). Calls the OpenAI image API through the game's
`remaster-walls.py` (`cfg.rw`, default `tools/remaster-walls.py` in the game dir); only `make_*` calls it.
Game data lives in `GAME/remaster/`: `remaster.json`, `areas_data.py` (the AREAS table, run by `areas.py`
with `area()/src()/origin()` in scope), `maps.txt` (optional: per-map walls + `index.txt`), `viewdump/`, `raw/`.

`remaster.json` (paths relative to GAME): `out` textures dir; `dump` command writing the wall pictures
(`{data}` `{out}` filled in); `data` original files dir; `walldump` command for maps.txt (optional);
`reuse_dir` (another game's textures, for `reuse`); `trim` [r,g,b]; `alias` {map wall: texture};
`context_replace` [[old,new]] edits of the prompt CONTEXT; `floor` {ref,prompt,raw,out}.

    python3 tools/remaster/make_textures.py ../cotab-go [AREA] [finish | VALUE...]   # finish = no API call
    python3 tools/remaster/make_floor.py ../cotab-go
    por-go:   same with ../por-go (remaster.json: dump ["go","run","./cmd/viewdump","-transparent",".","{out}"], trim [0,170,170], no maps.txt)
    sotsb-go: same with ../sotsb-go

Moving a game over: split its `areas.py` into `areas_data.py` (the AREAS entries only; helpers live here) and
write `remaster.json`. `groups.json` is not read by these scripts.
