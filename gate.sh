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
# Unversioned here: third-party checkouts, and the game data.
for dep in gogpu-wgpu goffi goplay gameimport; do ln -s "$projects/$dep" "$out/$dep"; done
for d in "$projects"/por-go/data/*; do
  [ -e "$out/por-go/data/$(basename "$d")" ] || ln -s "$d" "$out/por-go/data/"
done

check() { # dir, test packages and flags
  local dir="$1"; shift
  echo "== $dir"
  (cd "$out/$dir" && go build ./... && go vet ./... && go test "$@")
}
check rules1e ./...
check gobox-engine/core ./...
check por-go ./rules ./rules4e -run 'Monster|Thac0|Save|Train|Level|Turn|Charcalc|Recompute'

# Other games only see gobox-engine outside core; they build against it from their own commit.
if [ -n "${GATE_ALL:-}" ] || git -C "$projects/gobox-engine" diff --cached --name-only | grep -qv '^core/\|\.md$\|^gate.sh$'; then
  for mod in "$projects"/*/go.mod; do
    repo="$(basename "$(dirname "$mod")")"
    [ -e "$out/$repo" ] && continue
    grep -q 'gobox-engine/' "$mod" || continue
    mkdir "$out/$repo"
    git -C "$projects/$repo" archive HEAD | tar -x -C "$out/$repo"
    echo "== $repo"
    (cd "$out/$repo" && go build ./... && go vet ./...)
  done
fi
echo "gate passed"
