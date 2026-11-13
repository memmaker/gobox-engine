# scripts

`build_n_run.sh GAMEDIR [flags...]` runs `GAMEDIR/run.sh [flags...]` (cwd GAMEDIR, `GOBOX_ENGINE` set).
The hook is the game's: build, import data, launch. The shared script has no game-specific steps.

`check_replaces.sh GAMEDIR...` fails when a game's go.mod lacks the replace lines gobox/gpu needs (gpu/API.md).

Each game repo keeps a one-line `build_n_run.sh`:

    #!/bin/sh
    exec "$(dirname "$0")/../gobox-engine/scripts/build_n_run.sh" "$(cd "$(dirname "$0")" && pwd)" "$@"

## Migrate

Each game adds an executable `run.sh`. Examples:

Pool of Radiance (originals in `data/por/sources/dos`, Amiga conversion on first run, then the import and the one binary):

    #!/bin/sh
    set -e
    go build -o gobox ./cmd/gobox
    [ -d data/por/sources/amiga ] || ./gobox import -amiga por "${ADF_DIR:-$HOME/Downloads}"
    ./gobox import por data/por/sources/dos
    exec ./gobox "$@" por

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
