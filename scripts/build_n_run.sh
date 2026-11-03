#!/bin/sh
# Build a Go port and play it with its original game files.
#   build_n_run.sh GAMEDIR [flags...]
# GAMEDIR is the game repo. The script knows nothing of any game: it runs GAMEDIR/run.sh with the
# flags, from GAMEDIR, with GOBOX_ENGINE set to the engine's directory. The hook builds, imports
# data, and starts the game (examples in scripts/README.md).
set -e
[ -d "$1" ] || { echo "usage: $0 GAMEDIR [flags...]" >&2; exit 1; }
cd "$1"; shift
[ -x ./run.sh ] || { echo "no $(pwd)/run.sh: the game supplies its build-and-run hook (see gobox-engine/scripts/README.md)" >&2; exit 1; }
GOBOX_ENGINE=$(cd "$(dirname "$0")/.." && pwd) exec ./run.sh "$@"
