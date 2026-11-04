#!/bin/sh
# Put a game's original files into its data folder from the GOG installer.
#   install_data.sh DEST PATTERN
# DEST is the game's data folder (e.g. data/dos); PATTERN names the installer in
# $GOLDBOX (default ~/Desktop/Games/Goldbox), e.g. 'setup_pool_of_radiance_*.exe'.
# The script knows nothing of any game: the game's run.sh passes both.
set -e
[ $# -eq 2 ] || { echo "usage: $0 DEST PATTERN" >&2; exit 1; }
dest=$1
src=${GOLDBOX:-$HOME/Desktop/Games/Goldbox}
set -- "$src"/$2
[ -f "$1" ] || { echo "no installer $2 in $src (set GOLDBOX)" >&2; exit 1; }
command -v innoextract >/dev/null || { echo "needs innoextract (brew install innoextract)" >&2; exit 1; }
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
innoextract -s -d "$tmp" "$1"
mkdir -p "$dest"
# the game's own files only: GOG's DOSBox, redistributables and installer bits stay out
for f in "$tmp"/*; do
	case $(basename "$f") in DOSBOX | __redist | __support | commonappdata | tmp | app) continue ;; esac
	cp -R "$f" "$dest"/
done
echo "installed $(basename "$1") into $dest"
