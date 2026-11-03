# scripts

`build_n_run.sh GAMEDIR [flags...]` runs `GAMEDIR/run.sh [flags...]` (cwd GAMEDIR, `GOBOX_ENGINE` set).
The hook is the game's: build, import data, launch. The shared script has no game-specific steps.

`check_replaces.sh GAMEDIR...` fails when a game's go.mod lacks the replace lines gobox/gpu needs (gpu/API.md).

Each game repo keeps a one-line `build_n_run.sh`:

    #!/bin/sh
    exec "$(dirname "$0")/../gobox-engine/scripts/build_n_run.sh" "$(cd "$(dirname "$0")" && pwd)" "$@"

## Migrate

Each game adds an executable `run.sh`. Examples:

por-go/run.sh (files in `data/`, Amiga import and sound pack on first run):

    #!/bin/sh
    set -e
    [ -f data/dos/TITLE.DAX ] || { echo "no game data: copy your Pool of Radiance files (*.DAX, ITEMS, POOL.CFG) into $(pwd)/data/dos" >&2; exit 1; }
    go build -o por ./cmd/por
    [ -d data/amiga ] || ./por -data data -import-disk2 "${ADF_DIR:-$HOME/Downloads}"
    dasp="${DASP_DIR:-$HOME/Downloads/Dubtrain Angband Sound Pack v3.1.0}"
    [ -d data/assets/dasp ] || [ ! -f "$dasp/sound.cfg" ] || { mkdir -p data/assets/dasp && cp "$dasp"/* data/assets/dasp/; }
    exec ./por "$@"

cotab-go/run.sh and sotsb-go/run.sh (files in the repo itself; `G` is cotab or sotsb; `snap OUT save keys` makes headless screenshots):

    #!/bin/sh
    set -e
    G=cotab
    root=$(pwd)
    [ -f START.EXE ] || { echo "no game data: copy the original files (START.EXE, *.DAX, ITEMS, ...) into $root" >&2; exit 1; }
    mkdir -p bin
    if [ "$1" = snap ]; then
    	shift; mkdir -p "$1"
    	go build -o bin/snap ./cmd/snap
    	exec bin/snap "$root" "$@"
    fi
    go build -o "bin/$G" "./cmd/$G"
    exec "bin/$G" -data "$root" "$@"
