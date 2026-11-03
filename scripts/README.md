# scripts

`build_n_run.sh GAMEDIR [flags...]` builds and plays a game (por-go, cotab-go, sotsb-go).
Covers por's layout (files in `GAMEDIR/data`, Amiga import and sound pack on first run) and
cotab/sotsb's (files in `GAMEDIR`, plus `snap OUT save keys` for headless screenshots).

Each game repo keeps a one-line `build_n_run.sh`:

    #!/bin/sh
    exec "$(dirname "$0")/../gobox-engine/scripts/build_n_run.sh" "$(cd "$(dirname "$0")" && pwd)" "$@"
