#!/bin/sh
# Build a Go port and play it with its original game files.
#   build_n_run.sh GAMEDIR [flags...]       e.g. ../cotab-go -scale 4
#   build_n_run.sh GAMEDIR snap OUT save keys   headless screenshots (cotab, sotsb; see the game's cmd/snap)
# GAMEDIR is the game repo (por-go, cotab-go, sotsb-go); the binary is named after it (por, cotab, sotsb).
# por keeps its files in GAMEDIR/data (dos, amiga, assets, save; the Amiga import runs on first use from
# $ADF_DIR, default ~/Downloads, and the sound pack from $DASP_DIR); the others read GAMEDIR itself (START.EXE, *.DAX, ...).
set -e
[ -d "$1" ] || { echo "usage: $0 GAMEDIR [flags...]" >&2; exit 1; }
cd "$1"; shift
root=$(pwd)
game=$(basename "$root"); game=${game%-go}
if [ "$game" = por ]; then
	[ -f data/dos/TITLE.DAX ] || { echo "no game data: copy your Pool of Radiance files (*.DAX, ITEMS, POOL.CFG) into $root/data/dos" >&2; exit 1; }
	go build -o por ./cmd/por
	[ -d data/amiga ] || ./por -data data -import-disk2 "${ADF_DIR:-$HOME/Downloads}"
	dasp="${DASP_DIR:-$HOME/Downloads/Dubtrain Angband Sound Pack v3.1.0}"
	[ -d data/assets/dasp ] || [ ! -f "$dasp/sound.cfg" ] || { mkdir -p data/assets/dasp && cp "$dasp"/* data/assets/dasp/; }
	exec ./por "$@" # por reads ./data
fi
[ -f START.EXE ] || { echo "no game data: copy the original files (START.EXE, *.DAX, ITEMS, ...) into $root" >&2; exit 1; }
mkdir -p bin
if [ "$1" = snap ]; then
	shift
	mkdir -p "$1"
	go build -o bin/snap ./cmd/snap
	exec bin/snap "$root" "$@"
fi
go build -o "bin/$game" "./cmd/$game"
exec "bin/$game" -data "$root" "$@"
