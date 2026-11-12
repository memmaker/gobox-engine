#!/bin/bash
# Commit gate of the POR migration: builds, vets and tests what is staged, from a clean
# export of each repo, never the working tree. Run with the changes staged; commit if it passes.
set -euo pipefail
projects="${GATE_PROJECTS:-$HOME/Projects}"
out="$(mktemp -d)"
trap 'rm -rf "$out"' EXIT
export GOWORK=off
# A home of its own: caches the tests write (os.UserCacheDir, e.g. the imported scripts) come
# from the staged code, never from another agent's working tree. Go's build caches stay shared.
export GOCACHE="$(go env GOCACHE)" GOMODCACHE="$(go env GOMODCACHE)" GOPATH="$(go env GOPATH)"
export HOME="$out/home"
mkdir "$HOME"

# The engine's repos and every game built on gobox-engine, from what is staged in them (their HEAD when nothing is).
games=()
for mod in "$projects"/*/go.mod; do
  repo="$(basename "$(dirname "$mod")")"
  case "$repo" in por-go|rules1e|rulesxxvc|gobox-engine) continue ;; esac
  grep -q 'gobox-engine/' "$mod" && games+=("$repo")
done
for repo in por-go rules1e rulesxxvc gobox-engine "${games[@]}"; do
  mkdir "$out/$repo"
  tree="$(git -C "$projects/$repo" write-tree)"
  git -C "$projects/$repo" archive "$tree" | tar -x -C "$out/$repo"
done
# Tests that look at the other games (scripts/check-game-imports.sh) see these exports.
export PROJECTS="$out"
# Tests that find a sibling repo through $HOME/Projects (buckrogers bugclass_test, front sequence_test) see them too.
ln -s "$out" "$HOME/Projects"
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
for repo in por-go "${games[@]}"; do link_data "$repo"; done

check() { # dir, test packages and flags
  local dir="$1"; shift
  echo "== $dir"
  local unformatted
  unformatted="$(cd "$out/$dir" && gofmt -l .)"
  [ -z "$unformatted" ] || { echo "gofmt needed:"; echo "$unformatted"; return 1; }
  local t=$SECONDS
  (cd "$out/$dir" && go build ./... && go vet ./... && go test "$@")
  echo "-- $dir $((SECONDS - t))s"
}
# Known slow tests, kept out by name; everything else in a touched package runs.
slow='^(TestNoSilentDo|TestRefusedCastKeepsTurn)$'

# touched repo: the packages (as module dir and ./relative path) whose .go files the staged change touches.
# A behaviour commit must not leave a test stale in a package the gate would not run (por-go 6f9b20a).
touched() {
  local repo="$1" f d mod
  # Staged changes, and with GATE_BASE (foldwt land) every commit since that base too.
  { git -C "$projects/$repo" diff --cached --name-only HEAD -- '*.go'
    [ -z "${GATE_BASE:-}" ] || git -C "$projects/$repo" diff --name-only "$GATE_BASE"...HEAD -- '*.go'
  } | sort -u | while read -r f; do
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
    local t=$SECONDS
    (cd "$mod" && go test -skip "$slow" "$pkg") || exit 1
    echo "-- $repo: $pkg $((SECONDS - t))s"
  done
}

check rules1e ./...
check rulesxxvc ./...
check gobox-engine/core ./...
fixed='Monster|Thac0|Save|Train|Level|Turn|Charcalc|Recompute|AgainstPascal'
# POR's engine (gobox-engine/pool; por-go/rules before fold E2): the fixed set, and the slow tests the touched run skips.
if [ -d "$out/gobox-engine/pool" ]; then
  # pool's tests read POR's imported data (scripts, data/assets/data) and fail without it: the import
  # (por-go cmd/importstale, a no-op without the game files) runs first, here, not in the tests.
  (cd "$out/por-go" && go run ./cmd/importstale)
  check gobox-engine/pool . -run "$fixed"
  (cd "$out/gobox-engine/pool" && go test . -run "$slow")
  check por-go ./rules4e -run "$fixed"
elif touched por-go | grep -q ' \./rules$'; then
  check por-go ./rules4e -run "$fixed"
  (cd "$out/por-go" && go test ./rules -run "$slow")
else
  check por-go ./rules ./rules4e -run "$fixed"
fi
for repo in por-go rules1e gobox-engine; do test_touched "$repo"; done

# Other games build against gobox-engine outside core, from what is staged in them (their HEAD when nothing is).
if [ -n "${GATE_ALL:-}" ] || git -C "$projects/gobox-engine" diff --cached --name-only | grep -qv '^core/\|\.md$\|^gate.sh$'; then
  for repo in "${games[@]}"; do
    echo "== $repo"
    t=$SECONDS
    (cd "$out/$repo" && go build ./... && go vet ./...)
    echo "-- $repo build+vet $((SECONDS - t))s"
    # A layout name the fold or the parser's macros depend on (Shop, Temple, Portrait) can be overwritten by a
    # layout change (cotab 8ee0677): each game's script tests run with every gobox change.
    if [ -d "$out/$repo/rules/scripts/layout" ]; then
      (cd "$out/$repo" && go test ./rules -run 'Layout|Fold|Gate')
    fi
    test_touched "$repo"
  done
fi
echo "gate passed in ${SECONDS}s"
