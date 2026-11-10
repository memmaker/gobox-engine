#!/bin/bash
# Commit gate of the POR migration: builds, vets and tests what is staged, from a clean
# export of each repo, never the working tree. Run with the changes staged; commit if it passes.
set -euo pipefail
projects="$HOME/Projects"
out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT
export GOWORK=off

for repo in por-go rules1e gobox-engine; do
  mkdir "$out/$repo"
  tree="$(git -C "$projects/$repo" write-tree)"
  git -C "$projects/$repo" archive "$tree" | tar -x -C "$out/$repo"
done
# Unversioned here: third-party checkouts, the -re reconstructions (tests read ../../<game>-re), and the game data.
for dep in gogpu-wgpu goffi goplay gameimport; do ln -s "$projects/$dep" "$out/$dep"; done
for re in "$projects"/*-re; do ln -s "$re" "$out/$(basename "$re")"; done
# link_data repo: the repo's unversioned data/ entries into its export.
link_data() {
  local d
  [ -d "$projects/$1/data" ] || return 0
  mkdir -p "$out/$1/data"
  for d in "$projects/$1"/data/*; do
    [ -e "$out/$1/data/$(basename "$d")" ] || ln -s "$d" "$out/$1/data/"
  done
}
link_data por-go

check() { # dir, test packages and flags
  local dir="$1"; shift
  echo "== $dir"
  local unformatted
  unformatted="$(cd "$out/$dir" && gofmt -l .)"
  [ -z "$unformatted" ] || { echo "gofmt needed:"; echo "$unformatted"; return 1; }
  (cd "$out/$dir" && go build ./... && go vet ./... && go test "$@")
}
# Known slow tests, kept out by name; everything else in a touched package runs.
slow='^(TestNoSilentDo|TestRefusedCastKeepsTurn)$'

# touched repo: the packages (as module dir and ./relative path) whose .go files the staged change touches.
# A behaviour commit must not leave a test stale in a package the gate would not run (por-go 6f9b20a).
touched() {
  local repo="$1" f d mod
  git -C "$projects/$repo" diff --cached --name-only HEAD -- '*.go' | while read -r f; do
    d="$out/$repo/$(dirname "$f")"
    [ -d "$d" ] || continue
    mod="$d"
    while [ ! -e "$mod/go.mod" ] && [ "$mod" != "$out/$repo" ]; do mod="$(dirname "$mod")"; done
    if [ "$d" = "$mod" ]; then echo "$mod ."; else echo "$mod ./${d#"$mod"/}"; fi
  done | sort -u
}
test_touched() {
  local repo="$1" mod pkg
  touched "$repo" | while read -r mod pkg; do
    echo "== $repo: $pkg (touched)"
    (cd "$mod" && go test -skip "$slow" "$pkg") || exit 1
  done
}

check rules1e ./...
check gobox-engine/core ./...
check por-go ./rules ./rules4e -run 'Monster|Thac0|Save|Train|Level|Turn|Charcalc|Recompute|AgainstPascal'
for repo in por-go rules1e gobox-engine; do test_touched "$repo"; done

# Other games build against gobox-engine outside core, from what is staged in them (their HEAD when nothing is).
if [ -n "${GATE_ALL:-}" ] || git -C "$projects/gobox-engine" diff --cached --name-only | grep -qv '^core/\|\.md$\|^gate.sh$'; then
  for mod in "$projects"/*/go.mod; do
    repo="$(basename "$(dirname "$mod")")"
    [ -e "$out/$repo" ] && continue
    grep -q 'gobox-engine/' "$mod" || continue
    mkdir "$out/$repo"
    git -C "$projects/$repo" archive "$(git -C "$projects/$repo" write-tree)" | tar -x -C "$out/$repo"
    link_data "$repo"
    echo "== $repo"
    (cd "$out/$repo" && go build ./... && go vet ./...)
    # A layout name the fold or the parser's macros depend on (Shop, Temple, Portrait) can be overwritten by a
    # layout change (cotab 8ee0677): each game's script tests run with every gobox change.
    if [ -d "$out/$repo/rules/scripts/layout" ]; then
      (cd "$out/$repo" && go test ./rules -run 'Layout|Fold|Gate')
    fi
    test_touched "$repo"
  done
fi
echo "gate passed"
