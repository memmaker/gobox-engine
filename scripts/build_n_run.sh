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
# The user plays: the window comes to the front and takes the keys (without a TTY,
# say from an editor task, gogpu would otherwise open it behind, as for agents).
# stderr (a Go panic's trace) goes to crash_logs/<time>.log, kept only when the game fails
mkdir -p crash_logs
log=crash_logs/$(date +%Y-%m-%d_%H-%M-%S)_$$.log
status=0
GOBOX_FOREGROUND=${GOBOX_FOREGROUND-1} GOBOX_ENGINE=$(cd "$(dirname "$0")/.." && pwd) ./run.sh "$@" 2>"$log" || status=$?
if [ $status -eq 0 ]; then rm -f "$log"; else echo "exited with status $status, stderr in $(pwd)/$log" >&2; fi
exit $status
